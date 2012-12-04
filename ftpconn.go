package graval

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	welcomeMessage = "Welcome to the Go FTP Server"
)

type FTPConn struct {
	conn          *net.TCPConn
	controlReader *bufio.Reader
	controlWriter *bufio.Writer
	data          *net.TCPConn
	driver        FTPDriver
	namePrefix    string
	reqUser       string
	user          string
	renameFrom    string
}

// NewFTPConn constructs a new object that will handle the FTP protocol over
// an active net.TCPConn. The TCP connection should already be open before
// it is handed to this functions. driver is an instance of FTPDrive that
// will handle all auth and persistence details.
func NewFTPConn(tcpConn *net.TCPConn, driver FTPDriver) *FTPConn {
	c := new(FTPConn)
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
func (ftpConn *FTPConn) Serve() {
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
func (ftpConn *FTPConn) Close() {
	ftpConn.conn.Close()
	if ftpConn.data != nil {
		ftpConn.data.Close()
	}
}

// receiveLine accepts a single line FTP command and co-ordinates an
// appropriate response.
func (ftpConn *FTPConn) receiveLine(line string) {
	log.Print(line)
	command, param := ftpConn.parseLine(line)
	switch command {
	case "ALLO":
		ftpConn.cmdAllo()
		break
	case "CDUP", "XCUP":
		ftpConn.cmdCdup()
		break
	case "CWD", "XCWD":
		ftpConn.cmdCwd(param)
		break
	case "DELE":
		ftpConn.cmdDele(param)
		break
	case "MKD":
		ftpConn.cmdMkd(param)
		break
	case "MODE":
		ftpConn.cmdMode(param)
		break
	case "NOOP":
		ftpConn.cmdNoop()
		break
	case "PASS":
		ftpConn.cmdPass(param)
		break
	case "QUIT":
		ftpConn.Close()
		break
	case "RMD", "XRMD":
		ftpConn.cmdRmd(param)
		break
	case "RNFR":
		ftpConn.cmdRnfr(param)
		break
	case "RNTO":
		ftpConn.cmdRnto(param)
		break
	case "SIZE":
		ftpConn.cmdSize(param)
		break
	case "STRU":
		ftpConn.cmdStru(param)
		break
	case "SYST":
		ftpConn.cmdSyst()
		break
	case "TYPE":
		ftpConn.cmdType(param)
		break
	case "USER":
		ftpConn.cmdUser(param)
		break
	default:
		ftpConn.writeMessage(500, "Command not found")
	}
}

// cmdNoop responds to the ALLO FTP command.
//
// This is essentially a ping from the client so we just respond with an
// basic OK message.
func (ftpConn *FTPConn) cmdAllo() {
	ftpConn.writeMessage(202, "Obsolete")
}

// cmdCdup responds to the CDUP FTP command.
//
// Allows the client change their current directory to the parent.
func (ftpConn *FTPConn) cmdCdup() {
	ftpConn.cmdCwd("..")
}

// cmdCwd responds to the CWD FTP command. It allows the client to change the
// current working directory.
func (ftpConn *FTPConn) cmdCwd(param string) {
	path := ftpConn.buildPath(param)
	if ftpConn.driver.ChangeDir(path) {
		ftpConn.namePrefix = path
		ftpConn.writeMessage(250, "Directory changed to " + path)
	} else {
		ftpConn.writeMessage(550, "Action not taken")
	}
}

// cmdDele responds to the DELE FTP command. It allows the client to delete
// a file
func (ftpConn *FTPConn) cmdDele(param string) {
	path := ftpConn.buildPath(param)
	if ftpConn.driver.DeleteFile(path) {
		ftpConn.writeMessage(250, "File deleted")
	} else {
		ftpConn.writeMessage(550, "Action not taken")
	}
}

// cmdMkd responds to the MKD FTP command. It allows the client to create
// a new directory
func (ftpConn *FTPConn) cmdMkd(param string) {
	path := ftpConn.buildPath(param)
	if ftpConn.driver.MakeDir(path) {
		ftpConn.writeMessage(257, "Directory created")
	} else {
		ftpConn.writeMessage(550, "Action not taken")
	}
}

// cmdMode responds to the MODE FTP command.
//
// the original FTP spec had various options for hosts to negotiate how data
// would be sent over the data socket, In reality these days (S)tream mode
// is all that is used for the mode - data is just streamed down the data
// socket unchanged.
func (ftpConn *FTPConn) cmdMode(param string) {
	if strings.ToUpper(param) == "S" {
		ftpConn.writeMessage(200, "OK")
	} else {
		ftpConn.writeMessage(504, "MODE is an obsolete command")
	}
}

// cmdNoop responds to the NOOP FTP command.
//
// This is essentially a ping from the client so we just respond with an
// basic 200 message.
func (ftpConn *FTPConn) cmdNoop() {
	ftpConn.writeMessage(200, "OK")
}

// cmdPass respond to the PASS FTP command by asking the driver if the supplied
// username and password are valid
func (ftpConn *FTPConn) cmdPass(param string) {
	if ftpConn.driver.Authenticate(ftpConn.reqUser, param) {
		ftpConn.user = ftpConn.reqUser
		ftpConn.reqUser = ""
		ftpConn.writeMessage(230, "Password ok, continue")
	} else {
		ftpConn.writeMessage(530, "Incorrect password, not logged in")
	}
}

// cmdRmd responds to the RMD FTP command. It allows the client to delete a
// directory.
func (ftpConn *FTPConn) cmdRmd(param string) {
	path := ftpConn.buildPath(param)
	if ftpConn.driver.DeleteDir(path) {
		ftpConn.writeMessage(250, "Directory deleted")
	} else {
		ftpConn.writeMessage(550, "Action not taken")
	}
}

// cmdRnfr responds to the RNFR FTP command. It's the first of two commands
// required for a client to rename a file.
func (ftpConn *FTPConn) cmdRnfr(param string) {
	ftpConn.renameFrom = ftpConn.buildPath(param)
	ftpConn.writeMessage(350, "Requested file action pending further information.")
}

// cmdRnto responds to the RNTO FTP command. It's the second of two commands
// required for a client to rename a file.
func (ftpConn *FTPConn) cmdRnto(param string) {
	toPath := ftpConn.buildPath(param)
	if ftpConn.driver.Rename(ftpConn.renameFrom, toPath) {
		ftpConn.writeMessage(250, "File renamed")
	} else {
		ftpConn.writeMessage(550, "Action not taken")
	}
}

// cmdSize responds to the SIZE FTP command. It returns the size of the
// requested path in bytes.
func (ftpConn *FTPConn) cmdSize(param string) {
	path  := ftpConn.buildPath(param)
	bytes := ftpConn.driver.Bytes(path)
	if bytes >= 0 {
		ftpConn.writeMessage(213, strconv.Itoa(bytes))
	} else {
		ftpConn.writeMessage(450, "file not available")
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
func (ftpConn *FTPConn) cmdStru(param string) {
	if strings.ToUpper(param) == "F" {
		ftpConn.writeMessage(200, "OK")
	} else {
		ftpConn.writeMessage(504, "STRU is an obsolete command")
	}
}

// cmdSyst responds to the SYST FTP command by providing a canned response.
func (ftpConn *FTPConn) cmdSyst() {
	ftpConn.writeMessage(215, "UNIX Type: L8")
}

// cmdType responds to the TYPE FTP command.
//
//  like the MODE and STRU commands, TYPE dates back to a time when the FTP
//  protocol was more aware of the content of the files it was transferring, and
//  would sometimes be expected to translate things like EOL markers on the fly.
//
//  Valid options were A(SCII), I(mage), E(BCDIC) or LN (for local type). Since
//  we plan to just accept bytes from the client unchanged, I think Image mode is
//  adequate. The RFC requires we accept ASCII mode however, so accept it, but
//  ignore it.
func (ftpConn *FTPConn) cmdType(param string) {
	if strings.ToUpper(param) == "A" {
		ftpConn.writeMessage(200, "Type set to ASCII")
	} else if strings.ToUpper(param) == "I" {
		ftpConn.writeMessage(200, "Type set to binary")
	} else {
		ftpConn.writeMessage(500, "Invalid type")
	}
}

// cmdUser responds to the USER FTP command by asking for the password
func (ftpConn *FTPConn) cmdUser(param string) {
	ftpConn.reqUser = param
	ftpConn.writeMessage(331, "User name ok, password required")
}

func (ftpConn *FTPConn) parseLine(line string) (string, string) {
	params := strings.SplitN(strings.Trim(line, "\r\n"), " ", 2)
	if len(params) == 1 {
		return params[0], ""
	}
	return params[0], params[1]
}

// writeMessage will send a standard FTP response back to the client.
func (ftpConn *FTPConn) writeMessage(code int, message string) (wrote int, err error) {
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
func (ftpConn *FTPConn) buildPath(filename string) (fullPath string){
	if filename[0:1] == "/" {
		fullPath = filepath.Clean(filename)
	} else if filename != "" && filename != "-a" {
		fullPath = filepath.Clean(ftpConn.namePrefix + "/" + filename)
	} else {
		fullPath = filepath.Clean(ftpConn.namePrefix)
	}
	fullPath = strings.Replace(fullPath, "//", "/", -1)
	return
}
