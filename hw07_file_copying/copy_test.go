package main

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func createTempFile(t *testing.T, name string, text string) string {
	t.Helper()

	file, err := os.CreateTemp("", name)
	if err != nil {
		t.Error(err)
	}
	path := file.Name()

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

func readString(t *testing.T, path string) string {
	t.Helper()

	b, err := os.ReadFile(path)
	if err != nil {
		t.Error(err)
	}

	return string(b)
}

func TestCopyWithWrongParams(t *testing.T) {
	tests := []struct {
		name     string
		fromPath string
		toPath   string
		offset   int64
		limit    int64
		expected []string
	}{
		{
			name:     "empty fromPath",
			fromPath: "",
			toPath:   "out.txt",
			offset:   0,
			limit:    0,
			expected: []string{MsgEmptyFrom},
		},
		{
			name:     "empty toPath",
			fromPath: "in.txt",
			toPath:   "",
			offset:   0,
			limit:    0,
			expected: []string{MsgEmptyTo},
		},
		{
			name:     "negative offset",
			fromPath: "in.txt",
			toPath:   "out.txt",
			offset:   -1,
			limit:    0,
			expected: []string{MsgNegativeOffset},
		},
		{
			name:     "negative limit",
			fromPath: "in.txt",
			toPath:   "out.txt",
			offset:   0,
			limit:    -1,
			expected: []string{MsgNegativeLimit},
		},
		{
			name:     "all is wrong",
			fromPath: "",
			toPath:   "",
			offset:   -1,
			limit:    -1,
			expected: []string{MsgEmptyFrom, MsgEmptyTo, MsgNegativeOffset, MsgNegativeLimit},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := Copy(tc.fromPath, tc.toPath, tc.offset, tc.limit)
			require.EqualError(t, err, strings.Join(tc.expected, "\n"))
		})
	}
}

func TestCopyWithErrors(t *testing.T) {
	t.Run("no source file", func(t *testing.T) {
		err := Copy("no_existent_file.txt", "out.txt", 0, 0)
		require.ErrorContains(t, err, "no such file or directory")
	})

	t.Run("no measurable file", func(t *testing.T) {
		err := Copy("/dev/urandom", "out.txt", 0, 0)
		require.ErrorIs(t, err, ErrUnsupportedFile)
	})

	t.Run("large offset", func(t *testing.T) {
		fromPath := createTempFile(t, "source.txt", "1234567890")
		defer os.Remove(fromPath)

		info, err := os.Stat(fromPath)
		if err != nil {
			t.Error(err)
		}

		err = Copy(fromPath, "out.txt", info.Size()+1, 0)
		require.ErrorIs(t, err, ErrOffsetExceedsFileSize)
	})
}

func TestCopy(t *testing.T) {
	fromPath := createTempFile(t, "source.txt", "1234567890")
	toPath := createTempFile(t, "out.txt", "")
	defer os.Remove(fromPath)
	defer os.Remove(toPath)

	t.Run("full copy", func(t *testing.T) {
		err := Copy(fromPath, toPath, 0, 0)
		require.NoError(t, err)
		require.Equal(t, readString(t, toPath), "1234567890")
	})

	t.Run("copy with offset", func(t *testing.T) {
		err := Copy(fromPath, toPath, 5, 0)
		require.NoError(t, err)
		require.Equal(t, readString(t, toPath), "67890")
	})

	t.Run("copy with offset and large limit", func(t *testing.T) {
		err := Copy(fromPath, toPath, 5, 10)
		require.NoError(t, err)
		require.Equal(t, readString(t, toPath), "67890")
	})

	t.Run("copy with offset and small limit", func(t *testing.T) {
		err := Copy(fromPath, toPath, 4, 2)
		require.NoError(t, err)
		require.Equal(t, readString(t, toPath), "56")
	})

	t.Run("copy with limit", func(t *testing.T) {
		err := Copy(fromPath, toPath, 0, 7)
		require.NoError(t, err)
		require.Equal(t, readString(t, toPath), "1234567")
	})
}
