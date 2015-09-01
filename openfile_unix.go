// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package gotail

import "os"

func openFile(path string) (*os.File, error) {
	return os.Open(path)
}
