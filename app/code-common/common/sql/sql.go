package sql

import "database/sql"

// 将 string 转为 sql.NullString
func StringToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{
			String: s,
			Valid:  false,
		}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

// 将 sql.NullString 转为 string
func NullStringToString(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}
