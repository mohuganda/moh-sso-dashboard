package bootstrap

import (
	"context"
	"database/sql"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/moh-sso-dashboard/internal/cache"
	"github.com/moh-sso-dashboard/internal/config"
	kcClientPkg "github.com/moh-sso-dashboard/internal/keycloak"
	db "github.com/moh-sso-dashboard/internal/migrate"
	"github.com/moh-sso-dashboard/internal/ratelimit"
)

type databases struct {
	Primary *sql.DB
	Remote  *sql.DB
	DWH     *sql.DB
}

func initDatabases(ctx context.Context, cfg *config.Config) (databases, error) {
	dbConfig := func(dsn string) db.DBConfig {
		return db.DBConfig{
			Driver:          cfg.DBDriver,
			DSN:             dsn,
			MaxOpenConns:    25,
			MaxIdleConns:    10,
			ConnMaxLifetime: 30 * time.Minute,
			WaitTimeout:     30 * time.Second,
		}
	}

	primaryDB, err := db.InitDB(ctx, dbConfig(cfg.DbSource()))
	if err != nil {
		return databases{}, err
	}

	remoteDB, err := db.InitDB(ctx, dbConfig(cfg.RemoteDbSource()))
	if err != nil {
		primaryDB.Close()
		return databases{}, err
	}

	dwhDB, err := db.InitDB(ctx, dbConfig(cfg.DwhDbSource()))
	if err != nil {
		primaryDB.Close()
		remoteDB.Close()
		return databases{}, err
	}

	if err := db.MigrateDB(primaryDB, "file://internal/db/migrations"); err != nil {
		primaryDB.Close()
		remoteDB.Close()
		dwhDB.Close()
		return databases{}, err
	}

	return databases{
		Primary: primaryDB,
		Remote:  remoteDB,
		DWH:     dwhDB,
	}, nil
}

func (dbs databases) Close() {
	if dbs.Primary != nil {
		dbs.Primary.Close()
	}
	if dbs.Remote != nil {
		dbs.Remote.Close()
	}
	if dbs.DWH != nil {
		dbs.DWH.Close()
	}
}

type cacheRuntime struct {
	Redis       *redis.Client
	Cache       *cache.RedisCache
	RateLimiter *ratelimit.Limiter
}

func initCache(ctx context.Context, cfg *config.Config) cacheRuntime {
	rdb := cache.NewRedisClient(cache.RedisConfig{
		Host:         cfg.RedisHost,
		Port:         cfg.RedisPort,
		Password:     cfg.RedisPassword,
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	cache.MustPing(ctx, rdb)

	return cacheRuntime{
		Redis:       rdb,
		Cache:       cache.NewRedisCache(rdb),
		RateLimiter: ratelimit.New(rdb),
	}
}

type keycloakClients struct {
	Admin *kcClientPkg.KeyAdminClient
	Web   *kcClientPkg.Client
}

func initKeycloak(cfg *config.Config, cacheAdapter *cache.RedisCache) (keycloakClients, error) {
	adminKC := kcClientPkg.NewAdminClient(
		cfg.KeycloakBaseURL,
		cfg.KeycloakRealm,
		cfg.KeycloakAdminClientID,
		cfg.KeycloakAdminClientSecret,
	)

	if err := adminKC.Authenticate(); err != nil {
		return keycloakClients{}, err
	}

	webKC := kcClientPkg.NewWebClient(
		cfg.KeycloakBaseURL,
		cfg.KeycloakRealm,
		cfg.KeycloakWebClientID,
		cfg.KeycloakWebClientSecret,
		cacheAdapter,
		cfg,
	)

	return keycloakClients{
		Admin: adminKC,
		Web:   webKC,
	}, nil
}
