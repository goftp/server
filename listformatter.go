package server

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/jehiah/go-strftime"
)

type listFormatter struct {
	files []FileInfo
}

func newListFormatter(files []FileInfo) *listFormatter {
	f := new(listFormatter)
	f.files = files
	return f
}

// Short returns a string that lists the collection of files by name only,
// one per line
func (formatter *listFormatter) Short() string {
	var buf bytes.Buffer
	for _, file := range formatter.files {
		fmt.Fprintf(&buf, "%s\r\n", file.Name())
	}
	fmt.Fprintf(&buf, "\r\n")
	return buf.String()
}

// Detailed returns a string that lists the collection of files with extra
// detail, one per line
func (formatter *listFormatter) Detailed() string {
	var buf bytes.Buffer
	for _, file := range formatter.files {
		fmt.Fprintf(&buf, file.Mode().String())
		fmt.Fprintf(&buf, " 1 %s %s ", file.Owner(), file.Group())
		fmt.Fprintf(&buf, lpad(strconv.Itoa(int(file.Size())), 12))
		fmt.Fprintf(&buf, strftime.Format(" %b %d %H:%M ", file.ModTime()))
		fmt.Fprintf(&buf, "%s\r\n", file.Name())
	}
	fmt.Fprintf(&buf, "\r\n")
	return buf.String()
}

func lpad(input string, length int) (result string) {
	if len(input) < length {
		result = strings.Repeat(" ", length-len(input)) + input
	} else if len(input) == length {
		result = input
	} else {
		result = input[0:length]
	}
	return
}
