package response

import (
	"github.com/gin-gonic/gin"
	"net/http"

	"github.com/moh-sso-dashboard/internal/observability"
)

func OK[T any](c *gin.Context, status int, data T) {
	c.JSON(status, SuccessResponse[T]{
		Success: true,
		Data:    data,
		Meta:    metaFromContext(c),
	})
}

func Fail(
	c *gin.Context,
	status int,
	code string,
	message string,
) {
	c.JSON(status, ErrorResponse{
		Success: false,
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
		Meta: metaFromContext(c),
	})
}

func metaFromContext(c *gin.Context) Meta {
	if c == nil {
		return Meta{}
	}
	return MetaFromRequest(c.Request)
}

// MetaFromRequest carries the request ID resolved by host middleware into
// both host responses and standalone module callbacks.
func MetaFromRequest(r *http.Request) Meta {
	if r == nil {
		return Meta{}
	}
	return Meta{RequestID: observability.RequestIDFromContext(r.Context())}
}
