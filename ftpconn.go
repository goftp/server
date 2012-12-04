package graval

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

const (
	welcomeMessage = "Welcome to the Go FTP Server"
	USER           = "USER"
	PASS           = "PASS"
)

type FTPConn struct {
	cwd           string
	conn          *net.TCPConn
	controlReader *bufio.Reader
	controlWriter *bufio.Writer
	data          *net.TCPConn
	driver        FTPDriver
	reqUser       string
	user          string
}

func NewFTPConn(tcpConn *net.TCPConn, driver FTPDriver) *FTPConn {
	c := new(FTPConn)
	c.cwd = "/"
	c.conn = tcpConn
	c.controlReader = bufio.NewReader(tcpConn)
	c.controlWriter = bufio.NewWriter(tcpConn)
	c.driver = driver
	return c
}

func (ftpConn *FTPConn) Serve(terminated chan bool) {
	log.Print("Connection Established")
	// send welcome
	ftpConn.writeMessage(getMessageFormat(220), welcomeMessage)
	// read commands
	for {
		line, err := ftpConn.controlReader.ReadString('\n')
		if err == nil {
			ftpConn.receiveLine(line)
		}
	}
	terminated <- true
	log.Print("Connection Terminated")
}

func (ftpConn *FTPConn) Close() {
	ftpConn.conn.Close()
	if ftpConn.data != nil {
		ftpConn.data.Close()
	}
}

func (ftpConn *FTPConn) receiveLine(line string) {
	log.Print(line)
	params := strings.Split(strings.Trim(line, "\r\n"), " ")
	count := len(params)
	if count > 0 {
		command := params[0]
		switch command {
		case USER:
			ftpConn.reqUser = params[1]
			ftpConn.writeMessage(getMessageFormat(331), "User name ok, password required")
			break
		case PASS:
			if ftpConn.driver.Authenticate(ftpConn.reqUser, params[1]) {
				ftpConn.user = ftpConn.reqUser
				ftpConn.reqUser = ""
				ftpConn.writeMessage(getMessageFormat(230), "Password ok, continue")
			} else {
				ftpConn.writeMessage(getMessageFormat(530), "Incorrect password, not logged in")
			}
			break
		default:
			ftpConn.writeMessage(getMessageFormat(500), "Command not found")
		}
	} else {
		ftpConn.writeMessage(getMessageFormat(500), "Syntax error, zero parameters")
	}
}

func (ftpConn *FTPConn) writeMessage(messageFormat string, v ...interface{}) (wrote int, err error) {
	message := fmt.Sprintf(messageFormat, v...)
	wrote, err = ftpConn.controlWriter.WriteString(message)
	ftpConn.controlWriter.Flush()
	log.Print(message)
	return
}

func getMessageFormat(command int) (messageFormat string) {
	switch command {
	case 220:
		messageFormat = "220 %s"
		break
	case 230:
		messageFormat = "230 %s"
		break
	case 331:
		messageFormat = "331 %s"
		break
	case 500:
		messageFormat = "500 %s"
		break
	case 530:
		messageFormat = "530 %s"
		break
	}
	return messageFormat + "\r\n"
}
