package service

import (
	"context"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type Processor interface {
	Process(ctx context.Context, p db.Process) error
}
