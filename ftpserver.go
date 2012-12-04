package graval

import (
	"log"
	"net"
	"strconv"
)

type FTPServer struct {
	name          string
	listenTo      string
	driverFactory FTPDriverFactory
}

func NewFTPServer(hostname string, port int, factory FTPDriverFactory) *FTPServer {
	s := new(FTPServer)
	s.listenTo = hostname + ":" + strconv.Itoa(port)
	s.name = "Go FTP Server"
	s.driverFactory = factory
	return s
}

func (ftpServer *FTPServer) Listen() (err error) {
	laddr, err := net.ResolveTCPAddr("tcp4", ftpServer.listenTo)
	if err != nil {
		log.Fatal(err)
	}
	listener, err := net.ListenTCP("tcp4", laddr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		ftpConn, err := ftpServer.Accept(listener)
		if err != nil {
			break
		}
		go ftpConn.Serve()
	}
	return
}

func (ftpServer *FTPServer) Accept(listener *net.TCPListener) (ftpConn *FTPConn, err error) {
	tcpConn, err := listener.AcceptTCP()
	if err == nil {
		ftpConn = NewFTPConn(tcpConn, ftpServer.driverFactory.NewDriver())
	}
	return
}
