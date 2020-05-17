package fileutils

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	path_helpers "github.com/moisespsena-go/path-helpers"
	"github.com/pkg/errors"
)

type Destation struct {
	Dest string
}

func (d Destation) Check(dir string) (pth string, err error) {
	p := filepath.Join(dir, d.Dest)
	if dir := filepath.Dir(p); !path_helpers.IsExistingDir(dir) {
		var dirMode os.FileMode
		if dirMode, err = path_helpers.ResolveMode(dir); err != nil {
			return
		}
		err = os.MkdirAll(dir, dirMode)
		if err != nil {
			return "", err
		}
	}
	return pth, nil
}

type Src struct {
	Src  string
	Info *os.FileInfo
	Destation
}

func (s *Src) GetSrc() string {
	return s.Src
}

func (s *Src) CopyTo(dest string) (err error) {
	var pth string
	if pth, err = s.Check(dest); err != nil {
		return
	}
	if err = CopyFile(s.Src, pth); err != nil {
		return errors.Wrap(err, "Copy")
	}
	if s.Info != nil {
		return errors.Wrap(SetInfo(pth, *s.Info), "Set info")
	}
	return
}

type SrcData struct {
	Data []byte
	Info *os.FileInfo
	Destation
}

func (s *SrcData) CopyTo(dest string) (err error) {
	var pth string
	if pth, err = s.Check(dest); err != nil {
		return
	}
	if err = CopyBytes(s.Data, pth); err != nil {
		return errors.Wrap(err, "Copy bytes")
	}
	if s.Info != nil {
		return errors.Wrap(SetInfo(pth, *s.Info), "Set info")
	}
	return
}

type SrcReader struct {
	Info   *os.FileInfo
	Reader io.Reader
	Destation
}

func (s *SrcReader) CopyTo(dest string) (err error) {
	var pth string
	if pth, err = s.Check(dest); err != nil {
		return
	}
	if err = CopyReader(s.Reader, pth); err != nil {
		return errors.Wrap(err, "Copy bytes")
	}
	if s.Info != nil {
		return errors.Wrap(SetInfo(pth, *s.Info), "Set info")
	}
	return
}

type Sourcer interface {
	Copier
	GetSrc() string
}

type Dir struct {
	Src string
	Destation
	Ignore []func(pth string) bool
}

func (d *Dir) GetSrc() string {
	return d.Src
}

func (s *Dir) CopyTo(dest string) (err error) {
	var dirMode os.FileMode
	if dirMode, err = path_helpers.ResolveMode(s.Src); err != nil {
		return
	}
	src := strings.TrimSuffix(s.Src, string(os.PathSeparator)) + string(os.PathSeparator)

	accept := func(pth string) bool {
		for _, f := range s.Ignore {
			if f(pth) {
				return false
			}
		}
		return true
	}

	return filepath.Walk(s.Src, func(path string, info os.FileInfo, err error) error {
		if err == nil {
			var relPath = strings.TrimPrefix(src, s.Src)
			if !accept(relPath) {
				return nil
			}
			if s.Dest != "" {
				relPath = filepath.Join(s.Dest, relPath)
			}
			dst := filepath.Join(dest, relPath)
			if info.IsDir() {
				if err = os.MkdirAll(dst, dirMode); err != nil {
					return errors.Wrapf(err, "MkdirAll %q", dst)
				}
			} else if info.Mode().IsRegular() {
				return CopyFile(path, dst)
			}
		}
		return err
	})
}

type Copier interface {
	CopyTo(dir string) (err error)
}

func CopyTree(dest string, sources []Copier) (err error) {
	if !path_helpers.IsExistingDir(dest) {
		err = os.MkdirAll(dest, os.ModePerm)
		if err != nil {
			return err
		}
	}

	for i, c := range sources {
		if err = c.CopyTo(dest); err != nil {
			return errors.Wrapf(err, "source #%d", i)
		}
	}
	return
}
