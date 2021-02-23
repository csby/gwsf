package test

import (
	"fmt"
	"github.com/csby/gwsf/gtype"
	"testing"
	"time"
)

func Test_gwsf(t *testing.T) {
	fmt.Println("crt folder path:", crtFileFolder())

	err := createCrts()
	if err != nil {
		t.Fatal(err)
	}
	defer deleteCrts()

	server := &Server{}
	err = server.Run(func(server gtype.Server) {
		time.Sleep(time.Second * 5)

		client := &Client{}
		err := client.Run()

		select {
		case <-time.After(time.Second * 30):
		case <-err:
		}
		server.Shutdown()
	})

	if err != nil {
		t.Fatal(err)
	}
}
