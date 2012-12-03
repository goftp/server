package raval

import (
	"net"
	"strconv"
	"log"
)

type FTPServer struct {
	name        string
	listenTo    string
	connections []*FTPConn
}

func NewFTPServer(hostname string, port int) *FTPServer {
	s := new(FTPServer)
	s.listenTo = hostname + ":" + strconv.Itoa(port)
	s.name = "Go FTP Server"
	return s
}

func (ftpServer *FTPServer) Listen() (err error) {
	for {
		laddr, err := net.ResolveTCPAddr("tcp4", ftpServer.listenTo)
		if err != nil {
			log.Fatal(err)
		}
		listener, err := net.ListenTCP("tcp4", laddr)
		if err != nil {
			log.Fatal(err)
		}
		ftpConn, err := ftpServer.Accept(listener)
		if err != nil {
			break
		}
		ftpServer.connections = append(ftpServer.connections, ftpConn)
		terminated := make(chan bool)
		go ftpConn.Serve(terminated)
		<-terminated
		ftpServer.removeConnection(ftpConn)
		ftpConn.Close()
	}
	return
}

func (ftpServer *FTPServer) removeConnection(ftpConn *FTPConn) {
	i := ftpServer.indexOfConnection(ftpConn)
	ftpServer.connections[i] = ftpServer.connections[len(ftpServer.connections)-1]
	ftpServer.connections = ftpServer.connections[0:len(ftpServer.connections)-1]
	return
}

func (ftpServer *FTPServer) indexOfConnection(ftpConn *FTPConn) int {
	for p, v := range ftpServer.connections {
		if (v == ftpConn) {
			return p
		}
	}
	return -1
}

func (ftpServer *FTPServer) Accept(listener *net.TCPListener) (ftpConn *FTPConn, err error) {
	tcpConn, err := listener.AcceptTCP()
	if err == nil {
		ftpConn = NewFTPConn(tcpConn)
	}
	return
}

