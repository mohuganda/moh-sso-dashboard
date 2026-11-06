// Standalone schema initialization without requiring Keycloak or the primary database.
package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"github.com/moh-sso-dashboard/internal/config"
	db "github.com/moh-sso-dashboard/internal/migrate"
)

func main() {
	setting := func(key, fallback string) string {
		if value, ok := os.LookupEnv("OBSERVATORY_DB_" + key); ok {
			return value
		}
		if value, ok := os.LookupEnv("SURVEILLANCE_DB_" + key); ok {
			return value
		}
		return fallback
	}
	ssl, err := strconv.ParseBool(setting("ENABLE_SSL", "false"))
	if err != nil {
		log.Fatal("OBSERVATORY_DB_ENABLE_SSL must be true or false")
	}
	c := config.Config{ObservatoryDBHost: setting("HOST", ""), ObservatoryDBPort: setting("PORT", "5432"), ObservatoryDBName: setting("NAME", "observatory"), ObservatoryDBUser: setting("USER", "surveillance"), ObservatoryDBPassword: setting("PASSWORD", ""), ObservatoryDBEnableSSL: ssl}

	dsn, err := c.ObservatoryDBSource()
	if err != nil {
		log.Fatal(err)
	}
	conn, err := db.InitDB(context.Background(), db.DBConfig{Driver: "postgres", DSN: dsn, WaitTimeout: 30 * time.Second})
	if err != nil {
		log.Fatal("Cannot connect to observatory database")
	}
	defer conn.Close()
	if err := db.MigrateDB(conn, "file://internal/db/surveillance_migrations"); err != nil {
		log.Fatal(err)
	}
	log.Print("Surveillance schema is ready")
}
