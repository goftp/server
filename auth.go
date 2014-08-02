package server

const (
	Read  = 4
	Write = 2
)

type Auth interface {
	AllowAnonymous(bool)
	DefaultPerm(int)
	CheckPasswd(string, string) bool
	GetPerms(string, string) int
}

type AnonymousAuth struct {
}

func (AnonymousAuth) AllowAnonymous(bool) {
}

func (AnonymousAuth) DefaultPerm(perm int) {
}

func (AnonymousAuth) CheckPasswd(string, string) bool {
	return true
}

func (AnonymousAuth) GetPerms(string, string) int {
	return Read + Write
}
