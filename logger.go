package server

import (
	"fmt"
	"log"
)

// Use an instance of this to log in a standard format
type Logger struct {
	sessionId string
}

func newLogger(id string) *Logger {
	l := new(Logger)
	l.sessionId = id
	return l
}

func (logger *Logger) Print(message interface{}) {
	log.Printf("%s   %s", logger.sessionId, message)
}

func (logger *Logger) Printf(format string, v ...interface{}) {
	logger.Print(fmt.Sprintf(format, v...))
}

func (logger *Logger) PrintCommand(command string, params string) {
	if command == "PASS" {
		log.Printf("%s > PASS ****", logger.sessionId)
	} else {
		log.Printf("%s > %s %s", logger.sessionId, command, params)
	}
}

func (logger *Logger) PrintResponse(code int, message string) {
	log.Printf("%s < %d %s", logger.sessionId, code, message)
}
