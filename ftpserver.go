package raval

import (
	"net"
)

type FTPServer struct {
	name        string
	listener    *net.TCPListener
	connections *Array
}

func NewFTPServer(listener *net.TCPListener) *FTPServer {
	s := new(FTPServer)
	s.name = "Go FTP Server"
	s.listener = listener
	s.connections = new(Array)
	return s
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
		ftpConn = NewFTPConn(tcpConn)
	}
	return
}

