package main

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var ErrWrongFileName = errors.New("name contains forbidden symbol: =")

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	envs := make(Environment)
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		name := info.Name()
		err = checkFileName(name)
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() {
			continue
		}
		needRemove := info.Size() == 0

		v, err := getValue(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		envs[name] = EnvValue{
			Value:      v,
			NeedRemove: needRemove,
		}
	}
	return envs, nil
}

func checkFileName(fileName string) error {
	if strings.Contains(fileName, "=") {
		return ErrWrongFileName
	}
	return nil
}

func getValue(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	line, err := reader.ReadString('\n')

	if err != nil && err.Error() != "EOF" {
		return "", err
	}

	// Удаляем символ новой строки
	if len(line) > 0 {
		runes := []rune(line)
		lastRune := runes[len(runes)-1:]
		if string(lastRune) == "\n" {
			line = line[:len(line)-1]
		}
	}

	trimed := trimRight(line)
	return replaceNull(trimed), nil
}

func trimRight(s string) []byte {
	res := []byte(s)

	for {
		trimSet := []byte(" ")
		withoutSpace := bytes.TrimRight(res, string(trimSet))
		trimSet = []byte("\t")
		res = bytes.TrimRight(withoutSpace, string(trimSet))
		if len(res) == 0 {
			return res
		}
		runes := []rune(string(res))
		lastRune := runes[len(runes)-1:]
		if string(lastRune) != "\t" && string(lastRune) != " " {
			break
		}
	}

	return res
}

func replaceNull(s []byte) string {
	newString := bytes.ReplaceAll(s, []byte{0}, []byte("\n"))
	return string(newString)
}
