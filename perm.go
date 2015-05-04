package server

import "os"

type Perm interface {
	GetOwner(string) (string, error)
	GetGroup(string) (string, error)
	GetMode(string) (os.FileMode, error)

	ChOwner(string, string) error
	ChGroup(string, string) error
	ChMode(string, os.FileMode) error
}
