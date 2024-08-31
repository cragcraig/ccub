package buildlog

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"
)

func ReadFile(f string) (string, error) {
	fp, err := os.Open(f)
	if err != nil {
		return "", err
	}
	defer fp.Close()
	data, err := ioutil.ReadAll(fp)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func FileExists(f string) (bool, error) {
	if _, err := os.Stat(f); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

func LoadTemplateFromFile(f string) (*template.Template, error) {
	text, err := ReadFile(f)
	if err != nil {
		return nil, err
	}
	tmpl := template.New(f)
	return tmpl.Parse(text)
}

func EnsureDirExists(dir string) error {
	fp, err := os.Open(dir)
	if os.IsNotExist(err) {
		return os.Mkdir(dir, 0755)
	} else if err != nil {
		return err
	}
	defer fp.Close()
	fi, err := fp.Stat()
	if err != nil {
		return err
	}
	if !fi.Mode().IsDir() {
		return fmt.Errorf("File %s already exists but is not a directory", dir)
	}
	return nil
}

func LaunchEditor(f string) error {
	editor := os.Getenv("EDITOR")
	if len(editor) == 0 {
		return errors.New("Environment variable EDITOR is not set")
	}
	path, err := exec.LookPath(editor)
	if err != nil {
		return err
	}
	cmd := exec.Command(path, f)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}
