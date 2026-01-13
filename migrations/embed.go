package migrations

import (
	"embed"
	"io/fs"
)

//go:embed *.sql
var Files embed.FS

// GetFS returns the migrations filesystem
func GetFS() fs.FS {
	return Files
}
