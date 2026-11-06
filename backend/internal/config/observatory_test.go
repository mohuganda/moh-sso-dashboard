package config

import (
	"net/url"
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestObservatoryDBSource(t *testing.T) {
	c := Config{ObservatoryDBHost: "::1", ObservatoryDBPort: "55434", ObservatoryDBName: "observatory", ObservatoryDBUser: "reader@role", ObservatoryDBPassword: "p@ss:/?& +", ObservatoryDBEnableSSL: true}
	dsn, err := c.ObservatoryDBSource()
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	password, _ := u.User.Password()
	if password != c.ObservatoryDBPassword || u.User.Username() != c.ObservatoryDBUser || u.Host != "[::1]:55434" || u.Path != "/observatory" || u.Query().Get("sslmode") != "require" {
		t.Fatal("incorrect database URL encoding or target")
	}
	c.ObservatoryDBPassword = ""
	if _, err := c.ObservatoryDBSource(); err == nil {
		t.Fatal("missing isolated credential must not fall back to another database")
	}
}

func TestObservatoryEnvironment(t *testing.T) {
	for _, tc := range []struct {
		name                                 string
		env                                  map[string]string
		host, port, database, user, password string
		ssl                                  bool
	}{
		{name: "unconfigured", env: map[string]string{}, port: "5432", database: "observatory", user: "surveillance"},
		{name: "removed settings ignored", env: map[string]string{"SURVEILLANCE_DB_HOST": "legacy-db", "SURVEILLANCE_DB_PASSWORD": "legacy-pass"}, port: "5432", database: "observatory", user: "surveillance"},
		{name: "canonical", env: map[string]string{"OBSERVATORY_DB_HOST": "obs-db", "OBSERVATORY_DB_PASSWORD": "obs-pass"}, host: "obs-db", port: "5432", database: "observatory", user: "surveillance", password: "obs-pass"},
		{name: "canonical ignores removed settings including SSL", env: map[string]string{"OBSERVATORY_DB_HOST": "obs-db", "OBSERVATORY_DB_PASSWORD": "obs-pass", "OBSERVATORY_DB_ENABLE_SSL": "false", "SURVEILLANCE_DB_HOST": "legacy-db", "SURVEILLANCE_DB_PASSWORD": "legacy-pass", "SURVEILLANCE_DB_ENABLE_SSL": "true"}, host: "obs-db", port: "5432", database: "observatory", user: "surveillance", password: "obs-pass"},
		{name: "explicit disable", env: map[string]string{"OBSERVATORY_DB_HOST": "", "SURVEILLANCE_DB_HOST": "legacy-db"}, port: "5432", database: "observatory", user: "surveillance"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			viper.Reset()
			t.Cleanup(viper.Reset)
			for _, prefix := range []string{"OBSERVATORY_DB_", "SURVEILLANCE_DB_"} {
				for _, key := range []string{"HOST", "PORT", "NAME", "USER", "PASSWORD", "ENABLE_SSL"} {
					t.Setenv(prefix+key, "")
					if err := os.Unsetenv(prefix + key); err != nil {
						t.Fatal(err)
					}
				}
			}
			for key, value := range tc.env {
				t.Setenv(key, value)
			}
			setDefaults()
			viper.AutomaticEnv()
			viper.AllowEmptyEnv(true)
			for key, env := range envBindings() {
				if err := viper.BindEnv(key, env); err != nil {
					t.Fatal(err)
				}
			}
			var c Config
			if err := viper.Unmarshal(&c); err != nil {
				t.Fatal(err)
			}
			normalizeConfig(&c)
			if c.ObservatoryDBHost != tc.host || c.ObservatoryDBPort != tc.port || c.ObservatoryDBName != tc.database || c.ObservatoryDBUser != tc.user || c.ObservatoryDBPassword != tc.password || c.ObservatoryDBEnableSSL != tc.ssl {
				t.Fatal("incorrect effective Observatory settings")
			}
			if c.ObservatoryDBHost != "" {
				_, err := c.ObservatoryDBSource()
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestObservatoryRejectsPrimaryDatabaseTarget(t *testing.T) {
	c := Config{DBHost: "postgres", DBPort: "5432", DBName: "portal", ObservatoryDBHost: "postgres", ObservatoryDBPort: "5432", ObservatoryDBName: "portal", ObservatoryDBUser: "surveillance", ObservatoryDBPassword: "fixture-only"}
	if _, err := c.ObservatoryDBSource(); err == nil {
		t.Fatal("Observatory must not target the primary database")
	}
	c.ObservatoryDBName = "observatory"
	if _, err := c.ObservatoryDBSource(); err != nil {
		t.Fatalf("separate database on the same server should be supported: %v", err)
	}
}
