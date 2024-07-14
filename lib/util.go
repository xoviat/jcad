package lib

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/lxn/win"
)

func PrintHeader() {
	fmt.Println("JCAD: A JLCPCB PCB Generator for KiCad")
	fmt.Println("Copyright (c) 2023 Mars Galactic")
}

func Exists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	}

	return true
}

func FromCID(cid string) int64 {
	id, _ := strconv.ParseInt(strings.TrimPrefix(cid, "C"), 10, 64)

	return id
}

func NormalizePCB(src string) (string, error) {
	src, err := Normalize(src)
	if err != nil {
		return "", fmt.Errorf("failed to obtain pcb: %s\n", err)
	}

	if !strings.HasSuffix(src, ".kicad_pcb") {
		return "", fmt.Errorf("file is not a pcb: %s\n", err)
	}

	return src, nil
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
		"Ω", "Ohms", "Ohm", "F", "H", "f", "h",
	} {
		val = strings.TrimSuffix(val, suffix)
	}

	// todo: normalize 2k2, 1200, 0.01u, 0.01n

	return val
}

/*
return whether a basic part has an abnormal comment
*/
func IsAbnormal(val string) bool {
	return strings.HasSuffix(val, "K")
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
