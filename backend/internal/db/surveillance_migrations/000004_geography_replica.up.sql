-- Records identity mappings without changing existing geography IDs or fact FKs.
CREATE TABLE IF NOT EXISTS geography_replica_origin (
 singleton BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (singleton),
 registry_id UUID NOT NULL,
 adopted BOOLEAN NOT NULL DEFAULT FALSE,
 synchronized_at TIMESTAMPTZ
);
CREATE TABLE IF NOT EXISTS geography_replica_links (
 kind TEXT NOT NULL CHECK (kind IN ('region','district','sub-county','facility')),
 source_id UUID NOT NULL,
 local_id UUID NOT NULL,
 PRIMARY KEY(kind, source_id),
 UNIQUE(kind, local_id)
);
CREATE TABLE IF NOT EXISTS geography_replica_identity (
 singleton BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (singleton),
 id UUID NOT NULL DEFAULT gen_random_uuid()
);
INSERT INTO geography_replica_identity(singleton) VALUES(TRUE) ON CONFLICT DO NOTHING;
