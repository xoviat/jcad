package lib

import (
	"os"
	"os/exec"
	"path/filepath"
)

func ExecuteScript(script string, args []string) {
	KiCADPython := "C:\\Program Files\\KiCad\\bin\\python.exe"
	scripts := "..\\python"

	args = append([]string{filepath.Join(scripts, script)}, args...)
	command := exec.Command(KiCADPython, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	command.Run()
}
