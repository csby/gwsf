package gpkg

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func (s *packer) createFile(fileWriter io.Writer, folder string, files ...string) error {
	zw := zip.NewWriter(fileWriter)
	defer zw.Close()

	for _, item := range files {
		fi, err := os.Stat(item)
		if err == nil || os.IsExist(err) {
			fr, err := os.Open(item)
			if err != nil {
				return err
			}
			defer fr.Close()

			fn := fi.Name()
			if folder != "" {
				fn = fmt.Sprintf("%s/%s", folder, fi.Name())
			}
			fmt.Print("	=> ", fn)
			fh, err := zip.FileInfoHeader(fi)
			if err != nil {
				fmt.Println(",错误:", err)
				return err
			}
			fh.Name = fn
			fh.Method = zip.Deflate
			fw, err := zw.CreateHeader(fh)
			if err != nil {
				fmt.Println(",错误:", err)
				return err
			}
			_, err = io.Copy(fw, fr)
			if err != nil {
				fmt.Println(",错误:", err)
				return err
			}
			zw.Flush()

			fmt.Println(",成功")
			fr.Close()
		} else {
			fmt.Println("警告:", item, "不存在")
		}
	}

	return nil
}

func (s *packer) compressFolder(fileWriter io.Writer, folderPath, folderName string, ignore func(name string) bool) error {
	zw := zip.NewWriter(fileWriter)
	defer zw.Close()

	return s.createSubFolder(zw, folderPath, folderName, ignore)
}

func (s *packer) createSubFolder(zw *zip.Writer, folderPath, folderName string, ignore func(name string) bool) error {
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
			err = s.createSubFolder(zw, fp, subFolderName, nil)
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
			fh, err := zip.FileInfoHeader(fi)
			if err != nil {
				fmt.Println(",错误:", err)
				return err
			}
			fh.Name = fn
			fh.Method = zip.Deflate
			fw, err := zw.CreateHeader(fh)
			if err != nil {
				fmt.Println(",错误:", err)
				return err
			}
			_, err = io.Copy(fw, fr)
			if err != nil {
				fmt.Println(",错误:", err)
				return err
			}
			zw.Flush()

			fmt.Println(",成功")
			fr.Close()
		}
	}

	return nil
}

func (s *packer) pkgExt() string {
	return "zip"
}
