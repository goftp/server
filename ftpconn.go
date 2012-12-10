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

type FTPConn struct {
	conn          *net.TCPConn
	controlReader *bufio.Reader
	controlWriter *bufio.Writer
	dataConn      FTPDataSocket
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
	if ftpConn.dataConn != nil {
		ftpConn.dataConn.Close()
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
	case "CDUP", "XCUP":
		ftpConn.cmdCdup()
	case "CWD", "XCWD":
		ftpConn.cmdCwd(param)
	case "DELE":
		ftpConn.cmdDele(param)
	case "LIST":
		ftpConn.cmdList(param)
	case "MKD":
		ftpConn.cmdMkd(param)
	case "MODE":
		ftpConn.cmdMode(param)
	case "NLST":
		ftpConn.cmdNlst(param)
	case "NOOP":
		ftpConn.cmdNoop()
	case "PASS":
		ftpConn.cmdPass(param)
	case "PASV":
		ftpConn.cmdPasv(param)
	case "PORT":
		ftpConn.cmdPort(param)
	case "PWD", "XPWD":
		ftpConn.cmdPwd()
	case "QUIT":
		ftpConn.Close()
	case "RETR":
		ftpConn.cmdRetr(param)
	case "RMD", "XRMD":
		ftpConn.cmdRmd(param)
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
	case "SYST":
		ftpConn.cmdSyst()
	case "TYPE":
		ftpConn.cmdType(param)
	case "USER":
		ftpConn.cmdUser(param)
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
		ftpConn.writeMessage(250, "Directory changed to "+path)
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

// cmdList responds to the LIST FTP command. It allows the client to retreive
// a detailed listing of the contents of a directory.
func (ftpConn *FTPConn) cmdList(param string) {
	ftpConn.writeMessage(150, "Opening ASCII mode data connection for file list")
	path := ftpConn.buildPath(param)
	files := ftpConn.driver.DirContents(path)
	formatter := NewListFormatter(files)
	ftpConn.sendOutofbandData(formatter.Detailed())
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

// cmdNlst responds to the NLST FTP command. It allows the client to retreive
// a list of filenames in the current directory.
func (ftpConn *FTPConn) cmdNlst(param string) {
	ftpConn.writeMessage(150, "Opening ASCII mode data connection for file list")
	path := ftpConn.buildPath(param)
	files := ftpConn.driver.DirContents(path)
	formatter := NewListFormatter(files)
	ftpConn.sendOutofbandData(formatter.Short())
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

// cmdPasv responds to the PASV FTP command.
//
// The client is requesting us to open a new TCP listing socket and wait for them
// to connect to it.
func (ftpConn *FTPConn) cmdPasv(param string) {
	socket, err := NewPassiveSocket()
	if err != nil {
		ftpConn.writeMessage(425, "Data connection failed")
		return
	}
	ftpConn.dataConn = socket
	p1 := socket.Port() / 256
	p2 := socket.Port() - (p1 * 256)

	quads := strings.Split(socket.Host(), ".")
	target := fmt.Sprintf("(%s,%s,%s,%s,%d,%d)", quads[0], quads[1], quads[2], quads[3], p1, p2)
	msg := "Entering Passive Mode "+target
	ftpConn.writeMessage(227, msg)
}

// cmdPort responds to the PORT FTP command.
//
// The client has opened a listening socket for sending out of band data and
// is requesting that we connect to it
func (ftpConn *FTPConn) cmdPort(param string) {
	nums := strings.Split(param, ",")
	portOne, _ := strconv.Atoi(nums[4])
	portTwo, _ := strconv.Atoi(nums[5])
	port := (portOne * 256) + portTwo
	host := nums[0] + "." + nums[1] + "." + nums[2] + "." + nums[3]
	socket, err := NewActiveSocket(host, port)
	if err != nil {
		ftpConn.writeMessage(425, "Data connection failed")
		return
	}
	ftpConn.dataConn = socket
	ftpConn.writeMessage(200, "Connection established ("+strconv.Itoa(port)+")")
}

// cmdPwd responds to the PWD FTP command.
//
// Tells the client what the current working directory is.
func (ftpConn *FTPConn) cmdPwd() {
	ftpConn.writeMessage(257, "\""+ftpConn.namePrefix+"\" is the current directory")
}

// cmdRetr responds to the RETR FTP command. It allows the client to download a
// file.
func (ftpConn *FTPConn) cmdRetr(param string) {
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
func (ftpConn *FTPConn) cmdStor(param string) {
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
func (ftpConn *FTPConn) buildPath(filename string) (fullPath string) {
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
func (ftpConn *FTPConn) sendOutofbandData(data string) {
	bytes := len(data)
	ftpConn.dataConn.Write([]byte(data))
	ftpConn.dataConn.Close()
	message := "Closing data connection, sent " + strconv.Itoa(bytes) + " bytes"
	ftpConn.writeMessage(226, message)
}
