package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

const (
	rootDir        = "/"
	welcomeMessage = "Welcome to the Go FTP Server"
	USER           = "USER"
	PASS           = "PASS"
)

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
	}
	return messageFormat + "\r\n"
}

type Array struct {
	container []interface{}
}

func (a *Array) Append(object interface{}) {
	if a.container == nil {
		a.container = make([]interface{}, 0)
	}
	newContainer := make([]interface{}, len(a.container)+1)
	copy(newContainer, a.container)
	newContainer[len(newContainer)-1] = object
	a.container = newContainer
}

func (a *Array) Remove(object interface{}) (result bool) {
	result = false
	newContainer := make([]interface{}, len(a.container)-1)
	i := 0
	for _, target := range a.container {
		if target != object {
			newContainer[i] = target
		} else {
			result = true
		}
		i++
	}
	return
}

type FTPServer struct {
	name        string
	listener    *net.TCPListener
	connections *Array
}

func (ftpServer *FTPServer) Listen() (err error) {
	for {
		ftpConn, err := ftpServer.Accept()
		if err != nil {
			break
		}
		ftpServer.connections.Append(ftpConn)
		terminated := make(chan bool)
		go ftpConn.Serve(terminated)
		<-terminated
		ftpServer.connections.Remove(ftpConn)
		ftpConn.Close()
	}
	return
}

func (ftpServer *FTPServer) Accept() (ftpConn *FTPConn, err error) {
	tcpConn, err := ftpServer.listener.AcceptTCP()
	if err == nil {
		controlReader := bufio.NewReader(tcpConn)
		controlWriter := bufio.NewWriter(tcpConn)
		ftpConn = &FTPConn{
			rootDir,
			tcpConn,
			controlReader,
			controlWriter,
			nil,
		}
	}
	return
}

type FTPConn struct {
	cwd           string
	control       *net.TCPConn
	controlReader *bufio.Reader
	controlWriter *bufio.Writer
	data          *net.TCPConn
}

func (ftpConn *FTPConn) WriteMessage(messageFormat string, v ...interface{}) (wrote int, err error) {
	message := fmt.Sprintf(messageFormat, v...)
	wrote, err = ftpConn.controlWriter.WriteString(message)
	ftpConn.controlWriter.Flush()
	log.Print(message)
	return
}

func (ftpConn *FTPConn) Serve(terminated chan bool) {
	log.Print("Connection Established")
	// send welcome
	ftpConn.WriteMessage(getMessageFormat(220), welcomeMessage)
	// read commands
	for {
		line, err := ftpConn.controlReader.ReadString('\n')
		if err != nil {
			break
		}
		log.Print(line)
		params := strings.Split(strings.Trim(line, "\r\n"), " ")
		count := len(params)
		if count > 0 {
			command := params[0]
			switch command {
			case USER:
				ftpConn.WriteMessage(getMessageFormat(331), "User name ok, password required")
				break
			case PASS:
				ftpConn.WriteMessage(getMessageFormat(230), "Password ok, continue")
				break
			default:
				ftpConn.WriteMessage(getMessageFormat(500), "Command not found")
			}
		} else {
			ftpConn.WriteMessage(getMessageFormat(500), "Syntax error, zero parameters")
		}
	}
	terminated <- true
	log.Print("Connection Terminated")
}

func (ftpConn *FTPConn) Close() {
	ftpConn.control.Close()
	if ftpConn.data != nil {
		ftpConn.data.Close()
	}
}

func main() {
	laddr, err := net.ResolveTCPAddr("tcp", "localhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	listener, err := net.ListenTCP("tcp4", laddr)
	if err != nil {
		log.Fatal(err)
	}
	ftpServer := &FTPServer{
		"Go FTP Server",
		listener,
		new(Array),
	}
	log.Fatal(ftpServer.Listen())
}
