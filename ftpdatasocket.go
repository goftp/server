package graval

import (
	"errors"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

// A data socket is used to send non-control data between the client and
// server.
type FTPDataSocket interface {
	Host() string

	Port() int

	// the standard io.Reader interface
	Read(p []byte) (n int, err error)

	// the standard io.Writer interface
	Write(p []byte) (n int, err error)

	// the standard io.Closer interface
	Close() error
}

type FTPActiveSocket struct {
	conn *net.TCPConn
	host string
	port int
}

func NewActiveSocket(host string, port int) (FTPDataSocket, error) {
	connectTo := host + ":" + strconv.Itoa(port)
	log.Print("Opening active data connection to " + connectTo)
	raddr, err := net.ResolveTCPAddr("tcp", connectTo)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	tcpConn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	socket := new(FTPActiveSocket)
	socket.conn = tcpConn
	socket.host = host
	socket.port = port
	return socket, nil
}

func (socket *FTPActiveSocket) Host() string {
	return socket.host
}

func (socket *FTPActiveSocket) Port() int {
	return socket.port
}

func (socket *FTPActiveSocket) Read(p []byte) (n int, err error) {
	return socket.conn.Read(p)
}

func (socket *FTPActiveSocket) Write(p []byte) (n int, err error) {
	return socket.conn.Write(p)
}

func (socket *FTPActiveSocket) Close() error {
	return socket.conn.Close()
}


type FTPPassiveSocket struct {
	conn     *net.TCPConn
	port     int
	ingress  chan []byte
	egress   chan []byte
}

func NewPassiveSocket() (FTPDataSocket, error) {
	socket := new(FTPPassiveSocket)
	socket.ingress = make(chan []byte)
	socket.egress = make(chan []byte)
	go socket.ListenAndServe()
	for {
		if socket.Port() > 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return socket, nil
}

func (socket *FTPPassiveSocket) Host() string {
	return "127.0.0.1"
}

func (socket *FTPPassiveSocket) Port() int {
	return socket.port
}

func (socket *FTPPassiveSocket) Read(p []byte) (n int, err error) {
	if socket.waitForOpenSocket() == false {
		return 0, errors.New("data socket unavailable")
	}
	return socket.conn.Read(p)
}

func (socket *FTPPassiveSocket) Write(p []byte) (n int, err error) {
	if socket.waitForOpenSocket() == false {
		return 0, errors.New("data socket unavailable")
	}
	return socket.conn.Write(p)
}

func (socket *FTPPassiveSocket) Close() error {
	log.Print("closing passive data socket")
	return socket.conn.Close()
}

func (socket *FTPPassiveSocket) ListenAndServe() {
	laddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:0")
	if err != nil {
		log.Print(err)
		return
	}
	listener, err := net.ListenTCP("tcp4", laddr)
	if err != nil {
		log.Print(err)
		return
	}
	add   := listener.Addr()
	parts := strings.Split(add.String(), ":")
	port, err := strconv.Atoi(parts[1])
	if err == nil {
		socket.port = port
	}
	tcpConn, err := listener.AcceptTCP()
	if err != nil {
		log.Print(err)
		return
	}
	socket.conn = tcpConn
}

func (socket *FTPPassiveSocket) waitForOpenSocket() bool {
	retries := 0
	for {
		if socket.conn != nil {
			break
		}
		if retries > 3 {
			return false
		}
		log.Print("sleeping, socket isn't open")
		time.Sleep(500 * time.Millisecond)
		retries += 1
	}
	return true
}

