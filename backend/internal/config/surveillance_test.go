package config

import (
	"github.com/spf13/viper"
	"net/url"
	"testing"
)

func TestSurveillanceDBSource(t *testing.T) {
	c := Config{SurveillanceDBHost: "127.0.0.1", SurveillanceDBPort: "55434", SurveillanceDBName: "observatory", SurveillanceDBUser: "surveillance", SurveillanceDBPassword: "p@ss:/?&", SurveillanceDBEnableSSL: true}
	dsn, err := c.SurveillanceDBSource()
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	password, _ := u.User.Password()
	if password != c.SurveillanceDBPassword || u.Path != "/observatory" || u.Host != "127.0.0.1:55434" || u.Query().Get("sslmode") != "require" {
		t.Fatal("invalid isolated database DSN")
	}
	c.SurveillanceDBPassword = ""
	if _, err := c.SurveillanceDBSource(); err == nil {
		t.Fatal("missing dedicated credential must not fall back to primary")
	}
}

func TestSurveillanceConnectionEnvironment(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	setDefaults()
	viper.AutomaticEnv()
	t.Setenv("SURVEILLANCE_DB_HOST", "127.0.0.1")
	t.Setenv("SURVEILLANCE_DB_PASSWORD", "fixture-password")
	var c Config
	if err := viper.Unmarshal(&c); err != nil {
		t.Fatal(err)
	}
	if c.SurveillanceDBHost != "127.0.0.1" || c.SurveillanceDBPort != "5432" || c.SurveillanceDBName != "observatory" || c.SurveillanceDBPassword != "fixture-password" {
		t.Fatal("dedicated environment settings were not loaded")
	}
	if _, err := c.SurveillanceDBSource(); err != nil {
		t.Fatal(err)
	}
	// Existing explicit targets remain valid until their operators rename them.
	t.Setenv("SURVEILLANCE_DB_NAME", "surveillance")
	if err := viper.Unmarshal(&c); err != nil {
		t.Fatal(err)
	}
	if c.SurveillanceDBName != "surveillance" {
		t.Fatal("explicit legacy database name was overridden")
	}
}
