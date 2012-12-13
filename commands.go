package graval

import (
	"strings"
)

type ftpCommand interface {
	RequireParam() bool
	RequireAuth() bool
	Execute(*FTPConn, string)
}

type commandMap map[string]ftpCommand

var (
	commands = commandMap{
		"ALLO": commandAllo{},
		"CDUP": commandCdup{},
		"CWD":  commandCwd{},
		"DELE": commandDele{},
		"MODE": commandMode{},
		"SYST": commandSyst{},
		"TYPE": commandType{},
		"XCUP": commandCdup{},
		"XCWD": commandCwd{},
	}
)

// commandAllo responds to the ALLO FTP command.
//
// This is essentially a ping from the client so we just respond with an
// basic OK message.
type commandAllo struct{}

func (cmd commandAllo) RequireParam() bool {
	return false
}

func (cmd commandAllo) RequireAuth() bool {
	return false
}

func (cmd commandAllo) Execute(conn *FTPConn, param string) {
	conn.writeMessage(202, "Obsolete")
}

// cmdCdup responds to the CDUP FTP command.
//
// Allows the client change their current directory to the parent.
type commandCdup struct{}

func (cmd commandCdup) RequireParam() bool {
	return false
}

func (cmd commandCdup) RequireAuth() bool {
	return false
}

func (cmd commandCdup) Execute(conn *FTPConn, param string) {
	otherCmd := &commandCwd{}
	otherCmd.Execute(conn, "..")
}

// commandCwd responds to the CWD FTP command. It allows the client to change the
// current working directory.
type commandCwd struct{}

func (cmd commandCwd) RequireParam() bool {
	return false
}

func (cmd commandCwd) RequireAuth() bool {
	return false
}

func (cmd commandCwd) Execute(conn *FTPConn, param string) {
	path := conn.buildPath(param)
	if conn.driver.ChangeDir(path) {
		conn.namePrefix = path
		conn.writeMessage(250, "Directory changed to "+path)
	} else {
		conn.writeMessage(550, "Action not taken")
	}
}

// commandDele responds to the DELE FTP command. It allows the client to delete
// a file
type commandDele struct{}

func (cmd commandDele) RequireParam() bool {
	return false
}

func (cmd commandDele) RequireAuth() bool {
	return false
}

func (cmd commandDele) Execute(conn *FTPConn, param string) {
	path := conn.buildPath(param)
	if conn.driver.DeleteFile(path) {
		conn.writeMessage(250, "File deleted")
	} else {
		conn.writeMessage(550, "Action not taken")
	}
}

// cmdMode responds to the MODE FTP command.
//
// the original FTP spec had various options for hosts to negotiate how data
// would be sent over the data socket, In reality these days (S)tream mode
// is all that is used for the mode - data is just streamed down the data
// socket unchanged.
type commandMode struct{}

func (cmd commandMode) RequireParam() bool {
	return false
}

func (cmd commandMode) RequireAuth() bool {
	return false
}

func (cmd commandMode) Execute(conn *FTPConn, param string) {
	if strings.ToUpper(param) == "S" {
		conn.writeMessage(200, "OK")
	} else {
		conn.writeMessage(504, "MODE is an obsolete command")
	}
}

// commandSyst responds to the SYST FTP command by providing a canned response.
type commandSyst struct{}

func (cmd commandSyst) RequireParam() bool {
	return false
}

func (cmd commandSyst) RequireAuth() bool {
	return false
}

func (cmd commandSyst) Execute(conn *FTPConn, param string) {
	conn.writeMessage(215, "UNIX Type: L8")
}

// commandType responds to the TYPE FTP command.
//
//  like the MODE and STRU commands, TYPE dates back to a time when the FTP
//  protocol was more aware of the content of the files it was transferring, and
//  would sometimes be expected to translate things like EOL markers on the fly.
//
//  Valid options were A(SCII), I(mage), E(BCDIC) or LN (for local type). Since
//  we plan to just accept bytes from the client unchanged, I think Image mode is
//  adequate. The RFC requires we accept ASCII mode however, so accept it, but
//  ignore it.
type commandType struct{}

func (cmd commandType) RequireParam() bool {
	return false
}

func (cmd commandType) RequireAuth() bool {
	return false
}

func (cmd commandType) Execute(conn *FTPConn, param string) {
	if strings.ToUpper(param) == "A" {
		conn.writeMessage(200, "Type set to ASCII")
	} else if strings.ToUpper(param) == "I" {
		conn.writeMessage(200, "Type set to binary")
	} else {
		conn.writeMessage(500, "Invalid type")
	}
}
