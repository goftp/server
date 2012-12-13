package graval

type FTPCommand interface {
	RequireParam() bool
	RequireAuth() bool
	Execute(*FTPConn, string)
}

type commandMap map[string]FTPCommand

var (
	commands = commandMap{
		"ALLO": CommandAllo{},
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
