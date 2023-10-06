package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestLogger(t *testing.T) {
	l, err := New("dev", "info", "console", []string{"stdout"}, []string{"stderr"})
	require.NoError(t, err)
	require.IsType(t, &zap.Logger{}, l)
}
