package test

import (
	"github.com/csby/gsecurity/gcrt"
	"github.com/csby/gsecurity/grsa"
	"os"
	"path/filepath"
	"runtime"
)

var (
	caCrtFilePath     = getCrtFilePath("ca.crt")
	serverCrtFilePath = getCrtFilePath("server.pfx")
	serverCrtPassword = "server123"
	serverCrtOU       = "server-test"
	clientCrtFilePath = getCrtFilePath("client.pfx")
	clientCrtPassword = "client123"
	clientCrtOU       = "client-test"
)

func createCrts() error {
	// ca
	caPrivate := &grsa.Private{}
	err := caPrivate.Create(2048)
	if err != nil {
		return err
	}
	caPublic, err := caPrivate.Public()
	if err != nil {
		return err
	}
	caTemplate := &gcrt.Template{
		Organization:       "ca",
		OrganizationalUnit: "ca-test",
	}
	template, err := caTemplate.Template()
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	caCrt := &gcrt.Crt{}
	err = caCrt.Create(template, template, caPublic, caPrivate)
	if err != nil {
		return err
	}
	err = caCrt.ToFile(caCrtFilePath)
	if err != nil {
		return err
	}

	// server
	serverPrivate := &grsa.Private{}
	err = serverPrivate.Create(2048)
	if err != nil {
		return err
	}
	serverPublic, err := serverPrivate.Public()
	if err != nil {
		return err
	}
	serverTemplate := &gcrt.Template{
		Organization:       "server",
		OrganizationalUnit: serverCrtOU,
		Hosts: []string{
			"127.0.0.1",
			"localhost",
		},
	}
	template, err = serverTemplate.Template()
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	serverCrt := &gcrt.Pfx{}
	err = serverCrt.Create(template, caCrt.Certificate(), serverPublic, caPrivate)
	if err != nil {
		return err
	}
	err = serverCrt.ToFile(serverCrtFilePath, caCrt, serverPrivate, serverCrtPassword)
	if err != nil {
		return err
	}

	// client
	clientPrivate := &grsa.Private{}
	err = clientPrivate.Create(2048)
	if err != nil {
		return err
	}
	clientPublic, err := clientPrivate.Public()
	if err != nil {
		return err
	}
	clientTemplate := &gcrt.Template{
		Organization:       "client",
		OrganizationalUnit: clientCrtOU,
	}
	template, err = clientTemplate.Template()
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	clientCrt := &gcrt.Pfx{}
	err = clientCrt.Create(template, caCrt.Certificate(), clientPublic, caPrivate)
	if err != nil {
		return err
	}
	err = clientCrt.ToFile(clientCrtFilePath, caCrt, clientPrivate, clientCrtPassword)
	if err != nil {
		return err
	}

	return nil
}

func deleteCrts() error {
	folder := crtFileFolder()
	return os.RemoveAll(folder)
}

func getCrtFilePath(name string) string {
	return filepath.Join(crtFileFolder(), name)
}

func crtFileFolder() string {
	_, file, _, _ := runtime.Caller(0)

	return filepath.Join(filepath.Dir(file), "crts")
}
