package gotail

import (
	"os"
	"syscall"
)

func openFile(path string) (*os.File, error) {

	if len(path) == 0 {
		return nil, syscall.ERROR_FILE_NOT_FOUND
	}
	pathp, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}

	access := uint32(syscall.GENERIC_READ)
	sharemode := uint32(syscall.FILE_SHARE_READ | syscall.FILE_SHARE_WRITE | syscall.FILE_SHARE_DELETE)
	var sa *syscall.SecurityAttributes
	createmode := uint32(syscall.OPEN_EXISTING)

	h, e := syscall.CreateFile(pathp, access, sharemode, sa, createmode, syscall.FILE_ATTRIBUTE_NORMAL, 0)
	if e != nil {
		return nil, e
	}

	return os.NewFile(uintptr(h), path), nil

}
