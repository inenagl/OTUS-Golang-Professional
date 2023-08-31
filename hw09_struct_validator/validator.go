package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const ValidationTagName = "validate"

// Kind Украл идею из reflection. Поддерживаемый валидаторами тип поля.
type Kind uint

const (
	KindUnsupported Kind = iota
	KindString
	KindInt
	KindArrayOfString
	KindArrayOfInt
)

func isArrayKind(k Kind) bool {
	return k == KindArrayOfString || k == KindArrayOfInt
}

// SupportedFieldKinds Текстовое представление поддерживаемых типов полей.
var SupportedFieldKinds = map[Kind]string{
	KindString:        "string",
	KindInt:           "int",
	KindArrayOfString: "array of strings",
	KindArrayOfInt:    "array of int",
}

var (
	ErrUnsupportedArgument  = errors.New("unsupported argument")
	ErrInvalidValidationTag = errors.New("invalid validation tag")
	ErrValidation           = errors.New("validation error")
	ErrValidationErrors     = errors.New("array of validation errors")
	ErrLenValidation        = fmt.Errorf("%w (len)", ErrValidation)
	ErrInValidation         = fmt.Errorf("%w (in)", ErrValidation)
	ErrMinValidation        = fmt.Errorf("%w (min)", ErrValidation)
	ErrMaxValidation        = fmt.Errorf("%w (max)", ErrValidation)
	ErrRegexValidation      = fmt.Errorf("%w (regexp)", ErrValidation)
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	messages := make([]string, len(v))
	for i, ve := range v {
		messages[i] = fmt.Sprintf("%s: %s", ve.Field, ve.Err.Error())
	}

	return strings.Join(messages, "\n")
}

// Is Теперь инстанс ValidationErrors можно сравнивать через errors.Is c ErrValidationErrors.
func (v ValidationErrors) Is(target error) bool {
	return errors.Is(target, ErrValidationErrors)
}

// Возвращает тип поля (Kind) и значение поля.
// Если поле - массив, то тип возвращаемого значения []interface{} для удобства передачи элементов массива в валидаторы.
func getFieldKindAndValue(rfield reflect.StructField, rvalue reflect.Value) (Kind, interface{}) {
	kind := KindUnsupported
	var value interface{}

	fieldType := rfield.Type.Kind()
	if fieldType == reflect.Slice || fieldType == reflect.Array {
		switch rfield.Type.Elem().Kind() { //nolint: exhaustive
		case reflect.String:
			kind = KindArrayOfString
			v := make([]interface{}, rvalue.Len())
			for i := 0; i < rvalue.Len(); i++ {
				v[i] = rvalue.Index(i).String()
			}
			value = v
		case reflect.Int:
			kind = KindArrayOfInt
			v := make([]interface{}, rvalue.Len())
			for i := 0; i < rvalue.Len(); i++ {
				v[i] = int(rvalue.Index(i).Int())
			}
			value = v
		}
	} else {
		switch fieldType { //nolint: exhaustive
		case reflect.String:
			kind = KindString
			value = rvalue.String()
		case reflect.Int:
			kind = KindInt
			value = int(rvalue.Int())
		}
	}

	return kind, value
}

// Validate Функция получилась не очень короткой и с превышением проверки gocognit по сложности.
// Но я считаю, что она понятно читается, а при разбиении на более мелкие функции начинаются проблемы
// с передачей и возвратом разных данных, типов ошибок и т.д. в эти функции, что только усложняет код
// и делает его менее эффективным (т.е. много лишней работы по передаче аргументов в подфункции и обработке ответов).
func Validate(v interface{}) error { //nolint:gocognit
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("%w: argument is not a struct", ErrUnsupportedArgument)
	}

	var validationErrors ValidationErrors
	for i := 0; i < rv.NumField(); i++ {
		fieldTypeInfo := rv.Type().Field(i)
		fieldValueInfo := rv.Field(i)
		fieldName := fieldTypeInfo.Name

		validationTagValue, ok := fieldTypeInfo.Tag.Lookup(ValidationTagName)
		// Нет тэга валидации - пропускаем поле.
		if !ok {
			continue
		}

		// Определим тип поля и его значение.
		// Проверим поддерживается ли тип поля.
		fieldKind, fieldValue := getFieldKindAndValue(fieldTypeInfo, fieldValueInfo)
		if fieldKind == KindUnsupported {
			continue
		}

		// Выделим из тэга описания валидаторов.
		validationRules := strings.Split(validationTagValue, "|")
		if len(validationRules) == 0 {
			continue
		}

		// Получаем объекты валидаторов для нашего поля.
		// И валидируем поле.
		for _, ruleStr := range validationRules {
			ruleParts := strings.SplitN(ruleStr, ":", 2)
			if len(ruleParts) != 2 {
				return fmt.Errorf("%w: no validation rules desribed in \"%s\"", ErrInvalidValidationTag, ruleStr)
			}

			validator, err := createValidatorForKind(ruleParts[0], ruleParts[1], fieldKind)
			if err != nil {
				return err
			}

			if isArrayKind(fieldKind) { //nolint:nestif
				// Массивы валидируем поэлементно.
				for _, elem := range fieldValue.([]interface{}) {
					if err = validator.validate(elem); err != nil {
						if errors.Is(err, ErrValidation) {
							validationErrors = append(validationErrors, ValidationError{Field: fieldName, Err: err})
						} else {
							return err
						}
					}
				}
			} else {
				if err = validator.validate(fieldValue); err != nil {
					if errors.Is(err, ErrValidation) {
						validationErrors = append(validationErrors, ValidationError{Field: fieldName, Err: err})
					} else {
						return err
					}
				}
			}
		}
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}

	return nil
}
