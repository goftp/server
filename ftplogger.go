package graval

import (
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
