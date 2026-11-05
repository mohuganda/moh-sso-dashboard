package config

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// An empty host leaves existing installations on their primary database.
func (c *Config) SurveillanceDBSource() (string, error) {
	for name, value := range map[string]string{"HOST": c.SurveillanceDBHost, "PORT": c.SurveillanceDBPort, "NAME": c.SurveillanceDBName, "USER": c.SurveillanceDBUser, "PASSWORD": c.SurveillanceDBPassword} {
		if strings.TrimSpace(value) == "" {
			return "", fmt.Errorf("SURVEILLANCE_DB_%s is required for an isolated surveillance database", name)
		}
	}
	mode := "disable"
	if c.SurveillanceDBEnableSSL {
		mode = "require"
	}
	u := url.URL{Scheme: "postgresql", User: url.UserPassword(c.SurveillanceDBUser, c.SurveillanceDBPassword), Host: net.JoinHostPort(c.SurveillanceDBHost, c.SurveillanceDBPort), Path: "/" + c.SurveillanceDBName}
	q := url.Values{}
	q.Set("sslmode", mode)
	u.RawQuery = q.Encode()
	return u.String(), nil
}
