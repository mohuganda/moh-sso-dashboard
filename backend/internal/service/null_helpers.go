package service

import "database/sql"

func nullStringValue(v sql.NullString) string {
	if !v.Valid {
		return ""
	}

	return v.String
}
