package test

import "github.com/csby/gwsf/gcfg"

var cfg = &Config{
	Config: gcfg.Config{
		Svc: gcfg.Svc{
			Name: "gwsf-test",
		},
		Http: gcfg.Http{
			Enabled:     true,
			Port:        8086,
			BehindProxy: false,
		},
		Https: gcfg.Https{
			Enabled:     true,
			Port:        8446,
			BehindProxy: false,
			Cert: gcfg.Crt{
				Ca: gcfg.CrtCa{
					File: caCrtFilePath,
				},
				Server: gcfg.CrtPfx{
					File:     serverCrtFilePath,
					Password: serverCrtPassword,
				},
			},
			RequestClientCert: true,
		},
		Site: gcfg.Site{
			Doc: gcfg.SiteDoc{
				Enabled: true,
			},
			Apps: []gcfg.SiteApp{
				{
					Name: "App1",
					Uri:  "/app1",
				},
				{
					Name: "App2",
					Uri:  "/app2",
				},
			},
		},
	},
}

type Config struct {
	gcfg.Config
}
