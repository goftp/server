package graval

import (
	"time"
)

type FTPDriverFactory interface {
	NewDriver() FTPDriver
}

type FTPDriver interface {
	// params  - username, password
	// returns - true if the provided details are valid
	Authenticate(string, string) bool

	// params  - a file path
	// returns - an int with the number of bytes in the file or -1 if the file
	//           doesn't exist
	Bytes(string) int

	// params  - a file path
	// returns - a time indicating when the requested path was last modified
	ModifiedTime(string) time.Time

	// params  - path
	// returns - true if the current user is permitted to change to the
	//           requested path
	ChangeDir(string) bool

	// params  - path
	// returns - a collection of items describing the contents of the requested
	//           path
	DirContents(string) bool

	// params  - path
	// returns - true if the directory was deleted
	DeleteDir(string) bool

	// params  - path
	// returns - true if the file was deleted
	DeleteFile(string) bool

	// params  - from_path, to_path
	// returns - true if the file was renamed
	Rename(string, string) bool

	// params  - path
	// returns - true if the new directory was created
	MakeDir(string) bool

	// params  - path
	// returns - the file to send back to the current user
	GetFile(string) int

	// params  - desination path, temp file path
	// returns - the number of bytes saved to the desination path or -1 if
	//           there was an issue
	PutFile(string, string) int
}
