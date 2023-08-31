package hw09structvalidator

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLenValidator(t *testing.T) {
	tests := []struct {
		constraint  int
		value       interface{}
		expectedErr error
		expectedMsg string
	}{
		{
			constraint:  10,
			value:       "",
			expectedErr: ErrLenValidation,
			expectedMsg: ErrLenValidation.Error() + ": 0 is not equal to constraint value 10",
		},
		{
			constraint:  10,
			value:       "123456789",
			expectedErr: ErrLenValidation,
			expectedMsg: ErrLenValidation.Error() + ": 9 is not equal to constraint value 10",
		},
		{
			constraint:  10,
			value:       "1234567890A",
			expectedErr: ErrLenValidation,
			expectedMsg: ErrLenValidation.Error() + ": 11 is not equal to constraint value 10",
		},
		{
			constraint:  10,
			value:       0,
			expectedErr: ErrUnsupportedArgument,
			expectedMsg: ErrUnsupportedArgument.Error() + ": argument is not string",
		},
		{
			constraint:  10,
			value:       "1234567890",
			expectedErr: nil,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt

			validator := lenValidator{constraint: tt.constraint}
			err := validator.validate(tt.value)
			require.ErrorIs(t, err, tt.expectedErr)
			if tt.expectedMsg != "" {
				require.Equal(t, tt.expectedMsg, err.Error())
			}

			_ = tt
		})
	}
}

func TestInIntValidator(t *testing.T) {
	tests := []struct {
		constraint  []int
		value       interface{}
		expectedErr error
		expectedMsg string
	}{
		{
			constraint:  []int{},
			value:       1,
			expectedErr: ErrInValidation,
			expectedMsg: ErrInValidation.Error() + ": 1 is not in []",
		},
		{
			constraint:  []int{1, 22, 35},
			value:       21,
			expectedErr: ErrInValidation,
			expectedMsg: ErrInValidation.Error() + ": 21 is not in [1 22 35]",
		},
		{
			constraint:  []int{1, 22, 35},
			value:       "21",
			expectedErr: ErrUnsupportedArgument,
			expectedMsg: ErrUnsupportedArgument.Error() + ": argument is not int",
		},
		{
			constraint:  []int{1, 22, 35},
			value:       22,
			expectedErr: nil,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt

			validator := inIntValidator{constraint: tt.constraint}
			err := validator.validate(tt.value)
			require.ErrorIs(t, err, tt.expectedErr)
			if tt.expectedMsg != "" {
				require.Equal(t, tt.expectedMsg, err.Error())
			}

			_ = tt
		})
	}
}

func TestInStrValidator(t *testing.T) {
	tests := []struct {
		constraint  []string
		value       interface{}
		expectedErr error
		expectedMsg string
	}{
		{
			constraint:  []string{},
			value:       "",
			expectedErr: ErrInValidation,
			expectedMsg: ErrInValidation.Error() + ": \"\" is not in []",
		},
		{
			constraint:  []string{"one", "two", "three"},
			value:       "",
			expectedErr: ErrInValidation,
			expectedMsg: ErrInValidation.Error() + ": \"\" is not in [one two three]",
		},
		{
			constraint:  []string{"one", "two", "three"},
			value:       "123",
			expectedErr: ErrInValidation,
			expectedMsg: ErrInValidation.Error() + ": \"123\" is not in [one two three]",
		},
		{
			constraint:  []string{"one", "two", "three"},
			value:       21,
			expectedErr: ErrUnsupportedArgument,
			expectedMsg: ErrUnsupportedArgument.Error() + ": argument is not string",
		},
		{
			constraint:  []string{"one", "two", "three"},
			value:       "three",
			expectedErr: nil,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt

			validator := inStrValidator{constraint: tt.constraint}
			err := validator.validate(tt.value)
			require.ErrorIs(t, err, tt.expectedErr)
			if tt.expectedMsg != "" {
				require.Equal(t, tt.expectedMsg, err.Error())
			}

			_ = tt
		})
	}
}

func TestMinValidator(t *testing.T) {
	tests := []struct {
		constraint  int
		value       interface{}
		expectedErr error
		expectedMsg string
	}{
		{
			constraint:  10,
			value:       -1,
			expectedErr: ErrMinValidation,
			expectedMsg: ErrMinValidation.Error() + ": -1 less then constraint value 10",
		},
		{
			constraint:  10,
			value:       5,
			expectedErr: ErrMinValidation,
			expectedMsg: ErrMinValidation.Error() + ": 5 less then constraint value 10",
		},
		{
			constraint:  10,
			value:       "10",
			expectedErr: ErrUnsupportedArgument,
			expectedMsg: ErrUnsupportedArgument.Error() + ": argument is not int",
		},
		{
			constraint:  10,
			value:       10,
			expectedErr: nil,
		},
		{
			constraint:  10,
			value:       21,
			expectedErr: nil,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt

			validator := minValidator{constraint: tt.constraint}
			err := validator.validate(tt.value)
			require.ErrorIs(t, err, tt.expectedErr)
			if tt.expectedMsg != "" {
				require.Equal(t, tt.expectedMsg, err.Error())
			}

			_ = tt
		})
	}
}

func TestMaxValidator(t *testing.T) {
	tests := []struct {
		constraint  int
		value       interface{}
		expectedErr error
		expectedMsg string
	}{
		{
			constraint:  10,
			value:       21,
			expectedErr: ErrMaxValidation,
			expectedMsg: ErrMaxValidation.Error() + ": 21 is greater then constraint value 10",
		},
		{
			constraint:  10,
			value:       "21",
			expectedErr: ErrUnsupportedArgument,
			expectedMsg: ErrUnsupportedArgument.Error() + ": argument is not int",
		},
		{
			constraint:  10,
			value:       10,
			expectedErr: nil,
		},
		{
			constraint:  10,
			value:       -1,
			expectedErr: nil,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt

			validator := maxValidator{constraint: tt.constraint}
			err := validator.validate(tt.value)
			require.ErrorIs(t, err, tt.expectedErr)
			if tt.expectedMsg != "" {
				require.Equal(t, tt.expectedMsg, err.Error())
			}

			_ = tt
		})
	}
}

func TestRegexpValidator(t *testing.T) {
	regex := regexp.MustCompile(`^\d+$`)
	tests := []struct {
		constraint  *regexp.Regexp
		value       interface{}
		expectedErr error
		expectedMsg string
	}{
		{
			constraint:  regex,
			value:       21,
			expectedErr: ErrUnsupportedArgument,
			expectedMsg: ErrUnsupportedArgument.Error() + ": argument is not string",
		},
		{
			constraint:  regex,
			value:       "A6B0",
			expectedErr: ErrRegexValidation,
			expectedMsg: ErrRegexValidation.Error() + `: "A6B0" is not matched regexp "^\d+$"`,
		},
		{
			constraint:  regex,
			value:       "-1",
			expectedErr: ErrRegexValidation,
			expectedMsg: ErrRegexValidation.Error() + `: "-1" is not matched regexp "^\d+$"`,
		},
		{
			constraint:  regex,
			value:       "",
			expectedErr: ErrRegexValidation,
			expectedMsg: ErrRegexValidation.Error() + `: "" is not matched regexp "^\d+$"`,
		},
		{
			constraint:  regex,
			value:       "1024",
			expectedErr: nil,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt

			validator := regexpValidator{constraint: tt.constraint}
			err := validator.validate(tt.value)
			require.ErrorIs(t, err, tt.expectedErr)
			if tt.expectedMsg != "" {
				require.Equal(t, tt.expectedMsg, err.Error())
			}

			_ = tt
		})
	}
}
