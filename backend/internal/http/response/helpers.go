package response

import (
	"github.com/gin-gonic/gin"

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
	if c == nil || c.Request == nil {
		return Meta{}
	}
	return Meta{
		RequestID: observability.RequestIDFromContext(c.Request.Context()),
	}
}
