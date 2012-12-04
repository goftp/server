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

// NewFTPConn constructs a new object that will handle the FTP protocol over
// an active net.TCPConn. The TCP connection should already be open before
// it is handed to this functions. driver is an instance of FTPDrive that
// will handle all auth and persistence details.
func NewFTPConn(tcpConn *net.TCPConn, driver FTPDriver) *FTPConn {
	c := new(FTPConn)
	c.cwd = "/"
	c.conn = tcpConn
	c.controlReader = bufio.NewReader(tcpConn)
	c.controlWriter = bufio.NewWriter(tcpConn)
	c.driver = driver
	return c
}

// Serve starts an endless loop that reads FTP commands from the client and
// responds appropriately. terminated is a channel that will receive a true
// message when the connection closes. This loop will be running inside a
// goroutine, so use this channel to be notified when the connection can be
// cleaned up.
func (ftpConn *FTPConn) Serve(terminated chan bool) {
	log.Print("Connection Established")
	// send welcome
	ftpConn.writeMessage(220, welcomeMessage)
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

// Close will manually close this connection, even if the client isn't ready.
func (ftpConn *FTPConn) Close() {
	ftpConn.conn.Close()
	if ftpConn.data != nil {
		ftpConn.data.Close()
	}
}

// receiveLine accepts a single line FTP command and co-ordinates an
// appropriate response.
func (ftpConn *FTPConn) receiveLine(line string) {
	log.Print(line)
	command, params := ftpConn.parseLine(line)
	switch command {
	case USER:
		ftpConn.reqUser = params
		ftpConn.writeMessage(331, "User name ok, password required")
		break
	case PASS:
		if ftpConn.driver.Authenticate(ftpConn.reqUser, params) {
			ftpConn.user = ftpConn.reqUser
			ftpConn.reqUser = ""
			ftpConn.writeMessage(230, "Password ok, continue")
		} else {
			ftpConn.writeMessage(530, "Incorrect password, not logged in")
		}
		break
	default:
		ftpConn.writeMessage(500, "Command not found")
	}
}

func (ftpConn *FTPConn) parseLine(line string) (string, string) {
	params := strings.SplitN(strings.Trim(line, "\r\n"), " ", 2)
	if len(params) == 1 {
		return params[0], ""
	}
	return params[0], params[1]
}

// writeMessage will send a standard FTP response back to the client.
func (ftpConn *FTPConn) writeMessage(code int, message string) (wrote int, err error) {
	line := fmt.Sprintf("%d %s\r\n", code, message)
	wrote, err = ftpConn.controlWriter.WriteString(line)
	ftpConn.controlWriter.Flush()
	log.Print(message)
	return
}
