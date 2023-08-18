package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTempEnvDir(t *testing.T, name string) string {
	t.Helper()
	dir, err := os.MkdirTemp("", name)
	if err != nil {
		t.Error(err)
	}

	return dir
}

func createEnvFile(t *testing.T, dir string, name string, text string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	file, err := os.Create(path)
	if err != nil {
		t.Error(err)
	}

	if len(text) > 0 {
		if _, err := file.WriteString(text); err != nil {
			t.Error(err)
		}
	}

	if err := file.Close(); err != nil {
		t.Error(err)
	}

	return path
}

func TestReadDir(t *testing.T) {
	envDir := createTempEnvDir(t, "env")
	defer func() {
		err := os.RemoveAll(envDir)
		if err != nil {
			t.Error(err)
		}
	}()

	createEnvFile(t, envDir, "VAR", "Some right-trimmed string  \t \t")
	createEnvFile(t, envDir, "EMPTY", " ")
	createEnvFile(t, envDir, "VAR_TO_REMOVE", "")
	createEnvFile(t, envDir, "MULTILINE", "\t   First\tline \nSecond line\n\nFourth line")
	createEnvFile(t, envDir, "ZERO_CODE", " First line\t\u0000And second line \nNot used line")

	res, err := ReadDir(envDir)
	require.NoError(t, err)
	for envName, envValue := range res {
		switch envName {
		case "VAR":
			require.Equal(t, EnvValue{Value: "Some right-trimmed string", NeedRemove: false}, envValue)
		case "EMPTY":
			require.Equal(t, EnvValue{Value: "", NeedRemove: false}, envValue)
		case "VAR_TO_REMOVE":
			require.Equal(t, EnvValue{Value: "", NeedRemove: true}, envValue)
		case "MULTILINE":
			require.Equal(t, EnvValue{Value: "\t   First\tline", NeedRemove: false}, envValue)
		case "ZERO_CODE":
			require.Equal(t, EnvValue{Value: " First line\t\nAnd second line", NeedRemove: false}, envValue)
		default:
			assert.FailNow(t, fmt.Sprintf("Unexpected envVar: %s => %v", envName, envValue))
		}
	}
}

func TestReadDirErrors(t *testing.T) {
	envDir := createTempEnvDir(t, "env")
	defer func() {
		err := os.RemoveAll(envDir)
		if err != nil {
			t.Error(err)
		}
	}()

	expectedMessages := make([]string, 0, 2)

	// not a file - error.
	subDir := filepath.Join(envDir, "DIR_VAR")
	err := os.Mkdir(subDir, 0700) //nolint:gofumpt
	if err != nil {
		t.Error(err)
	}
	expectedMessages = append(expectedMessages, fmt.Sprintf("%s is directory", subDir))

	// "=" in filename - error.
	fileVar := createEnvFile(t, envDir, "V=A=R", "Var 1")
	expectedMessages = append(expectedMessages, fmt.Sprintf(`%s have name with "="`, fileVar))

	// Valid files - no errors
	createEnvFile(t, envDir, "SUCCESS", "Some success value")
	createEnvFile(t, envDir, "SUCCESS_DELETE", "")

	_, err = ReadDir(envDir)
	require.Error(t, err)
	errMessages := strings.Split(err.Error(), "\n")

	require.Equal(t, expectedMessages, errMessages)
}
