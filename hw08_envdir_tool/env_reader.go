package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// Получить значение для переменной окружения из файла.
// Читаем первую строчку файла, тримим пробельные символы справа, заменяем терминальные нули на переводы строки.
func readEnvValue(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(exitCode)
		}
	}()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	if scanner.Err() != nil {
		return "", err
	}
	result := scanner.Text()
	result = strings.TrimRightFunc(result, unicode.IsSpace)
	result = strings.ReplaceAll(result, string([]byte{0x00}), "\n")

	return result, nil
}

// Обрабатываем элемент из директории, представляющий переменную окружения.
// Возвращаем имя переменной окружения, и структуру, содержащую значение этой переменной. Или ошибку.
func processDirEntry(dir string, entry os.DirEntry) (string, EnvValue, error) {
	path := filepath.Join(dir, entry.Name())
	if entry.IsDir() {
		message := fmt.Sprintf("%s is directory", path)
		return "", EnvValue{}, errors.New(message)
	}

	if strings.Contains(entry.Name(), "=") {
		message := fmt.Sprintf(`%s have name with "="`, path)
		return "", EnvValue{}, errors.New(message)
	}

	info, err := entry.Info()
	if err != nil {
		return "", EnvValue{}, err
	}

	envName := entry.Name()
	if info.Size() == 0 {
		return envName, EnvValue{Value: "", NeedRemove: true}, nil
	}

	envValue, err := readEnvValue(path)
	if err != nil {
		return "", EnvValue{}, err
	}

	return envName, EnvValue{Value: envValue, NeedRemove: false}, nil
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	result := make(Environment)
	errorsList := make([]string, 0, len(files))
	for _, file := range files {
		envName, envValue, err := processDirEntry(dir, file)
		if err != nil {
			errorsList = append(errorsList, err.Error())
			continue
		}

		result[envName] = envValue
	}

	if len(errorsList) > 0 {
		return nil, errors.New(strings.Join(errorsList, "\n"))
	}

	return result, nil
}
