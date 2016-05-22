package server

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	defaultWelcomeMessage = "Welcome to the Go FTP Server"
)

type Conn struct {
	conn          net.Conn
	controlReader *bufio.Reader
	controlWriter *bufio.Writer
	dataConn      DataSocket
	driver        Driver
	auth          Auth
	logger        *Logger
	server        *Server
	tlsConfig     *tls.Config
	sessionId     string
	namePrefix    string
	reqUser       string
	user          string
	renameFrom    string
	lastFilePos   int64
	appendData    bool
	closed        bool
	tls           bool
}

func (conn *Conn) LoginUser() string {
	return conn.user
}

func (conn *Conn) IsLogin() bool {
	return len(conn.user) > 0
}

// returns a random 20 char string that can be used as a unique session ID
func newSessionId() string {
	hash := sha256.New()
	_, err := io.CopyN(hash, rand.Reader, 50)
	if err != nil {
		return "????????????????????"
	}
	md := hash.Sum(nil)
	mdStr := hex.EncodeToString(md)
	return mdStr[0:20]
}

// Serve starts an endless loop that reads FTP commands from the client and
// responds appropriately. terminated is a channel that will receive a true
// message when the connection closes. This loop will be running inside a
// goroutine, so use this channel to be notified when the connection can be
// cleaned up.
func (conn *Conn) Serve() {
	conn.logger.Print("Connection Established")
	// send welcome
	conn.writeMessage(220, conn.server.WelcomeMessage)
	// read commands
	for {
		line, err := conn.controlReader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				continue
			}

			conn.logger.Print(fmt.Sprintln("read error:", err))
			break
		}
		conn.receiveLine(line)
		// QUIT command closes connection, break to avoid error on reading from
		// closed socket
		if conn.closed == true {
			break
		}
	}
	conn.Close()
	conn.logger.Print("Connection Terminated")
}

// Close will manually close this connection, even if the client isn't ready.
func (conn *Conn) Close() {
	conn.conn.Close()
	conn.closed = true
	if conn.dataConn != nil {
		conn.dataConn.Close()
		conn.dataConn = nil
	}
}

func (Conn *Conn) upgradeToTls() error {
	Conn.logger.Print("Upgrading connectiion to TLS")
	tlsConn := tls.Server(Conn.conn, Conn.tlsConfig)
	err := tlsConn.Handshake()
	if err == nil {
		Conn.conn = tlsConn
		Conn.controlReader = bufio.NewReader(tlsConn)
		Conn.controlWriter = bufio.NewWriter(tlsConn)
		Conn.tls = true
	}
	return err
}

// receiveLine accepts a single line FTP command and co-ordinates an
// appropriate response.
func (Conn *Conn) receiveLine(line string) {
	command, param := Conn.parseLine(line)
	Conn.logger.PrintCommand(command, param)
	cmdObj := commands[strings.ToUpper(command)]
	if cmdObj == nil {
		Conn.writeMessage(500, "Command not found")
		return
	}
	if cmdObj.RequireParam() && param == "" {
		Conn.writeMessage(553, "action aborted, required param missing")
	} else if cmdObj.RequireAuth() && Conn.user == "" {
		Conn.writeMessage(530, "not logged in")
	} else {
		cmdObj.Execute(Conn, param)
	}
}

func (Conn *Conn) parseLine(line string) (string, string) {
	params := strings.SplitN(strings.Trim(line, "\r\n"), " ", 2)
	if len(params) == 1 {
		return params[0], ""
	}
	return params[0], strings.TrimSpace(params[1])
}

// writeMessage will send a standard FTP response back to the client.
func (Conn *Conn) writeMessage(code int, message string) (wrote int, err error) {
	Conn.logger.PrintResponse(code, message)
	line := fmt.Sprintf("%d %s\r\n", code, message)
	wrote, err = Conn.controlWriter.WriteString(line)
	Conn.controlWriter.Flush()
	return
}

// buildPath takes a client supplied path or filename and generates a safe
// absolute path within their account sandbox.
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
func (Conn *Conn) buildPath(filename string) (fullPath string) {
	if len(filename) > 0 && filename[0:1] == "/" {
		fullPath = filepath.Clean(filename)
	} else if len(filename) > 0 && filename != "-a" {
		fullPath = filepath.Clean(Conn.namePrefix + "/" + filename)
	} else {
		fullPath = filepath.Clean(Conn.namePrefix)
	}
	fullPath = strings.Replace(fullPath, "//", "/", -1)
	return
}

// sendOutofbandData will send a string to the client via the currently open
// data socket. Assumes the socket is open and ready to be used.
func (Conn *Conn) sendOutofbandData(data []byte) {
	bytes := len(data)
	if Conn.dataConn != nil {
		Conn.dataConn.Write(data)
		Conn.dataConn.Close()
		Conn.dataConn = nil
	}
	message := "Closing data connection, sent " + strconv.Itoa(bytes) + " bytes"
	Conn.writeMessage(226, message)
}

func (Conn *Conn) sendOutofBandDataWriter(data io.ReadCloser) error {
	Conn.lastFilePos = 0
	bytes, err := io.Copy(Conn.dataConn, data)
	if err != nil {
		Conn.dataConn.Close()
		Conn.dataConn = nil
		return err
	}
	message := "Closing data connection, sent " + strconv.Itoa(int(bytes)) + " bytes"
	Conn.writeMessage(226, message)
	Conn.dataConn.Close()
	Conn.dataConn = nil

	return nil
}
