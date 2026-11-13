// Package geography maintains primary-owned shared reference data and an
// Observatory replica. It never deletes geography or rewrites fact-table IDs.
package geography

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type record map[string]any
type snapshot map[string][]record
type level struct {
	table, kind string
	parents     map[string]string
}

var levels = []level{
	{"regions", "region", nil},
	{"districts", "district", map[string]string{"region_id": "regions"}},
	{"sub_counties", "sub-county", map[string]string{"district_id": "districts"}},
	{"facilities", "facility", map[string]string{"region_id": "regions", "district_id": "districts", "sub_county_id": "sub_counties"}},
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func read(ctx context.Context, db queryer, prefix string) (snapshot, error) {
	result := snapshot{}
	for _, l := range levels {
		var raw []byte
		// prefix and table are internal constants, never request inputs.
		if err := db.QueryRowContext(ctx, `SELECT COALESCE(jsonb_agg(to_jsonb(g) ORDER BY id), '[]'::jsonb) FROM `+prefix+l.table+` g`).Scan(&raw); err != nil {
			return nil, err
		}
		var rows []record
		if err := json.Unmarshal(raw, &rows); err != nil {
			return nil, err
		}
		result[l.table] = rows
	}
	return result, nil
}

func value(r record, key string) string { s, _ := r[key].(string); return s }
func normal(s string) string            { return strings.ToLower(strings.TrimSpace(s)) }
func copyRecord(r record) record {
	c := record{}
	for k, v := range r {
		c[k] = v
	}
	return c
}
func naturalKey(table string, r record) string {
	key := normal(value(r, "name"))
	if table == "sub_counties" || table == "facilities" {
		key += "|" + value(r, "district_id")
	}
	return key
}

type recordIndex struct {
	ids             map[string]record
	external, names map[string][]record
}

func index(table string, rows []record) recordIndex {
	i := recordIndex{map[string]record{}, map[string][]record{}, map[string][]record{}}
	for _, r := range rows {
		i.ids[value(r, "id")] = r
		if external := value(r, "external_id"); external != "" {
			i.external[external] = append(i.external[external], r)
		}
		key := naturalKey(table, r)
		i.names[key] = append(i.names[key], r)
	}
	return i
}

// match refuses ambiguous identities instead of merging unrelated areas. Parent
// IDs in candidates have already been translated into the destination registry.
func match(table string, candidate record, rows recordIndex, preferred string) (record, error) {
	byID := rows.ids[preferred]
	external := rows.external[value(candidate, "external_id")]
	names := rows.names[naturalKey(table, candidate)]
	if len(external) > 1 || len(names) > 1 {
		return nil, errors.New("ambiguous geography identity")
	}
	var byExternal, byName record
	if len(external) == 1 {
		byExternal = external[0]
	}
	if len(names) == 1 {
		byName = names[0]
	}
	if byID != nil {
		if (byExternal != nil && value(byExternal, "id") != preferred) || (byName != nil && value(byName, "id") != preferred) {
			return nil, errors.New("conflicting geography identities")
		}
		return byID, nil
	}
	if byExternal != nil {
		if byName != nil && value(byExternal, "id") != value(byName, "id") {
			return nil, errors.New("conflicting facility identity")
		}
		return byExternal, nil
	}
	return byName, nil
}

func remap(r record, l level, ids map[string]map[string]string) (record, error) {
	c := copyRecord(r)
	for column, parent := range l.parents {
		if original := value(r, column); original != "" {
			mapped := ids[parent][original]
			if mapped == "" {
				return nil, errors.New("geography parent is missing from replica")
			}
			c[column] = mapped
		}
	}
	return c, nil
}

// adopt runs once for a replica. Existing Observatory metadata wins this initial
// transfer; primary UUIDs and health-context assignments remain stable. Subsequent
// synchronization is exclusively primary -> Observatory.
func adopt(primary, remote snapshot) (snapshot, error) {
	for _, data := range []snapshot{primary, remote} {
		for _, l := range levels {
			lookup := index(l.table, data[l.table])
			for _, matches := range lookup.names {
				if len(matches) > 1 {
					return nil, errors.New("ambiguous geography natural identity during adoption")
				}
			}
			for _, matches := range lookup.external {
				if len(matches) > 1 {
					return nil, errors.New("ambiguous facility external identity during adoption")
				}
			}
		}
	}
	ids := map[string]map[string]string{}
	for _, l := range levels {
		ids[l.table] = map[string]string{}
		lookup := index(l.table, primary[l.table])
		positions := map[string]int{}
		for i, row := range primary[l.table] {
			positions[value(row, "id")] = i
		}
		for _, r := range remote[l.table] {
			c, err := remap(r, l, ids)
			if err != nil {
				return nil, err
			}
			found, err := match(l.table, c, lookup, value(r, "id"))
			if err != nil {
				return nil, err
			}
			if found != nil {
				original := value(found, "id")
				c["id"] = original
				primary[l.table][positions[original]] = c
			} else {
				primary[l.table] = append(primary[l.table], c)
			}
			ids[l.table][value(r, "id")] = value(c, "id")
		}
	}
	return primary, nil
}

func write(ctx context.Context, tx *sql.Tx, prefix string, s snapshot) error {
	for _, l := range levels {
		raw, err := json.Marshal(s[l.table])
		if err != nil {
			return err
		}
		// Explicit mutable columns exclude IDs and creation timestamps. Avoid firing
		// context triggers on unchanged rows during periodic synchronization.
		columns := []string{"name", "code"}
		switch l.table {
		case "districts":
			columns = append(columns, "region_id")
		case "sub_counties":
			columns = append(columns, "district_id")
		case "facilities":
			columns = []string{"name", "external_id", "district_id", "sub_county_id", "region_id", "facility_level", "facility_type", "dhis2_org_unit_id", "latitude", "longitude"}
		}
		setters, oldValues, newValues := []string{}, []string{}, []string{}
		for _, column := range columns {
			setters = append(setters, column+"=EXCLUDED."+column)
			oldValues = append(oldValues, "existing."+column)
			newValues = append(newValues, "EXCLUDED."+column)
		}
		query := `INSERT INTO ` + prefix + l.table + ` AS existing SELECT * FROM jsonb_populate_recordset(NULL::` + prefix + l.table + `,$1::jsonb)
   ON CONFLICT(id) DO UPDATE SET ` + strings.Join(setters, ",") + `,updated_at=NOW()
   WHERE ROW(` + strings.Join(oldValues, ",") + `) IS DISTINCT FROM ROW(` + strings.Join(newValues, ",") + `)`
		if _, err := tx.ExecContext(ctx, query, string(raw)); err != nil {
			return fmt.Errorf("write %s registry: %w", l.table, err)
		}
	}
	return nil
}

// Replicator serializes background refreshes and immediate location-write replication.
type Replicator struct {
	mu              sync.Mutex
	primary, remote *sql.DB
	lastSync        time.Time
}

func NewReplicator(primary, remote *sql.DB) *Replicator {
	return &Replicator{primary: primary, remote: remote}
}
func (r *Replicator) Refresh(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if time.Since(r.lastSync) < time.Minute {
		return nil
	}
	return r.synchronize(ctx)
}

func (r *Replicator) Sync(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.synchronize(ctx)
}

func (r *Replicator) synchronize(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// Lock primary first on every attempt; a second backend cannot interleave the
	// adoption, destination commit and alias publication of the same registry.
	primary, err := r.primary.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return err
	}
	defer primary.Rollback()
	if _, err = primary.ExecContext(ctx, `SELECT pg_advisory_xact_lock(716039124)`); err != nil {
		return err
	}
	var registryID string
	if err = primary.QueryRowContext(ctx, `SELECT id FROM geography.registry WHERE singleton`).Scan(&registryID); err != nil {
		return err
	}
	remote, err := r.remote.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return err
	}
	defer remote.Rollback()
	if _, err = remote.ExecContext(ctx, `INSERT INTO geography_replica_origin(singleton,registry_id) VALUES(TRUE,$1) ON CONFLICT DO NOTHING`, registryID); err != nil {
		return err
	}
	var owner string
	var adopted bool
	var replicaID string
	if err = remote.QueryRowContext(ctx, `SELECT id FROM geography_replica_identity WHERE singleton`).Scan(&replicaID); err != nil {
		return err
	}
	if err = remote.QueryRowContext(ctx, `SELECT registry_id,adopted FROM geography_replica_origin WHERE singleton FOR UPDATE`).Scan(&owner, &adopted); err != nil {
		return err
	}
	if owner != registryID {
		return errors.New("Observatory replica belongs to a different primary registry")
	}
	canonical, err := read(ctx, primary, "geography.")
	if err != nil {
		return err
	}
	existing, err := read(ctx, remote, "")
	if err != nil {
		return err
	}
	if err = primary.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM geography.adoptions WHERE replica_id=$1)`, replicaID).Scan(&adopted); err != nil {
		return err
	}
	if !adopted {
		canonical, err = adopt(canonical, existing)
		if err != nil {
			return err
		}
		if err = write(ctx, primary, "geography.", canonical); err != nil {
			return err
		}
		// Commit adoption before publishing a destination marker. If either process
		// stops here, repeating the missing marker transfer is idempotent.
		if _, err = primary.ExecContext(ctx, `INSERT INTO geography.adoptions(replica_id) VALUES($1) ON CONFLICT DO NOTHING`, replicaID); err != nil {
			return err
		}
		if err = primary.Commit(); err != nil {
			return err
		}
		primary, err = r.primary.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		if err != nil {
			return err
		}
		defer primary.Rollback()
		// Do not reacquire the advisory lock while holding the remote row lock:
		// another process can hold primary and wait for that remote transaction.
		canonical, err = read(ctx, primary, "geography.")
		if err != nil {
			return err
		}
	}
	mappings := map[string]map[string]string{}
	output := snapshot{}
	for _, l := range levels {
		mappings[l.table] = map[string]string{}
		var linksJSON []byte
		if err = remote.QueryRowContext(ctx, `SELECT COALESCE(jsonb_object_agg(source_id::text,local_id::text),'{}'::jsonb) FROM geography_replica_links WHERE kind=$1`, l.kind).Scan(&linksJSON); err != nil {
			return err
		}
		links := map[string]string{}
		if err = json.Unmarshal(linksJSON, &links); err != nil {
			return err
		}
		lookup := index(l.table, existing[l.table])
		for _, row := range canonical[l.table] {
			c, e := remap(row, l, mappings)
			if e != nil {
				return e
			}
			source := value(row, "id")
			preferred := links[source]
			if preferred == "" {
				preferred = source
			}
			found, e := match(l.table, c, lookup, preferred)
			if e != nil {
				return e
			}
			local := source
			if found != nil {
				local = value(found, "id")
				c["created_at"] = found["created_at"]
			}
			c["id"] = local
			mappings[l.table][source] = local
			output[l.table] = append(output[l.table], c)
		}
	}
	if err = write(ctx, remote, "", output); err != nil {
		return err
	}
	for _, l := range levels {
		raw, _ := json.Marshal(mappings[l.table])
		if _, err = remote.ExecContext(ctx, `INSERT INTO geography_replica_links(kind,source_id,local_id) SELECT $1,key::uuid,value::uuid FROM jsonb_each_text($2::jsonb) ON CONFLICT(kind,source_id) DO UPDATE SET local_id=EXCLUDED.local_id`, l.kind, string(raw)); err != nil {
			return err
		}
	}
	if _, err = remote.ExecContext(ctx, `UPDATE geography_replica_origin SET adopted=TRUE,synchronized_at=NOW() WHERE singleton`); err != nil {
		return err
	}
	if err = remote.Commit(); err != nil {
		return err
	}
	for _, l := range levels {
		raw, _ := json.Marshal(mappings[l.table])
		// Keep the public surveillance namespace bound to destination IDs. Canonical
		// geography aliases use a separate namespace, so differing UUIDs are safe.
		if _, err = primary.ExecContext(ctx, `INSERT INTO health_context_aliases(context_node_id,namespace,external_id)
   SELECT a.context_node_id,$1,m.value FROM jsonb_each_text($2::jsonb) m
   JOIN health_context_aliases a ON a.namespace=$3 AND a.external_id=m.key
   ON CONFLICT(context_node_id,namespace) DO UPDATE SET external_id=EXCLUDED.external_id`, "surveillance-"+l.kind, string(raw), "geography-"+l.kind); err != nil {
			return err
		}
	}
	if err = primary.Commit(); err != nil {
		return err
	}
	r.lastSync = time.Now()
	return nil
}
