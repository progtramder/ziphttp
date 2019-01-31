package ziphttp

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var strJsonContent string

type command interface {
	pathBrowser() string
	pathMainJs() string
	pathJson() string
	pathJsForJson() string
}

type cmdWin struct {
}

func (cmd *cmdWin) pathBrowser() string {
	return filepath.Dir(os.Args[0]) + "\\lib\\Browser.exe"
}

func (cmd *cmdWin) pathMainJs() string {
	return filepath.Dir(os.Args[0]) + "\\lib\\main.js"
}

func (cmd *cmdWin) pathJson() string {
	return filepath.Dir(os.Args[0]) + "\\lib\\resources\\app\\package.json"
}

func (cmd *cmdWin) pathJsForJson() string {
	return "../../main.js"
}

type cmdMac struct {
}

func (cmd *cmdMac) pathBrowser() string {
	return filepath.Dir(os.Args[0]) + "/lib/Browser.app/Contents/MacOS/Browser"
}

func (cmd *cmdMac) pathMainJs() string {
	return filepath.Dir(os.Args[0]) + "/lib/Browser.app/Contents/Resources/main.js"
}

func (cmd *cmdMac) pathJson() string {
	return filepath.Dir(os.Args[0]) + "/lib/Browser.app/Contents/Resources/app/package.json"
}

func (cmd *cmdMac) pathJsForJson() string {
	return "../main.js"
}

var osCommands = map[string]command{
	"darwin":  &cmdMac{},
	"windows": &cmdWin{},
}

func osCommand() command {
	if cmd, ok := osCommands[runtime.GOOS]; ok {
		return cmd
	} else {
		panic("Unsupported platform!")
	}
}

func Exec(args ...string) error {
	if err := CreateMainJs(); err != nil {
		return err
	}

	if err := ModifyPacketJson(); err != nil {
		return RemoveMainJs()
	}

	path := osCommand().pathBrowser()
	return exec.Command(path, args...).Run()
}

func CreateMainJs() error {
	path := osCommand().pathMainJs()
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer file.Close()
	_, err = file.WriteString(fmt.Sprintf(mainJS, GUID, startPage))

	return err
}

func RemoveMainJs() error {
	return os.Remove(osCommand().pathMainJs())
}

func ModifyPacketJson() error {
	path := osCommand().pathJson()
	file, err := os.OpenFile(path, os.O_RDWR, 0600|os.ModeExclusive)
	if err != nil {
		return err
	}

	defer file.Close()
	content, _ := ioutil.ReadAll(file)
	//Save the content of json file
	strJsonContent = string(content)
	file.Truncate(0)
	file.Seek(0, 0)
	_, err = file.WriteString(fmt.Sprintf(packageJson, osCommand().pathJsForJson()))
	return err
}

func RecoverPacketJson() error {
	path := osCommand().pathJson()
	file, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC, 0600|os.ModeExclusive)
	if err != nil {
		return err
	}

	defer file.Close()
	_, err = file.WriteString(strJsonContent)
	return err
}
