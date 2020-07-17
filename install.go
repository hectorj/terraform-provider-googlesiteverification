package main

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
)

func install() {
	isHomeInstallation := len(os.Args) > 2 && (os.Args[2] == "home")

	var destinationBaseDir string
	if !isHomeInstallation {
		destinationBaseDir = filepath.Join(".", ".terraform")
	} else {
		if runtime.GOOS == "windows" {
			appDataDir, appDataExists := os.LookupEnv("APPDATA")
			if !appDataExists {
				panic("APPDATA env var is not set")
			}
			destinationBaseDir = filepath.Join(appDataDir, "terraform.d")
		} else {
			homeDir, homeErr := os.UserHomeDir()
			if homeErr != nil {
				panic(homeErr)
			}
			destinationBaseDir = filepath.Join(homeDir, ".terraform.d")
		}
	}
	destinationPath := filepath.Join(destinationBaseDir, "plugins", runtime.GOOS+"_"+runtime.GOARCH, "terraform-provider-googlesiteverification")

	pluginSrc, openErr := os.Open(os.Args[0])
	if openErr != nil {
		panic(openErr)
	}

	if mkdirErr := os.MkdirAll(filepath.Dir(destinationPath), 0755); mkdirErr != nil {
		panic(mkdirErr)
	}

	pluginDest, createErr := os.Create(destinationPath)
	if createErr != nil {
		panic(createErr)
	}

	if _, copyErr := io.Copy(pluginDest, pluginSrc); copyErr != nil {
		panic(copyErr)
	}

	_ = pluginSrc.Close()
	if closeErr := pluginDest.Close(); closeErr != nil {
		panic(closeErr)
	}
}
