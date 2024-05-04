package lib

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	vlib "github.com/mcuadros/go-version"
)

type KiCadInterface struct {
	binPath string
}

func NewKicadInterface() (*KiCadInterface, error) {
	rootDir := filepath.Join(GetProgramFiles(), "KiCad")

	versions, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}

	if len(versions) == 0 {
		return nil, errors.New("no KiCad versions found in program folder")
	}

	latestVersion := "0.0.1"
	for _, e := range versions {
		version := e.Name()
		if vlib.CompareSimple(latestVersion, version) == -1 {
			latestVersion = version
		}
	}

	binPath := filepath.Join(
		GetProgramFiles(), "KiCad", latestVersion, "bin",
	)

	if _, err := os.Stat(filepath.Join(binPath, "kicad-cli.exe")); err != nil {
		return nil, errors.New("KiCad binPath does not exist or does not have kicad-cli")
	}

	return &KiCadInterface{binPath}, nil
}

func (ki *KiCadInterface) GetBinPath() string {
	return ki.binPath
}

func (ki *KiCadInterface) ExecuteCommand(args []string, cwd string) {

	cmd := exec.Command(
		filepath.Join(ki.binPath, "kicad-cli"), args...,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = cwd
	cmd.Run()
}
