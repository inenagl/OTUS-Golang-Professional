package hw09structvalidator

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int      `validate:"min:18|max:50"`
		Email  string   `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole `validate:"in:admin,stuff"`
		Phones []string `validate:"len:11"`
		meta   json.RawMessage
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}
)

// Проверки общего плана. Что валидируются только структуры, что валидации могут проходить и не проходить.
func TestValidate(t *testing.T) {
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		{
			in:          "Some string",
			expectedErr: ErrUnsupportedArgument,
		},
		{
			in:          []string{"Some", "string"},
			expectedErr: ErrUnsupportedArgument,
		},
		{
			in:          100,
			expectedErr: ErrUnsupportedArgument,
		},
		{
			in:          []int{100, 101, 102},
			expectedErr: ErrUnsupportedArgument,
		},
		{
			in:          UserRole("some role"),
			expectedErr: ErrUnsupportedArgument,
		},
		{
			in: User{
				ID:     "Not UUID",
				Name:   "Name",
				Age:    2,
				Email:  "user@email.ru",
				Role:   "user",
				Phones: []string{"+7(999)1234567", "(80)1234567"},
				meta:   json.RawMessage{1, 2, 3},
			},
			expectedErr: ErrValidationErrors,
		},
		{
			in:          App{Version: "1.2.0"},
			expectedErr: nil,
		},
		{
			in:          App{Version: "1.2.10"},
			expectedErr: ErrValidationErrors,
		},
		{
			in:          Token{Header: []byte{0, 1}, Payload: []byte{100}, Signature: []byte{128, 255}},
			expectedErr: nil,
		},
		{
			in:          Response{Code: 200, Body: "Some response text"},
			expectedErr: nil,
		},
		{
			in:          Response{Code: 600, Body: "Some response text"},
			expectedErr: ErrValidationErrors,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			err := Validate(tt.in)
			require.ErrorIs(t, err, tt.expectedErr)

			_ = tt
		})
	}
}

// Проверки неправильного использования тэгов валидации.
// (Несовместимость с типом поля, неправильное написание правил валидации и т.п.).
func TestValidateWrongTag(t *testing.T) {
	type InvalidTagForInt struct {
		IntField int `validate:"len:1"`
	}
	type InvalidTagForInt2 struct {
		IntField int `validate:"regexp:\\d+"`
	}
	type InvalidTagForStr struct {
		StrField string `validate:"min:10"`
	}
	type InvalidTagForStr2 struct {
		StrField string `validate:"max:10"`
	}
	type EmptyValidationRule struct {
		IntField int `validate:"min:"`
	}
	type IncorrectValidationRule struct {
		IntField int `validate:"max:AAA"`
	}

	tests := []interface{}{
		InvalidTagForInt{1},
		InvalidTagForInt2{100},
		InvalidTagForStr{"some string"},
		InvalidTagForStr2{"some string"},
		EmptyValidationRule{10},
		IncorrectValidationRule{2},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			err := Validate(tt)
			require.ErrorIs(t, err, ErrInvalidValidationTag)

			_ = tt
		})
	}
}

// Проверки того, как возвращаются ошибки валидации.
// (Накопление ошибок по всем полям структуры, накопление ошибок валидации полей-массивов).
func TestValidateCheckComplexValidationErrors(t *testing.T) {
	type CheckValidationErrors struct {
		Int      int      `validate:"min:5|max:10"`
		Str      string   `validate:"len:3|regexp:[ab]+"`
		IntEnum  int      `validate:"in:1,2,3"`
		StrEnum  string   `validate:"in:a,b,c"`
		ArrOfInt []int    `validate:"min:5|max:10|in:6,7"`
		ArrOfStr []string `validate:"len:3|regexp:[ab]+"`
	}

	tests := []struct {
		in       interface{}
		expected ValidationErrors
	}{
		{
			in: CheckValidationErrors{
				Int:      6,                             // Valid
				Str:      "aaa",                         // Valid
				IntEnum:  2,                             // Valid
				StrEnum:  "b",                           // Valid
				ArrOfInt: []int{6, 7, 7},                // { Valid, Valid, Valid }
				ArrOfStr: []string{"aab", "aaa", "bbb"}, // { Valid, Valid, Valid }
			},
			expected: ValidationErrors{},
		},
		{
			in: CheckValidationErrors{
				Int:      11,                            // ErrMaxValidation
				Str:      "aaa",                         // Valid
				IntEnum:  3,                             // Valid
				StrEnum:  "a",                           // Valid
				ArrOfInt: []int{6, 6, 7},                // { Valid, Valid, Valid }
				ArrOfStr: []string{"aab", "aaa", "bbb"}, // { Valid, Valid, Valid }
			},
			expected: ValidationErrors{
				ValidationError{Field: "Int", Err: ErrMaxValidation},
			},
		},
		{
			in: CheckValidationErrors{
				Int:      4,                             // ErrMinValidation
				Str:      "aaab",                        // ErrLenValidation
				IntEnum:  33,                            // ErrInValidation
				StrEnum:  "ab",                          // ErrInValidation
				ArrOfInt: []int{7, 7, 7},                // { Valid, Valid, Valid }
				ArrOfStr: []string{"aab", "aaa", "bbb"}, // { Valid, Valid, Valid }
			},
			expected: ValidationErrors{
				ValidationError{Field: "Int", Err: ErrMinValidation},
				ValidationError{Field: "Str", Err: ErrLenValidation},
				ValidationError{Field: "IntEnum", Err: ErrInValidation},
				ValidationError{Field: "StrEnum", Err: ErrInValidation},
			},
		},
		{
			in: CheckValidationErrors{
				Int:     5,     // Valid
				Str:     "abb", // Valid
				IntEnum: 1,     // Valid
				StrEnum: "d",   // ErrInValidation
				// { (ErrMinValidation, ErrInValidation), Valid, (ErrMaxValidation, ErrInValidation) }
				// expects: ErrMinValidation, ErrMaxValidation, ErrInValidation, ErrInValidation
				ArrOfInt: []int{4, 6, 11},
				ArrOfStr: []string{"aab", "aaa", "bbb"}, // { Valid, Valid, Valid }
			},
			expected: ValidationErrors{
				ValidationError{Field: "StrEnum", Err: ErrInValidation},
				ValidationError{Field: "ArrOfInt", Err: ErrMinValidation},
				ValidationError{Field: "ArrOfInt", Err: ErrMaxValidation},
				ValidationError{Field: "ArrOfInt", Err: ErrInValidation},
				ValidationError{Field: "ArrOfInt", Err: ErrInValidation},
			},
		},
		{
			in: CheckValidationErrors{
				Int:     4,     // ErrMinValidation
				Str:     "abc", // ErrRegexValidation
				IntEnum: 5,     // ErrInValidation
				StrEnum: "aa",  // ErrInValidation
				// { (ErrMinValidation, ErrInValidation), Valid, (ErrMaxValidation, ErrInValidation) }
				// expects: ErrMinValidation, ErrMaxValidation, ErrInValidation, ErrInValidation.
				ArrOfInt: []int{4, 6, 11},
				// { (ErrLenValidation), (ErrLenValidation, ErrRegexValidation), (ErrRegexValidation) }
				// expects: ErrLenValidation, ErrRegexValidation, ErrRegexValidation, ErrRegexValidation.
				ArrOfStr: []string{"aabb", "ac", "cde"},
			},
			expected: ValidationErrors{
				ValidationError{Field: "Int", Err: ErrMinValidation},
				ValidationError{Field: "Str", Err: ErrRegexValidation},
				ValidationError{Field: "IntEnum", Err: ErrInValidation},
				ValidationError{Field: "StrEnum", Err: ErrInValidation},
				ValidationError{Field: "ArrOfInt", Err: ErrMinValidation},
				ValidationError{Field: "ArrOfInt", Err: ErrMaxValidation},
				ValidationError{Field: "ArrOfInt", Err: ErrInValidation},
				ValidationError{Field: "ArrOfInt", Err: ErrInValidation},
				ValidationError{Field: "ArrOfStr", Err: ErrLenValidation},
				ValidationError{Field: "ArrOfStr", Err: ErrLenValidation},
				ValidationError{Field: "ArrOfStr", Err: ErrRegexValidation},
				ValidationError{Field: "ArrOfStr", Err: ErrRegexValidation},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			e := Validate(tt.in)
			if len(tt.expected) == 0 {
				require.NoError(t, e)
			} else {
				require.ErrorIs(t, e, ErrValidationErrors)
				for i, e := range e.(ValidationErrors) { //nolint:errorlint
					require.Equal(t, tt.expected[i].Field, e.Field)
					require.ErrorIs(t, e.Err, tt.expected[i].Err)
				}
			}

			_ = tt
		})
	}
}
