// Standalone schema initialization without requiring Keycloak or the primary database.
package main

import (
	"context"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/moh-sso-dashboard/internal/config"
	db "github.com/moh-sso-dashboard/internal/migrate"
)

func main() {
	c := config.Config{SurveillanceDBHost: os.Getenv("SURVEILLANCE_DB_HOST"), SurveillanceDBPort: os.Getenv("SURVEILLANCE_DB_PORT"), SurveillanceDBName: os.Getenv("SURVEILLANCE_DB_NAME"), SurveillanceDBUser: os.Getenv("SURVEILLANCE_DB_USER"), SurveillanceDBPassword: os.Getenv("SURVEILLANCE_DB_PASSWORD"), SurveillanceDBEnableSSL: os.Getenv("SURVEILLANCE_DB_ENABLE_SSL") == "true"}
	dsn, err := c.SurveillanceDBSource()
	if err != nil {
		log.Fatal(err)
	}
	conn, err := db.InitDB(context.Background(), db.DBConfig{Driver: "postgres", DSN: dsn, WaitTimeout: 30 * time.Second})
	if err != nil {
		log.Fatal("Cannot connect to surveillance database")
	}
	defer conn.Close()
	if err := db.MigrateDB(conn, "file://internal/db/surveillance_migrations"); err != nil {
		log.Fatal(err)
	}
	log.Print("Surveillance schema is ready")
}
