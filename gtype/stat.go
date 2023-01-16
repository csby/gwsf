package gtype

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

type StatCodeLine struct {
	sync.Mutex

	SrcRootPath    string
	FileSuffixName string
	ExcludeDirs    []string
	FileCount      int
	LineCount      int
	CommentCount   int
	EmptyCount     int
}

func (s *StatCodeLine) Stat() {
	fmt.Println("统计目录:", s.SrcRootPath)
	fmt.Println("======================================")

	done := make(chan bool)
	go s.codeLineSum(s.SrcRootPath, done)
	<-done

	if s.FileCount > 0 {
		fmt.Println("======================================")
	}

	fmt.Println("统计完成")
	fmt.Println("文件数:", s.FileCount)
	fmt.Println("代码行:", s.LineCount)
	fmt.Println("--------------------------------------")
	fmt.Println("注释行:", s.CommentCount)
	fmt.Println("空白行:", s.EmptyCount)
	fmt.Println("有效行:", s.LineCount-s.EmptyCount-s.CommentCount)
	fmt.Println("--------------------------------------")
}

func (s *StatCodeLine) codeLineSum(root string, done chan bool) {
	var goes int
	goDone := make(chan bool)
	isDstDir := s.checkDir(root)
	defer func() {
		if pan := recover(); pan != nil {
			fmt.Printf("root: %s, panic:%#v\n", root, pan)
		}

		for i := 0; i < goes; i++ {
			<-goDone
		}

		done <- true
	}()
	if !isDstDir {
		return
	}

	rootInfo, err := os.Lstat(root)
	s.checkErr(err)

	rootDir, err := os.Open(root)
	s.checkErr(err)
	defer rootDir.Close()

	if rootInfo.IsDir() {
		fis, err := rootDir.Readdir(0)
		s.checkErr(err)
		for _, fi := range fis {
			if strings.HasPrefix(fi.Name(), ".") {
				continue
			}
			goes++
			if fi.IsDir() {
				go s.codeLineSum(root+"/"+fi.Name(), goDone)
			} else {
				go s.readFile(root+"/"+fi.Name(), goDone)
			}
		}
	} else {
		goes = 1
		go s.readFile(root, goDone)
	}
}

func (s *StatCodeLine) readFile(fileName string, done chan bool) {
	var line int
	var comment int
	var empty int
	isDstFile := strings.HasSuffix(fileName, s.FileSuffixName)
	defer func() {
		if pan := recover(); pan != nil {
			fmt.Printf("filename: %s, panic:%#v\n", fileName, pan)
		}
		if isDstFile {
			s.addLineNum(line, comment, empty)
			fmt.Printf("file %s complete, 总行数 = %d, 注释行 = %d, 空白行 = %d\n", fileName, line, comment, empty)
		}
		done <- true
	}()
	if !isDstFile {
		return
	}

	file, err := os.Open(fileName)
	s.checkErr(err)
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		lineData, isPrefix, err := reader.ReadLine()
		if err != nil {
			break
		}
		if !isPrefix {
			line++

			if lineData == nil {
				empty++
			} else {
				lineText := strings.TrimSpace((string)(lineData))
				lineTextLen := len(lineText)
				if lineTextLen == 0 {
					empty++
				} else if lineTextLen > 1 {
					if lineText[0] == '/' && lineText[1] == '/' {
						comment++
					}
				}
			}
		}
	}
}

func (s *StatCodeLine) checkDir(dirPath string) bool {
	if s.ExcludeDirs == nil {
		return true
	}

	for _, dir := range s.ExcludeDirs {
		if s.SrcRootPath+dir == dirPath {
			return false
		}
	}

	return true
}

func (s *StatCodeLine) addLineNum(total, comment, empty int) {
	s.Lock()
	defer s.Unlock()

	s.LineCount += total
	s.CommentCount += comment
	s.EmptyCount += empty
	s.FileCount += 1

}

func (s *StatCodeLine) checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}
