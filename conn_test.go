// Copyright 2018 The goftp Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package server

import (
	"testing"
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
