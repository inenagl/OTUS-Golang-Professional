package main

import (
	"math/rand"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	const cmdPath = "./testdata/testExitCode.sh"

	t.Run("no args", func(t *testing.T) {
		res := RunCmd([]string{cmdPath, ""}, Environment{})
		require.Equal(t, 0, res)
	})

	t.Run("with arg", func(t *testing.T) {
		exitCode := rand.Intn(127)
		res := RunCmd([]string{cmdPath, strconv.Itoa(exitCode)}, Environment{})
		require.Equal(t, exitCode, res)
	})

	t.Run("with env", func(t *testing.T) {
		exitCode := rand.Intn(127)
		env := Environment{
			"EXIT_CODE": EnvValue{Value: strconv.Itoa(exitCode), NeedRemove: false},
		}
		res := RunCmd([]string{cmdPath}, env)
		require.Equal(t, exitCode, res)

		// При запуске тестов в режиме -race эта переменная влияет на соседние тесты.
		// Поэтому нужно удалить её.
		if err := os.Unsetenv("EXIT_CODE"); err != nil {
			t.Error(err)
		}
	})
}

func TestRunCmdErrors(t *testing.T) {
	t.Run("empty command", func(t *testing.T) {
		res := RunCmd([]string{}, Environment{})
		require.Equal(t, exitCode, res)
	})

	t.Run("not existent command", func(t *testing.T) {
		res := RunCmd([]string{"not/existent/command", ""}, Environment{})
		require.Equal(t, exitCode, res)
	})
}

// В данном случае очень сильно удобнее напрямую протестировать метод установки окружения у команды,
// хоть он и приватный, чем проверять косвенно через RunCmd.
func TestProcessEnv(t *testing.T) {
	t.Run("empty env", func(t *testing.T) {
		currentEnv := os.Environ()
		res, err := processEnv(Environment{})
		require.NoError(t, err)
		require.Equal(t, currentEnv, res)
	})

	t.Run("not empty env", func(t *testing.T) {
		setEnv := func(key, value string) {
			if err := os.Setenv(key, value); err != nil {
				t.Error(err)
			}
		}
		unsetEnv := func(key string) {
			if err := os.Unsetenv(key); err != nil {
				t.Error(err)
			}
		}
		// Как должно быть после processEnv.
		setEnv("TO_UPDATE", "Updated")
		setEnv("TO_EMPTY", "")
		setEnv("NO_CHANGE", "Does not changed")
		setEnv("TO_ADD", "Added")
		expected := os.Environ()

		env := Environment{
			"TO_REMOVE": EnvValue{Value: "", NeedRemove: true},
			"TO_UPDATE": EnvValue{Value: "Updated", NeedRemove: false},
			"TO_EMPTY":  EnvValue{Value: "", NeedRemove: false},
			"TO_ADD":    EnvValue{Value: "Added", NeedRemove: false},
		}

		// Устанавливаем исходные значения окружения, которые будем менять в processEnv.
		setEnv("TO_REMOVE", "Need be removed")
		setEnv("TO_UPDATE", "Need be updated")
		setEnv("TO_EMPTY", "Need be cleared")
		unsetEnv("TO_ADD")
		require.NotEqual(t, expected, os.Environ())

		res, err := processEnv(env)
		require.NoError(t, err)
		require.Equal(t, expected, res)
	})
}
