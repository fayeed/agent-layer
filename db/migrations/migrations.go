package migrations

import _ "embed"

//go:embed 0001_v0_core.sql
var v0CoreSQL string

func V0CoreSQL() string {
	return v0CoreSQL
}
