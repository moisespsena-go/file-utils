package fileutils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/moisespsena/go-error-wrap"
)

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.
func CopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	if err = CopyFileContents(src, dst); err == nil {
		err = SetInfo(dst, sfi)
	}
	return
}

// CopyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func CopyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	err = CopyReader(in, dst)
	return
}

func CopyBytes(in []byte, dst string) (err error) {
	return CopyReader(bytes.NewBuffer(in), dst)
}

func CopyReader(in io.Reader, dst string) (err error) {
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func CopyReaderInfo(in io.Reader, info os.FileInfo, dst string) (err error) {
	if !info.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", info.Name(), info.Mode().String())
	}

	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(info, dfi) {
			return nil
		}
	}

	if err = errwrap.Wrap(CopyReader(in, dst), "Copy contents"); err == nil {
		err = errwrap.Wrap(SetInfo(dst, info), "Set info")
	}
	return
}

func SetInfo(pth string, info os.FileInfo) (err error) {
	if err = os.Chtimes(pth, time.Now(), info.ModTime()); err == nil {
		err = os.Chmod(pth, info.Mode())
	}
	return
}
