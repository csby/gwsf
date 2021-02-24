package gpkg

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func (s *packer) createFile(fileWriter io.Writer, folder string, files ...string) error {
	gw := gzip.NewWriter(fileWriter)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, item := range files {
		fi, err := os.Stat(item)
		if err == nil || os.IsExist(err) {
			fr, err := os.Open(item)
			if err != nil {
				return err
			}
			defer fr.Close()

			h := new(tar.Header)
			if folder != "" {
				h.Name = fmt.Sprintf("%s/%s", folder, fi.Name())
			} else {
				h.Name = fi.Name()
			}
			h.Size = fi.Size()
			h.Mode = int64(fi.Mode())
			h.ModTime = fi.ModTime()
			fmt.Print("	=> ", h.Name)
			err = tw.WriteHeader(h)
			if err != nil {
				fmt.Println(",错误:", err)
				return err
			}
			_, err = io.Copy(tw, fr)
			if err != nil {
				fmt.Println(",错误:", err)
				return err
			}
			fmt.Println(",成功")
			fr.Close()
		} else {
			fmt.Println("警告:", item, "不存在")
		}
	}

	return nil
}

func (s *packer) compressFolder(fileWriter io.Writer, folderPath, folderName string, ignore func(name string) bool) error {
	gw := gzip.NewWriter(fileWriter)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	return s.createSubFolder(tw, folderPath, folderName, ignore)
}

func (s *packer) createSubFolder(tw *tar.Writer, folderPath, folderName string, ignore func(name string) bool) error {
	paths, err := ioutil.ReadDir(folderPath)
	if err != nil {
		return err
	}
	for _, path := range paths {
		if ignore != nil {
			if ignore(path.Name()) {
				continue
			}
		}

		fp := filepath.Join(folderPath, path.Name())
		if path.IsDir() {
			subFolderName := path.Name()
			if folderName != "" {
				subFolderName = fmt.Sprintf("%s/%s", folderName, path.Name())
			}
			err = s.createSubFolder(tw, fp, subFolderName, nil)
			if err != nil {
				return err
			}
		} else {
			fi, err := os.Stat(fp)
			if err != nil {
				return err
			}

			fr, err := os.Open(fp)
			if err != nil {
				return err
			}
			defer fr.Close()

			fn := fi.Name()
			if folderName != "" {
				fn = fmt.Sprintf("%s/%s", folderName, fi.Name())
			}
			fmt.Print("	=> ", fn)

			fh := new(tar.Header)
			fh.Name = fn
			fh.Size = fi.Size()
			fh.Mode = int64(fi.Mode())
			fh.ModTime = fi.ModTime()
			err = tw.WriteHeader(fh)
			if err != nil {
				fmt.Println(",错误:", err)
				return err
			}
			_, err = io.Copy(tw, fr)
			if err != nil {
				fmt.Println(",错误:", err)
				return err
			}
			tw.Flush()

			fmt.Println(",成功")
			fr.Close()
		}
	}

	return nil
}

func (s *packer) pkgExt() string {
	return "tar.gz"
}
