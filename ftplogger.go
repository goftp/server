package graval

import (
	"fmt"
	"log"
)

// Use an instance of this to log in a standard format
type ftpLogger struct {
	sessionId  string
}

func newFtpLogger(id string) *ftpLogger {
	l := new(ftpLogger)
	l.sessionId = id
	return l
}

func (logger *ftpLogger) Print(message string) {
	log.Printf("%s - %s", logger.sessionId, message)
}

func (logger *ftpLogger) Printf(format string, v ...interface{}) {
	logger.Print(fmt.Sprintf(format, v...))
}

func (logger *ftpLogger) PrintCommand(command string, params string) {
	if command == "PASS" {
		logger.Print("PASS ****")
	} else {
		logger.Printf("%s %s", command, params)
	}
}

func (logger *ftpLogger) PrintResponse(code int, message string) {
	logger.Printf("%d %s", code, message)
}
