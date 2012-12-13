package graval

import (
	"log"
	"net"
	"strconv"
	"strings"
)

type FTPServer struct {
	name          string
	listenTo      string
	driverFactory FTPDriverFactory
}

func NewFTPServer(hostname string, port int, factory FTPDriverFactory) *FTPServer {
	s := new(FTPServer)
	s.listenTo = buildTcpString(hostname, port)
	s.name = "Go FTP Server"
	s.driverFactory = factory
	return s
}

func (ftpServer *FTPServer) ListenAndServe() (err error) {
	laddr, err := net.ResolveTCPAddr("tcp", ftpServer.listenTo)
	if err != nil {
		log.Fatal(err)
	}
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		ftpConn, err := ftpServer.accept(listener)
		if err != nil {
			break
		}
		go ftpConn.Serve()
	}
	return
}

func (ftpServer *FTPServer) accept(listener *net.TCPListener) (ftpConn *ftpConn, err error) {
	tcpConn, err := listener.AcceptTCP()
	if err == nil {
		ftpConn = newftpConn(tcpConn, ftpServer.driverFactory.NewDriver())
	}
	return
}

func buildTcpString(hostname string, port int) (result string) {
	if strings.Contains(hostname, ":") {
		// ipv6
		if port == 0 {
			result = "["+hostname+"]"
		} else {
			result = "["+hostname +"]:" + strconv.Itoa(port)
		}
	} else {
		// ipv4
		if port == 0 {
			result = hostname
		} else {
			result = hostname +":" + strconv.Itoa(port)
		}
	}
	return
}
