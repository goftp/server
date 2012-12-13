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
	case "PASS":
		ftpConn.cmdPass(param)
	case "QUIT":
		ftpConn.Close()
	case "RETR":
		ftpConn.cmdRetr(param)
	case "RNFR":
		ftpConn.cmdRnfr(param)
	case "RNTO":
		ftpConn.cmdRnto(param)
	case "SIZE":
		ftpConn.cmdSize(param)
	case "STOR":
		ftpConn.cmdStor(param)
	case "STRU":
		ftpConn.cmdStru(param)
	case "USER":
		ftpConn.cmdUser(param)
	default:
		ftpConn.writeMessage(500, "Command not found")
	}
}

// cmdPass respond to the PASS FTP command by asking the driver if the supplied
// username and password are valid
func (ftpConn *ftpConn) cmdPass(param string) {
	if ftpConn.driver.Authenticate(ftpConn.reqUser, param) {
		ftpConn.user = ftpConn.reqUser
		ftpConn.reqUser = ""
		ftpConn.writeMessage(230, "Password ok, continue")
	} else {
		ftpConn.writeMessage(530, "Incorrect password, not logged in")
	}
}

// cmdRetr responds to the RETR FTP command. It allows the client to download a
// file.
func (ftpConn *ftpConn) cmdRetr(param string) {
	path := ftpConn.buildPath(param)
	data, err := ftpConn.driver.GetFile(path)
	if err == nil {
		bytes := strconv.Itoa(len(data))
		ftpConn.writeMessage(150, "Data transfer starting "+bytes+" bytes")
		ftpConn.sendOutofbandData(data)
	} else {
		ftpConn.writeMessage(551, "File not available")
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

// cmdSize responds to the SIZE FTP command. It returns the size of the
// requested path in bytes.
func (ftpConn *ftpConn) cmdSize(param string) {
	path := ftpConn.buildPath(param)
	bytes := ftpConn.driver.Bytes(path)
	if bytes >= 0 {
		ftpConn.writeMessage(213, strconv.Itoa(bytes))
	} else {
		ftpConn.writeMessage(450, "file not available")
	}
}

// cmdStor responds to the STOR FTP command. It allows the user to upload a new
// file.
func (ftpConn *ftpConn) cmdStor(param string) {
	targetPath := ftpConn.buildPath(param)
	ftpConn.writeMessage(150, "Data transfer starting")
	tmpFile, err := ioutil.TempFile("", "stor")
	if err != nil {
		ftpConn.writeMessage(450, "error during transfer")
		return
	}
	bytes, err := io.Copy(tmpFile, ftpConn.dataConn)
	if err != nil {
		ftpConn.writeMessage(450, "error during transfer")
		return
	}
	tmpFile.Seek(0,0)
	uploadSuccess := ftpConn.driver.PutFile(targetPath, tmpFile)
	tmpFile.Close()
	os.Remove(tmpFile.Name())
	if uploadSuccess {
		msg := "OK, received "+strconv.Itoa(int(bytes))+" bytes"
		ftpConn.writeMessage(226, msg)
	} else {
		ftpConn.writeMessage(550, "Action not taken")
	}
}

// cmdStru responds to the STRU FTP command.
//
// like the MODE and TYPE commands, stru[cture] dates back to a time when the
// FTP protocol was more aware of the content of the files it was transferring,
// and would sometimes be expected to translate things like EOL markers on the
// fly.
//
// These days files are sent unmodified, and F(ile) mode is the only one we
// really need to support.
func (ftpConn *ftpConn) cmdStru(param string) {
	if strings.ToUpper(param) == "F" {
		ftpConn.writeMessage(200, "OK")
	} else {
		ftpConn.writeMessage(504, "STRU is an obsolete command")
	}
}

// cmdUser responds to the USER FTP command by asking for the password
func (ftpConn *ftpConn) cmdUser(param string) {
	ftpConn.reqUser = param
	ftpConn.writeMessage(331, "User name ok, password required")
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
