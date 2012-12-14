// An example FTP server build on top of graval. graval handles the details
// of the FTP protocol, we just provide a persistence driver for rackspace
// cloud files.
//
// If you're looking to create a custom graval driver, this example is a
// reasonable starting point. I suggest copying this file and changing the
// function bodies as required.
package main

import (
	"errors"
	"github.com/ncw/swift"
	"github.com/yob/graval"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// A minimal driver for graval that stores all data on rackspace cloudfiles. The
// authentication are ignored, any username and password will work.
//
// This really just exists as a minimal demonstration of the interface graval
// drivers are required to implement.
type SwiftDriver struct{
	connection *swift.Connection
	container  string
	user       string
}

func (driver *SwiftDriver) Authenticate(user string, pass string) bool {
	log.Printf("Authenticate: %s %s", user, pass)
	driver.user = user
	return true
}
func (driver *SwiftDriver) Bytes(path string) (bytes int) {
	log.Printf("Bytes: %s", path)
	bytes = -1
	return
}
func (driver *SwiftDriver) ModifiedTime(path string) (time.Time, error) {
	log.Printf("ModifiedTime: %s", path)
	return time.Now(), nil
}
func (driver *SwiftDriver) ChangeDir(path string) bool {
	log.Printf("ChangeDir: %s", path)
	return false
}
func (driver *SwiftDriver) DirContents(path string) (files []os.FileInfo) {
	path = scoped_path_with_trailing_slash(driver.user, path)
	log.Printf("DirContents: %s", path)
	opts    := &swift.ObjectsOpts{Prefix:path}
	objects, err := driver.connection.ObjectsAll(driver.container, opts)
	if err != nil {
		return // error connecting to cloud files
	}
	for _, object := range objects {
		tail  := strings.Replace(object.Name, path, "", 1)
        basename := strings.Split(tail, "/")[0]
		if object.ContentType == "application/directory" {
			files = append(files, graval.NewDirItem(basename))
		} else  {
			files = append(files, graval.NewFileItem(basename, int(object.Bytes)))
		}
	}
	return
}

func (driver *SwiftDriver) DeleteDir(path string) bool {
	log.Printf("DeleteDir: %s", path)
	return false
}
func (driver *SwiftDriver) DeleteFile(path string) bool {
	log.Printf("DeleteFile: %s", path)
	return false
}
func (driver *SwiftDriver) Rename(fromPath string, toPath string) bool {
	log.Printf("Rename: %s %s", fromPath, toPath)
	return false
}
func (driver *SwiftDriver) MakeDir(path string) bool {
	path = scoped_path_with_trailing_slash(driver.user, path)
	log.Printf("MakeDir: %s", path)
	opts    := &swift.ObjectsOpts{Prefix:path}
	objects, err := driver.connection.ObjectNames(driver.container, opts)
	if err != nil {
		return false // error connection to cloud files
	}
	if len(objects) > 0 {
		return false // the dir already exists
	}
	driver.connection.ObjectPutString(driver.container, path, "", "application/directory")
	return true
}
func (driver *SwiftDriver) GetFile(path string) (data string, err error) {
	log.Printf("GetFile: %d", len(data))
	return "", errors.New("foo")
}
func (driver *SwiftDriver) PutFile(destPath string, data io.Reader) bool {
	destPath = scoped_path(driver.user, destPath)
	log.Printf("PutFile: %s", destPath)
	contents, err := ioutil.ReadAll(data)
	if err != nil {
		return false
	}
	err = driver.connection.ObjectPutBytes(driver.container, destPath, contents, "application/octet-stream")
	if err != nil {
		return false
	}
	return true
}

func scoped_path_with_trailing_slash(user string, path string) string {
	path = scoped_path(user, path)
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	if path == "/" {
		return ""
	}
	return path
}

func scoped_path(user string, path string) string {
	if path == "/" {
		path = ""
	}
	return filepath.Join("/", user, path)
}

// graval requires a factory that will create a new driver instance for each
// client connection. Generally the factory will be fairly minimal. This is
// a good place to read any required config for your driver.
type SwiftDriverFactory struct{}

func (factory *SwiftDriverFactory) NewDriver() (graval.FTPDriver, error) {
	driver := &SwiftDriver{}
	driver.container  = "rba-uploads"
	driver.connection = &swift.Connection{
		UserName: os.Getenv("UserName"),
		ApiKey:   os.Getenv("ApiKey"),
		AuthUrl:  os.Getenv("AuthUrl"),
	}
	err := driver.connection.Authenticate()
	if err != nil {
		return nil, err
	}
	return driver, nil
}

// it's alive!
func main() {
	factory := &SwiftDriverFactory{}
	ftpServer := graval.NewFTPServer(&graval.FTPServerOpts{ Factory: factory })
	err := ftpServer.ListenAndServe()
	if err != nil {
		log.Fatal("Error starting server!")
		log.Fatal(err)
	}
}
