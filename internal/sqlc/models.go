// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package sqlc

import (
	"time"
)

type Migration struct {
	ID     string
	Crdttm time.Time
}

type Path struct {
	Name string
	Path string
}

type User struct {
	ID             int64
	Handle         string
	HashedPassword string
	Clan           string
	Magic          string
	Enabled        string
}
