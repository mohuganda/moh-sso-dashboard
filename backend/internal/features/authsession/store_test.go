package authsession

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// fakeRedis is an in-memory stand-in for the narrow redis interface the
// store uses. TTLs are recorded but not enforced (expiry behavior belongs
// to Redis itself).
type fakeRedis struct {
	values map[string]string
	ttls   map[string]time.Duration
}

func newFakeRedis() *fakeRedis {
	return &fakeRedis{
		values: map[string]string{},
		ttls:   map[string]time.Duration{},
	}
}

func (f *fakeRedis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	switch v := value.(type) {
	case []byte:
		f.values[key] = string(v)
	case string:
		f.values[key] = v
	}
	f.ttls[key] = expiration
	return redis.NewStatusResult("OK", nil)
}

func (f *fakeRedis) Get(ctx context.Context, key string) *redis.StringCmd {
	v, ok := f.values[key]
	if !ok {
		return redis.NewStringResult("", redis.Nil)
	}
	return redis.NewStringResult(v, nil)
}

func (f *fakeRedis) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	var n int64
	for _, k := range keys {
		if _, ok := f.values[k]; ok {
			delete(f.values, k)
			delete(f.ttls, k)
			n++
		}
	}
	return redis.NewIntResult(n, nil)
}

func TestStoreRoundTrip(t *testing.T) {
	ctx := context.Background()
	fake := newFakeRedis()
	store := newStoreWithClient(fake)

	data := Data{
		AccessToken:  "access-jwt",
		RefreshToken: "refresh-jwt",
		IDToken:      "id-jwt",
	}

	id, err := store.Create(ctx, data, 30*time.Minute)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if len(id) != 64 {
		t.Errorf("session id should be 64 hex chars, got %d", len(id))
	}
	if ttl := fake.ttls[keyPrefix+id]; ttl != 30*time.Minute {
		t.Errorf("ttl = %s, want 30m", ttl)
	}

	got, err := store.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if *got != data {
		t.Errorf("Get = %+v, want %+v", got, data)
	}
}

func TestStoreCreateUniqueIDs(t *testing.T) {
	ctx := context.Background()
	store := newStoreWithClient(newFakeRedis())

	a, _ := store.Create(ctx, Data{AccessToken: "x"}, time.Minute)
	b, _ := store.Create(ctx, Data{AccessToken: "x"}, time.Minute)
	if a == b {
		t.Error("two sessions got the same id")
	}
}

func TestStoreCreateRejectsZeroTTL(t *testing.T) {
	store := newStoreWithClient(newFakeRedis())
	if _, err := store.Create(context.Background(), Data{}, 0); err == nil {
		t.Error("expected error for zero ttl")
	}
}

func TestStoreGetMissing(t *testing.T) {
	store := newStoreWithClient(newFakeRedis())

	if _, err := store.Get(context.Background(), "nope"); err != ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
	if _, err := store.Get(context.Background(), ""); err != ErrNotFound {
		t.Errorf("empty id err = %v, want ErrNotFound", err)
	}
}

func TestStoreUpdateAndDelete(t *testing.T) {
	ctx := context.Background()
	fake := newFakeRedis()
	store := newStoreWithClient(fake)

	id, err := store.Create(ctx, Data{AccessToken: "old"}, time.Minute)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := store.Update(ctx, id, Data{AccessToken: "new"}, 2*time.Minute); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := store.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.AccessToken != "new" {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, "new")
	}
	if ttl := fake.ttls[keyPrefix+id]; ttl != 2*time.Minute {
		t.Errorf("ttl after update = %s, want 2m", ttl)
	}

	if err := store.Delete(ctx, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := store.Get(ctx, id); err != ErrNotFound {
		t.Errorf("after delete err = %v, want ErrNotFound", err)
	}

	// Deleting again (or empty id) is a no-op.
	if err := store.Delete(ctx, id); err != nil {
		t.Errorf("second Delete: %v", err)
	}
	if err := store.Delete(ctx, ""); err != nil {
		t.Errorf("empty-id Delete: %v", err)
	}
}
