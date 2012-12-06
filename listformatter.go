package graval

import (
	"github.com/jehiah/go-strftime"
	"os"
	"strconv"
	"strings"
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
		output += lpad(strconv.Itoa(int(file.Size())), 12)
		output += " " + strftime.Format("%b %d %H:%M", file.ModTime())
		output += " " + file.Name()
		output += "\r\n"
	}
	output += "\r\n"
	return output
}

func lpad(input string, length int) (result string) {
	if len(input) < length {
		result = strings.Repeat(" ", length-len(input))+input
	} else if len(input) == length {
		result = input
	} else {
		result = input[0:length]
	}
	return
}
