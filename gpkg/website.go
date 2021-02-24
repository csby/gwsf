package gpkg

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Website struct {
	Enable bool   `json:"enable"`
	Name   string `json:"name"`
	Src    Source `json:"src"`
}

func (s *Website) GetVersion() (string, error) {
	/*
		const version = "1.0.1.1"

		export default {
			version
		}
	*/
	folderPath := filepath.Join(s.Src.Root, "src")
	filePath := filepath.Join(folderPath, "version.js")
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	bufReader := bufio.NewReader(file)
	for {
		line, err := bufReader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if len(line) <= 0 {
			continue
		}
		if line[0] == '/' {
			continue
		}

		keyValue := strings.Split(line, "=")
		if len(keyValue) < 2 {
			continue
		}
		if strings.TrimSpace(keyValue[0]) != "const version" {
			continue
		}
		value := strings.TrimSpace(keyValue[1])
		value = strings.TrimLeft(value, "'")
		value = strings.TrimLeft(value, "\"")
		value = strings.TrimRight(value, "'")
		value = strings.TrimRight(value, "\"")

		return value, nil
	}

	return "", nil
}
