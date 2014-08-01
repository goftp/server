package server

type Auth interface {
	AllowAnonymous() bool
	CheckPasswd(string, string) bool
	HasPerm(string, string, int) bool
}
