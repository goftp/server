package raval

import (
	"net"
)

type FTPServer struct {
	name        string
	listener    *net.TCPListener
	connections []*FTPConn
}

func NewFTPServer(listener *net.TCPListener) *FTPServer {
	s := new(FTPServer)
	s.name = "Go FTP Server"
	s.listener = listener
	return s
}

func (ftpServer *FTPServer) Listen() (err error) {
	for {
		ftpConn, err := ftpServer.Accept()
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

func (ftpServer *FTPServer) Accept() (ftpConn *FTPConn, err error) {
	tcpConn, err := ftpServer.listener.AcceptTCP()
	if err == nil {
		ftpConn = NewFTPConn(tcpConn)
	}
	return
}

