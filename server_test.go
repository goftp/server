package server_test

import (
	"testing"

	filedriver "github.com/goftp/file-driver"
	"github.com/goftp/server"
	"github.com/stretchr/testify/assert"
)

type TestAuth struct {
	Name     string
	Password string
}

func (a *TestAuth) CheckPasswd(name, pass string) (bool, error) {
	if name != a.Name || pass != a.Password {
		return false, nil
	}
	return true, nil
}

func runServer(t *testing.T, execute func()) {
	var perm = server.NewSimplePerm("test", "test")
	opt := &server.ServerOpts{
		Name: "test ftpd",
		Factory: &filedriver.FileDriverFactory{
			"./ftp_test",
			perm,
		},
		Port: ":2121",
		Auth: new(TestAuth),
	}

	server := server.NewServer(opt)
	go func() {
		err := server.ListenAndServe()
		assert.NoError(t, err)
	}

	execute()

	server.Shutdown()
}
