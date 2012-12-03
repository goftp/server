package graval

type FTPDriverFactory interface {
	NewDriver() FTPDriver
}

type FTPDriver interface {
	Authenticate(string, string) bool
}
