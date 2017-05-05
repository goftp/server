package server

import (
	"fmt"
	"log"
)

type Logger interface {
	Print(sessionId string, message interface{})
	Printf(sessionId string, format string, v ...interface{})
	PrintCommand(sessionId string, command string, params string)
	PrintResponse(sessionId string, code int, message string)
}

// Use an instance of this to log in a standard format
type StdLogger struct{}

func (logger *StdLogger) Print(sessionId string, message interface{}) {
	log.Printf("%s  %s", sessionId, message)
}

func (logger *StdLogger) Printf(sessionId string, format string, v ...interface{}) {
	logger.Print(sessionId, fmt.Sprintf(format, v...))
}

func (logger *StdLogger) PrintCommand(sessionId string, command string, params string) {
	if command == "PASS" {
		log.Printf("%s > PASS ****", sessionId)
	} else {
		log.Printf("%s > %s %s", sessionId, command, params)
	}
}

func (logger *StdLogger) PrintResponse(sessionId string, code int, message string) {
	log.Printf("%s < %d %s", sessionId, code, message)
}
