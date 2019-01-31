package fileutils

import (
	"io"
	"os"
	"path/filepath"

	"github.com/moisespsena/go-path-helpers"
)

type WriteMode int

const (
	WO_SetPerm WriteMode = 1 << iota
	WO_SetTimes
	Wo_Sync
	WO_ALL = WO_SetPerm | WO_SetTimes
)

func (f WriteMode) IsSetPerm() bool {
	return (f & WO_SetPerm) != 0
}

func (f WriteMode) IsSetTimes() bool {
	return (f & WO_SetTimes) != 0
}

func (f WriteMode) IsSync() bool {
	return (f & Wo_Sync) != 0
}

// CreateFile writes data to a file named by filename.
// If the file does not exist, CreateFile creates it with permissions perm;
// otherwise CreateFile truncates it before writing.
func CreateFile(filename string, r io.Reader, info os.FileInfo, opt WriteMode) (err error) {
	var f *os.File
	if err = path_helpers.MkdirAllIfNotExists(filepath.Dir(filename)); err != nil {
		return
	}

	if opt.IsSetPerm() {
		f, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	} else {
		f, err = os.Create(filename)
	}
	if err != nil {
		return err
	}
	defer func() {
		if err == nil && opt.IsSync() {
			err = f.Sync()
		}
		if err2 := f.Close(); err != nil && err == nil {
			err = err2
		}
	}()

	if _, err = io.Copy(f, r); err == nil && opt.IsSetTimes() {
		err = os.Chtimes(filename, info.ModTime(), info.ModTime())
	}
	return
}

// CreateFileSync writes data to a file named by filename and set perms and times.
// If the file does not exist, CreateFile creates it with permissions perm;
// otherwise CreateFile truncates it before writing.
func CreateFileSync(filename string, r io.Reader, info os.FileInfo) (err error) {
	return CreateFile(filename, r, info, WO_ALL)
}
