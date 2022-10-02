package lib

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func Normalize(src string) (string, error) {
	validate := func(src string) (string, error) {
		if !Exists(src) {
			return "", fmt.Errorf("failed to stat file: %s", src)
		}

		return src, nil
	}

	if !strings.HasPrefix(src, "~") {
		return validate(src)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to obtain user home dir: %s", err)
	}

	return validate(home + strings.TrimPrefix(src, "~"))
}

/*
return a normalized resistor, capacitor, or inductor value

- 10k -> 10k
- 10kOhms -> 10k
- 10uF -> 10u
- 10uH -> 10u
- 1200 -> 1.2k
- 0.01u -> 10n
- 0.01n -> 10p
*/
func NormalizeValue(val string) string {
	for _, suffix := range []string{
		"Ohms", "Ohm", "F", "H", "f", "h",
	} {
		val = strings.TrimSuffix(val, suffix)
	}

	// todo: normalize 2k2, 1200, 0.01u, 0.01n

	return val
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
	programFiles := GetProgramFiles()
	bins := []string{
		filepath.Join(programFiles, "KiCad", "bin"),
		filepath.Join(programFiles, "KiCad", "6.0", "bin"),
	}

	bin := ""
	for _, b := range bins {
		if _, err := os.Stat(b); err != nil {
			continue
		}

		bin = b
		break
	}

	return filepath.Join(bin, "python.exe")
}

func FindScripts() string {
	return "python"
}

func ExecuteScript(script string, args []string) {
	args = append([]string{filepath.Join(FindScripts(), script)}, args...)
	command := exec.Command(FindkPython(), args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	// fmt.Println(FindkPython())
	// fmt.Println(args)

	command.Run()
}
