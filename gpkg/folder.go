package gpkg

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Folder struct {
}

func (s *Folder) Copy(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return s.copy(src, dst, info)
}

func (s *Folder) copy(src, dst string, info os.FileInfo) error {
	if info.IsDir() {
		return s.copyDirectory(src, dst, info)
	}
	return s.copyFile(src, dst, info)
}

func (s *Folder) copyFile(src, dst string, info os.FileInfo) error {

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if err = os.Chmod(dstFile.Name(), info.Mode()); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func (s *Folder) copyDirectory(src, dst string, info os.FileInfo) error {

	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}

	items, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, item := range items {
		err := s.copy(filepath.Join(src, item.Name()), filepath.Join(dst, item.Name()), item)
		if err != nil {
			return err
		}
	}

	return nil
}
