package graval

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	welcomeMessage = "Welcome to the Go FTP Server"
)

type ftpConn struct {
	conn          *net.TCPConn
	controlReader *bufio.Reader
	controlWriter *bufio.Writer
	dataConn      ftpDataSocket
	driver        FTPDriver
	namePrefix    string
	reqUser       string
	user          string
	renameFrom    string
}

// NewftpConn constructs a new object that will handle the FTP protocol over
// an active net.TCPConn. The TCP connection should already be open before
// it is handed to this functions. driver is an instance of FTPDrive that
// will handle all auth and persistence details.
func newftpConn(tcpConn *net.TCPConn, driver FTPDriver) *ftpConn {
	c := new(ftpConn)
	c.namePrefix = "/"
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
func (ftpConn *ftpConn) Serve() {
	log.Print("Connection Established")
	// send welcome
	ftpConn.writeMessage(220, welcomeMessage)
	// read commands
	for {
		line, err := ftpConn.controlReader.ReadString('\n')
		if err != nil {
			break
		}
		ftpConn.receiveLine(line)
	}
	log.Print("Connection Terminated")
}

// Close will manually close this connection, even if the client isn't ready.
func (ftpConn *ftpConn) Close() {
	ftpConn.conn.Close()
	if ftpConn.dataConn != nil {
		ftpConn.dataConn.Close()
	}
}

// receiveLine accepts a single line FTP command and co-ordinates an
// appropriate response.
func (ftpConn *ftpConn) receiveLine(line string) {
	log.Print(line)
	command, param := ftpConn.parseLine(line)
	cmdObj := commands[command]
	if cmdObj != nil {
		cmdObj.Execute(ftpConn, param)
		return
	}
	switch command {
	case "RNFR":
		ftpConn.cmdRnfr(param)
	case "RNTO":
		ftpConn.cmdRnto(param)
	default:
		ftpConn.writeMessage(500, "Command not found")
	}
}

// cmdRnfr responds to the RNFR FTP command. It's the first of two commands
// required for a client to rename a file.
func (ftpConn *ftpConn) cmdRnfr(param string) {
	ftpConn.renameFrom = ftpConn.buildPath(param)
	ftpConn.writeMessage(350, "Requested file action pending further information.")
}

// cmdRnto responds to the RNTO FTP command. It's the second of two commands
// required for a client to rename a file.
func (ftpConn *ftpConn) cmdRnto(param string) {
	toPath := ftpConn.buildPath(param)
	if ftpConn.driver.Rename(ftpConn.renameFrom, toPath) {
		ftpConn.writeMessage(250, "File renamed")
	} else {
		ftpConn.writeMessage(550, "Action not taken")
	}
}

func (ftpConn *ftpConn) parseLine(line string) (string, string) {
	params := strings.SplitN(strings.Trim(line, "\r\n"), " ", 2)
	if len(params) == 1 {
		return params[0], ""
	}
	return params[0], params[1]
}

// writeMessage will send a standard FTP response back to the client.
func (ftpConn *ftpConn) writeMessage(code int, message string) (wrote int, err error) {
	line := fmt.Sprintf("%d %s\r\n", code, message)
	log.Print(line)
	wrote, err = ftpConn.controlWriter.WriteString(line)
	ftpConn.controlWriter.Flush()
	return
}

// buildPath takes a client supplied path or filename and generates a safe
// absolute path withing their account sandbox.
//
//    buildpath("/")
//    => "/"
//    buildpath("one.txt")
//    => "/one.txt"
//    buildpath("/files/two.txt")
//    => "/files/two.txt"
//    buildpath("files/two.txt")
//    => "files/two.txt"
//    buildpath("/../../../../etc/passwd")
//    => "/etc/passwd"
//
// The driver implementation is responsible for deciding how to treat this path.
// Obviously they MUST NOT just read the path off disk. The probably want to
// prefix the path with something to scope the users access to a sandbox.
func (ftpConn *ftpConn) buildPath(filename string) (fullPath string) {
	if len(filename) > 0 && filename[0:1] == "/" {
		fullPath = filepath.Clean(filename)
	} else if len(filename) > 0 && filename != "-a" {
		fullPath = filepath.Clean(ftpConn.namePrefix + "/" + filename)
	} else {
		fullPath = filepath.Clean(ftpConn.namePrefix)
	}
	fullPath = strings.Replace(fullPath, "//", "/", -1)
	return
}

// sendOutofbandData will send a string to the client via the currently open
// data socket. Assumes the socket is open and ready to be used.
func (ftpConn *ftpConn) sendOutofbandData(data string) {
	bytes := len(data)
	ftpConn.dataConn.Write([]byte(data))
	ftpConn.dataConn.Close()
	message := "Closing data connection, sent " + strconv.Itoa(bytes) + " bytes"
	ftpConn.writeMessage(226, message)
}
