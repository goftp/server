// An example FTP server build on top of go-raval. go-raval handles the details
// of the FTP protocol, we just provide a basic in-memory persistence driver.
package main

import (
	"github.com/yob/graval"
	"log"
	"time"
)

const (
	fileOne = "This is the first file available for download.\n\nBy Jàmes"
	fileTwo = "This is file number two.\n\n2012-12-04"
)

type MemDriver struct{}

func (driver *MemDriver) Authenticate(user string, pass string) bool {
	return user == "test" && pass == "1234"
}
func (driver *MemDriver) Bytes(path string) (bytes int) {
	switch path {
	case "/one.txt":
		bytes = len(fileOne)
		break
	case "/files/two.txt":
		bytes = len(fileTwo)
		break
	default:
		bytes = -1
	}
	return
}
func (driver *MemDriver) ModifiedTime(path string) time.Time {
	return time.Now()
}
func (driver *MemDriver) ChangeDir(path string) bool {
	return path == "/" || path == "/files"
}
func (driver *MemDriver) DirContents(path string) bool {
	return false
}
func (driver *MemDriver) DeleteDir(path string) bool {
	return false
}
func (driver *MemDriver) DeleteFile(path string) bool {
	return false
}
func (driver *MemDriver) Rename(fromPath string, toPath string) bool {
	return false
}
func (driver *MemDriver) MakeDir(path string) bool {
	return false
}
func (driver *MemDriver) GetFile(path string) int {
	return -1
}
func (driver *MemDriver) PutFile(destPath string, tmpPath string) int {
	return -1
}

type MemDriverFactory struct{}

func (factory *MemDriverFactory) NewDriver() graval.FTPDriver {
	return &MemDriver{}
}

func main() {
	factory := &MemDriverFactory{}
	ftpServer := graval.NewFTPServer("localhost", 3000, factory)
	log.Fatal(ftpServer.Listen())
}
