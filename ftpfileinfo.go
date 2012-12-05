package graval

import (
	"os"
	"time"
)

type FTPFileInfo struct {
	name      string
	bytes     int64
	mode      os.FileMode
}

func (info *FTPFileInfo) Name() string {
	return info.name
}

func (info *FTPFileInfo) Size() int64 {
	return info.bytes
}

func (info *FTPFileInfo) Mode() os.FileMode {
	return info.mode
}

func (info *FTPFileInfo) ModTime() time.Time {
	return time.Now()
}

func (info *FTPFileInfo) IsDir() bool {
	return (info.mode | os.ModeDir) == os.ModeDir
}

func (info *FTPFileInfo) Sys() interface{} {
	return nil
}

// NewDirItem creates a new FTPFileInfo that represents a single diretory. Use
// this function to build the response to DirContents() in your FTPDriver
// implementation.
func NewDirItem(name string) os.FileInfo {
	d := new(FTPFileInfo)
	d.name = name
	d.bytes = int64(0)
	d.mode = os.ModeDir | 666
	return d
}

// NewFileItem creates a new FTPFileInfo that represents a single file. Use
// this function to build the response to DirContents() in your FTPDriver
// implementation.
func NewFileItem(name string, bytes int) os.FileInfo {
	f := new(FTPFileInfo)
	f.name = name
	f.bytes = int64(bytes)
	f.mode = 666
	return f
}
