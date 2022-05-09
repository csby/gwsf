package gtype

import (
	"crypto/md5"
	"fmt"
	"io"
)

func ToMd5(a ...interface{}) string {
	v := fmt.Sprint(a...)
	h := md5.New()

	_, err := io.WriteString(h, v)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}
