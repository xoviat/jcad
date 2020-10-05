package lib

import (
	"bytes"
	"encoding/gob"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/lxn/win"
)

func Exists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	}

	return true
}

/*
	return an encoded object as bytes
*/
func Marshal(v interface{}) ([]byte, error) {
	b := new(bytes.Buffer)
	err := gob.NewEncoder(b).Encode(v)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

/*
	return a decoded object from bytes
*/
func Unmarshal(data []byte, v interface{}) error {
	b := bytes.NewBuffer(data)
	return gob.NewDecoder(b).Decode(v)
}

func GetProgramFiles() string {
	buf := make([]uint16, win.MAX_PATH)
	win.SHGetSpecialFolderPath(win.HWND(0), &buf[0], win.CSIDL_PROGRAM_FILES, false)

	return syscall.UTF16ToString(buf)
}

func GetLocalAppData() string {
	buf := make([]uint16, win.MAX_PATH)
	win.SHGetSpecialFolderPath(win.HWND(0), &buf[0], win.CSIDL_LOCAL_APPDATA, false)

	return syscall.UTF16ToString(buf)
}

func FindkPython() string {
	return filepath.Join(GetProgramFiles(), "KiCad", "bin", "python.exe")
}

func FindScripts() string {
	return "python"
}

func ExecuteScript(script string, args []string) {
	args = append([]string{filepath.Join(FindScripts(), script)}, args...)
	command := exec.Command(FindkPython(), args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	command.Run()
}

func bcKey(component *BoardComponent) []byte {
	key, _ := Marshal([]string{
		re1.ReplaceAllString(component.Designator, ""),
		component.Comment,
		component.Footprint,
	})

	return key
}

func BcKey(component *BoardComponent) []byte {
	return bcKey(component)
}
