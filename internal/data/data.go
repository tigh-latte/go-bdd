package data

import (
	"io/fs"
	"path"
)

type Dir struct {
	Prefix string
	FS     fs.FS
}

func (d *Dir) Open(parts string) (fs.File, error) {
	return d.FS.Open(path.Join(d.Prefix, parts))
}
