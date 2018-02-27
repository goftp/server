package server_test

import (
	"os"
	"strings"
	"testing"

	filedriver "github.com/goftp/file-driver"
	"github.com/goftp/server"
	"github.com/jlaffaye/ftp"
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
	os.MkdirAll("./testdata", os.ModePerm)

	var perm = server.NewSimplePerm("test", "test")
	opt := &server.ServerOpts{
		Name: "test ftpd",
		Factory: &filedriver.FileDriverFactory{
			RootPath: "./testdata",
			Perm:     perm,
		},
		Port: 2121,
		Auth: &TestAuth{
			Name:     "admin",
			Password: "admin",
		},
	}

	server := server.NewServer(opt)
	go func() {
		err := server.ListenAndServe()
		assert.NoError(t, err)
	}()

	execute()

	assert.NoError(t, server.Shutdown())
}

func TestConnect(t *testing.T) {
	runServer(t, func() {
		ftp, err := ftp.Connect("localhost:2121")
		assert.NoError(t, err)

		assert.NoError(t, ftp.Login("admin", "admin"))
		assert.Error(t, ftp.Login("admin", ""))

		var content = `test`
		assert.NoError(t, ftp.Stor("server_test.go", strings.NewReader(content)))
	})
}
