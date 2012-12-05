package graval

import (
	"os"
)

type ListFormatter struct {
	files    []os.FileInfo
}

func NewListFormatter(files []os.FileInfo) *ListFormatter {
	f := new(ListFormatter)
	f.files = files
	return f
}

func (formatter *ListFormatter) Short() string {
	output := ""
	for _, file := range formatter.files {
		output += file.Name() + "\r\n"
	}
	output += "\r\n"
	return output
}

