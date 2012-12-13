package graval

import (
	"strings"
)

type FTPCommand interface {
	RequireParam() bool
	RequireAuth() bool
	Execute(*FTPConn, string)
}

type commandMap map[string]FTPCommand

var (
	commands = commandMap{
		"ALLO": CommandAllo{},
		"SYST": CommandSyst{},
		"TYPE": CommandType{},
	}
)

// CommandAllo responds to the ALLO FTP command.
//
// This is essentially a ping from the client so we just respond with an
// basic OK message.
type CommandAllo struct{}

func (cmd CommandAllo) RequireParam() bool {
	return false
}

func (cmd CommandAllo) RequireAuth() bool {
	return false
}

func (cmd CommandAllo) Execute(conn *FTPConn, param string) {
	conn.writeMessage(202, "Obsolete")
}

// CommandSyst responds to the SYST FTP command by providing a canned response.
type CommandSyst struct{}

func (cmd CommandSyst) RequireParam() bool {
	return false
}

func (cmd CommandSyst) RequireAuth() bool {
	return false
}

func (cmd CommandSyst) Execute(conn *FTPConn, param string) {
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
type CommandType struct{}

func (cmd CommandType) RequireParam() bool {
	return false
}

func (cmd CommandType) RequireAuth() bool {
	return false
}

func (cmd CommandType) Execute(conn *FTPConn, param string) {
	if strings.ToUpper(param) == "A" {
		conn.writeMessage(200, "Type set to ASCII")
	} else if strings.ToUpper(param) == "I" {
		conn.writeMessage(200, "Type set to binary")
	} else {
		conn.writeMessage(500, "Invalid type")
	}
}
