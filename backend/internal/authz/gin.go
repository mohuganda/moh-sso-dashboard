package authz

import "github.com/gin-gonic/gin"

const ContextKey = "authz"

func FromGin(c *gin.Context) (Context, bool) {
	value, ok := c.Get(ContextKey)
	if !ok {
		return Context{}, false
	}

	authContext, ok := value.(Context)
	return authContext, ok
}
