package ratelimit

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

func Key(parts ...string) string {
	clean := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		clean = append(clean, part)
	}
	return "rl:" + strings.Join(clean, ":")
}

func ByIP(c *gin.Context) string {
	return identity("ip", c.ClientIP())
}

func ByUser(c *gin.Context) string {
	if v, ok := c.Get("user_id"); ok {
		return identity("user", fmt.Sprint(v))
	}
	return ByIP(c)
}

func ByIPAndUsername(c *gin.Context) string {
	username := c.PostForm("username")
	if username == "" {
		username = c.Query("username")
	}
	return identity("ip", c.ClientIP(), "username", username)
}

func ByPolicyUser(policyName string) func(*gin.Context) string {
	return func(c *gin.Context) string {
		return Key(policyName, ByUser(c))
	}
}

func ByPolicyIP(policyName string) func(*gin.Context) string {
	return func(c *gin.Context) string {
		return Key(policyName, ByIP(c))
	}
}

func ByPolicyIPAndUsername(policyName string) func(*gin.Context) string {
	return func(c *gin.Context) string {
		return Key(policyName, ByIPAndUsername(c))
	}
}

func identity(parts ...string) string {
	clean := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		clean = append(clean, part)
	}
	return strings.Join(clean, ":")
}
