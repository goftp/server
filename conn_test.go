// Copyright 2018 The goftp Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package server

import (
	"net"
	"testing"
	"time"
)

func TestConnBuildPath(t *testing.T) {
	c := &Conn{
		namePrefix: "",
	}
	var pathtests = []struct {
		in  string
		out string
	}{
		{"/", "/"},
		{"one.txt", "/one.txt"},
		{"/files/two.txt", "/files/two.txt"},
		{"files/two.txt", "/files/two.txt"},
		{"/../../../../etc/passwd", "/etc/passwd"},
		{"rclone-test-roxarey8facabob5tuwetet4/hello? sausage/êé/Hello, 世界/ \" ' @ < > & ? + ≠/z.txt", "/rclone-test-roxarey8facabob5tuwetet4/hello? sausage/êé/Hello, 世界/ \" ' @ < > & ? + ≠/z.txt"},
	}
	for _, tt := range pathtests {
		t.Run(tt.in, func(t *testing.T) {
			s := c.buildPath(tt.in)
			if s != tt.out {
				t.Errorf("got %q, want %q", s, tt.out)
			}
		})
	}
}

type mockConn struct {
	ip   net.IP
	port int
}

func (m mockConn) Read(b []byte) (n int, err error) {
	return 0, nil
}
func (m mockConn) Write(b []byte) (n int, err error) {
	return 0, nil
}
func (m mockConn) Close() error {
	return nil
}
func (m mockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{
		IP:   m.ip,
		Port: m.port,
	}
}
func (m mockConn) RemoteAddr() net.Addr {
	return nil
}
func (m mockConn) SetDeadline(t time.Time) error {
	return nil
}
func (m mockConn) SetReadDeadline(t time.Time) error {
	return nil
}
func (m mockConn) SetWriteDeadline(t time.Time) error {
	return nil
}
func TestPassiveListenIP(t *testing.T) {
	c := &Conn{
		server: &Server{
			ServerOpts: &ServerOpts{
				PublicIp: "1.1.1.1",
			},
		},
	}
	if c.passiveListenIP() != "1.1.1.1" {
		t.Fatalf("Expected passive listen IP to be 1.1.1.1 but got %s", c.passiveListenIP())
	}

	c = &Conn{
		conn: mockConn{
			ip: net.IPv4(1, 1, 1, 1),
		},
		server: &Server{
			ServerOpts: &ServerOpts{},
		},
	}
	if c.passiveListenIP() != "1.1.1.1" {
		t.Fatalf("Expected passive listen IP to be 1.1.1.1 but got %s", c.passiveListenIP())
	}
}
