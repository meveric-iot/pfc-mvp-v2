package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func EstablishDir(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(dir, 0777) // dir does not exist
		} else {
			return err
		}
	} else {
		return nil // exist
	}
}

func GetTargetPathByDate() (string, error) {
	current := time.Now()
	dir := fmt.Sprintf("%02d.%02d.%04d", current.Day(), current.Month(), current.Year())
	dir = "out" + string(filepath.Separator) + dir
	err := EstablishDir(dir)
	if err == nil {
		dir += string(filepath.Separator)
	} else {
		dir = ""
	}
	return dir, err
}

func GenerateStringByCurrentTime() string {
	current := time.Now()
	name := fmt.Sprintf("%02d-%02d-%02d.%04d", current.Hour(), current.Minute(), current.Second(), current.Nanosecond()/100000)
	return name
}

func GenerateTimestamp() string {
	current := time.Now()
	name := fmt.Sprintf("[%02d:%02d:%02d]", current.Hour(), current.Minute(), current.Second())
	return name
}

func AppendLineToLog(filename, str string) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if str != "" {
		if len(str) > 2 {
			if str[len(str)-2:] != "\r\n" {
				str += "\r\n"
			}
		}
		_, err = f.WriteString(str)
	}
	return err
}

func BoolToString(b bool) string {
	if b == true {
		return "true"
	}
	return "false"
}
