package data

import (
	"io/fs"
	"path"
)

type DataDir struct {
	Prefix string
	FS     fs.FS
}

func (d *DataDir) Open(parts string) (fs.File, error) {
	return d.FS.Open(path.Join(d.Prefix, parts))
}
