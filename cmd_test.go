package server

import "testing"

func TestParseListParam(t *testing.T) {
	var paramTests = []struct {
		param    string // input
		expected string // expected result
	}{
		{".", "."},
		{"-la", ""},
		{"-al", ""},
		{"rclone-test-qumelah4himezac1bogajow0", "rclone-test-qumelah4himezac1bogajow0"},
		{"-la rclone-test-qumelah4himezac1bogajow0", "rclone-test-qumelah4himezac1bogajow0"},
		{"-al rclone-test-qumelah4himezac1bogajow0", "rclone-test-qumelah4himezac1bogajow0"},
		{"rclone-test-goximif1kinarez5fakayuw7/new_name/sub_new_name", "rclone-test-goximif1kinarez5fakayuw7/new_name/sub_new_name"},
		{"rclone-test-qumelah4himezac1bogajow0/hello? sausage", "rclone-test-qumelah4himezac1bogajow0/hello? sausage"},
		{"rclone-test-qumelah4himezac1bogajow0/hello? sausage/êé/Hello, 世界/ \" ' @ < > & ? + ≠", "rclone-test-qumelah4himezac1bogajow0/hello? sausage/êé/Hello, 世界/ \" ' @ < > & ? + ≠"},
		{"rclone-test-qumelah4himezac1bogajow0/hello? sausage/êé/Hello, 世界/ \" ' @ < > & ? + ≠/z.txt", "rclone-test-qumelah4himezac1bogajow0/hello? sausage/êé/Hello, 世界/ \" ' @ < > & ? + ≠/z.txt"},
		{"rclone-test-qumelah4himezac1bogajow0/piped data.txt", "rclone-test-qumelah4himezac1bogajow0/piped data.txt"},
		{"rclone-test-qumelah4himezac1bogajow0/not found.txt", "rclone-test-qumelah4himezac1bogajow0/not found.txt"},
	}

	for _, tt := range paramTests {
		path := parseListParam(tt.param)
		if path != tt.expected {
			t.Errorf("parseListParam(%s): expected %s, actual %s", tt.param, tt.expected, path)
		}
	}
}
