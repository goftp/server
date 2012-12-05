package graval

import (
	"os"
	"strconv"
)

type ListFormatter struct {
	files    []os.FileInfo
}

func NewListFormatter(files []os.FileInfo) *ListFormatter {
	f := new(ListFormatter)
	f.files = files
	return f
}

// Short returns a string that lists the collection of files by name only,
// one per line
func (formatter *ListFormatter) Short() string {
	output := ""
	for _, file := range formatter.files {
		output += file.Name() + "\r\n"
	}
	output += "\r\n"
	return output
}

// Detailed returns a string that lists the collection of files with extra
// detail, one per line
func (formatter *ListFormatter) Detailed() string {
	output := ""
	for _, file := range formatter.files {
		output += file.Mode().String()
		output += " 1 owner group "
		output += strconv.Itoa(int(file.Size()))
		output += " " + file.Name()
		output += "\r\n"
	}
	output += "\r\n"
	return output
}

