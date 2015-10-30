package server

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"
)

// A data socket is used to send non-control data between the client and
// server.
type DataSocket interface {
	Host() string

	Port() int

	// the standard io.Reader interface
	Read(p []byte) (n int, err error)

	// the standard io.Writer interface
	Write(p []byte) (n int, err error)

	// the standard io.Closer interface
	Close() error
}

type ftpActiveSocket struct {
	conn   *net.TCPConn
	host   string
	port   int
	logger *Logger
}

func newActiveSocket(host string, port int, logger *Logger) (DataSocket, error) {
	connectTo := buildTcpString(host, port)
	logger.Print("Opening active data connection to " + connectTo)
	raddr, err := net.ResolveTCPAddr("tcp", connectTo)
	if err != nil {
		logger.Print(err)
		return nil, err
	}
	tcpConn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		logger.Print(err)
		return nil, err
	}
	socket := new(ftpActiveSocket)
	socket.conn = tcpConn
	socket.host = host
	socket.port = port
	socket.logger = logger
	return socket, nil
}

func (socket *ftpActiveSocket) Host() string {
	return socket.host
}

func (socket *ftpActiveSocket) Port() int {
	return socket.port
}

func (socket *ftpActiveSocket) Read(p []byte) (n int, err error) {
	return socket.conn.Read(p)
}

func (socket *ftpActiveSocket) Write(p []byte) (n int, err error) {
	return socket.conn.Write(p)
}

func (socket *ftpActiveSocket) Close() error {
	return socket.conn.Close()
}

type ftpPassiveSocket struct {
	conn    *net.TCPConn
	port    int
	host    string
	ingress chan []byte
	egress  chan []byte
	logger  *Logger
	wg      sync.WaitGroup
}

func newPassiveSocket(host string, logger *Logger) (DataSocket, error) {
	socket := new(ftpPassiveSocket)
	socket.ingress = make(chan []byte)
	socket.egress = make(chan []byte)
	socket.logger = logger
	socket.host = host
	if err := socket.GoListenAndServe(); err != nil {
		return nil, err
	}
	return socket, nil
}

func (socket *ftpPassiveSocket) Host() string {
	return socket.host
}

func (socket *ftpPassiveSocket) Port() int {
	return socket.port
}

func (socket *ftpPassiveSocket) Read(p []byte) (n int, err error) {
	if socket.waitForOpenSocket() == false {
		return 0, errors.New("data socket unavailable")
	}
	return socket.conn.Read(p)
}

func (socket *ftpPassiveSocket) Write(p []byte) (n int, err error) {
	if socket.waitForOpenSocket() == false {
		return 0, errors.New("data socket unavailable")
	}
	return socket.conn.Write(p)
}

func (socket *ftpPassiveSocket) Close() error {
	//socket.logger.Print("closing passive data socket")
	if socket.conn != nil {
		return socket.conn.Close()
	}
	return nil
}

func (socket *ftpPassiveSocket) GoListenAndServe() (err error) {
	laddr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	if err != nil {
		socket.logger.Print(err)
		return
	}
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		socket.logger.Print(err)
		return
	}
	add := listener.Addr()
	parts := strings.Split(add.String(), ":")
	port, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		socket.logger.Print(err)
		return
	}

	socket.port = port
	socket.wg.Add(1)
	go func() {
		tcpConn, err := listener.AcceptTCP()
		socket.wg.Done()
		if err != nil {
			socket.logger.Print(err)
			return
		}
		socket.conn = tcpConn
	}()
	return nil
}

func (socket *ftpPassiveSocket) waitForOpenSocket() bool {
	if socket.conn != nil {
		return true
	}
	socket.wg.Wait()
	return socket.conn != nil
}
