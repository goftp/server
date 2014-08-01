package server

const (
	Read  = 4
	Write = 2
)

type Auth interface {
	AllowAnonymous() bool
	CheckPasswd(string, string) bool
	GetPerms(string, string) int
	HasPerm(string, string, int) bool
}

type AnonymousAuth struct {
}

func (AnonymousAuth) AllowAnonymous() bool {
	return true
}

func (AnonymousAuth) CheckPasswd(string, string) bool {
	return true
}

func (AnonymousAuth) GetPerms(string, string) int {
	return Read + Write
}
func (AnonymousAuth) HasPerm(string, string, int) bool {
	return true
}
