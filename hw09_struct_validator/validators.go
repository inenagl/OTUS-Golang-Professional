package hw09structvalidator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Имена валидаторов, как они должны писаться в тэге валидации.
const (
	lenValidatorName    = "len"
	regexpValidatorName = "regexp"
	inValidatorName     = "in"
	minValidatorName    = "min"
	maxValidatorName    = "max"
)

// Список поддерживаемых имён валидаторов.
var validatorNames = []string{
	lenValidatorName,
	regexpValidatorName,
	inValidatorName,
	minValidatorName,
	maxValidatorName,
}

// Интерфейс валидатора.
type validator interface {
	validate(v interface{}) error
}

// Далее различные валидаторы и их методы.
type lenValidator struct {
	constraint int
}

type inIntValidator struct {
	constraint []int
}

type inStrValidator struct {
	constraint []string
}

type minValidator struct {
	constraint int
}

type maxValidator struct {
	constraint int
}

type regexpValidator struct {
	constraint *regexp.Regexp
}

func (v lenValidator) validate(value interface{}) error {
	switch val := value.(type) {
	case string:
		if len(val) != v.constraint {
			return fmt.Errorf("%w: %d is not equal to constraint value %d",
				ErrLenValidation,
				len(val),
				v.constraint)
		}
	default:
		return fmt.Errorf("%w: argument is not string", ErrUnsupportedArgument)
	}

	return nil
}

func (v inIntValidator) validate(value interface{}) error {
	switch val := value.(type) {
	case int:
		for _, c := range v.constraint {
			if val == c {
				return nil
			}
		}
	default:
		return fmt.Errorf("%w: argument is not int", ErrUnsupportedArgument)
	}

	return fmt.Errorf("%w: %d is not in %v", ErrInValidation, value, v.constraint)
}

func (v inStrValidator) validate(value interface{}) error {
	switch val := value.(type) {
	case string:
		for _, c := range v.constraint {
			if val == c {
				return nil
			}
		}
	default:
		return fmt.Errorf("%w: argument is not string", ErrUnsupportedArgument)
	}

	return fmt.Errorf("%w: \"%s\" is not in %v", ErrInValidation, value, v.constraint)
}

func (v minValidator) validate(value interface{}) error {
	switch val := value.(type) {
	case int:
		if val < v.constraint {
			return fmt.Errorf("%w: %d less then constraint value %d", ErrMinValidation, val, v.constraint)
		}
	default:
		return fmt.Errorf("%w: argument is not int", ErrUnsupportedArgument)
	}

	return nil
}

func (v maxValidator) validate(value interface{}) error {
	switch val := value.(type) {
	case int:
		if val > v.constraint {
			return fmt.Errorf("%w: %d is greater then constraint value %d",
				ErrMaxValidation,
				val,
				v.constraint)
		}
	default:
		return fmt.Errorf("%w: argument is not int", ErrUnsupportedArgument)
	}

	return nil
}

func (v regexpValidator) validate(value interface{}) error {
	switch val := value.(type) {
	case string:
		if v.constraint.MatchString(val) {
			return nil
		}
	default:
		return fmt.Errorf("%w: argument is not string", ErrUnsupportedArgument)
	}

	return fmt.Errorf(`%w: "%s" is not matched regexp "%s"`, ErrRegexValidation, value, v.constraint.String())
}

// Интерфейс стратегии создания валидаторов.
// isSuitable определяет пригодность стратегии.
// makeValidator - реализация стратегии.
type validatorCreator interface {
	isSuitable(name string, fieldKind Kind) bool
	makeValidator(constraint string) (validator, error)
}

// Типы стратегий создания валидаторов.
type lenValidatorCreator struct{}

type regexpValidatorCreator struct{}

type inStrValidatorCreator struct{}

type inIntValidatorCreator struct{}

type minValidatorCreator struct{}

type maxValidatorCreator struct{}

// Список стратегий создания валидаторов.
var validatorCreatorsList = []validatorCreator{
	lenValidatorCreator{},
	regexpValidatorCreator{},
	inStrValidatorCreator{},
	inIntValidatorCreator{},
	minValidatorCreator{},
	maxValidatorCreator{},
}

// Далее список методов, реализующих интерфейс стратегий создания валидаторов.
func (v lenValidatorCreator) isSuitable(name string, fieldKind Kind) bool {
	return name == lenValidatorName && (fieldKind == KindString || fieldKind == KindArrayOfString)
}

func (v regexpValidatorCreator) isSuitable(name string, fieldKind Kind) bool {
	return name == regexpValidatorName && (fieldKind == KindString || fieldKind == KindArrayOfString)
}

func (v inStrValidatorCreator) isSuitable(name string, fieldKind Kind) bool {
	return name == inValidatorName && (fieldKind == KindString || fieldKind == KindArrayOfString)
}

func (v inIntValidatorCreator) isSuitable(name string, fieldKind Kind) bool {
	return name == inValidatorName && (fieldKind == KindInt || fieldKind == KindArrayOfInt)
}

func (v minValidatorCreator) isSuitable(name string, fieldKind Kind) bool {
	return name == minValidatorName && (fieldKind == KindInt || fieldKind == KindArrayOfInt)
}

func (v maxValidatorCreator) isSuitable(name string, fieldKind Kind) bool {
	return name == maxValidatorName && (fieldKind == KindInt || fieldKind == KindArrayOfInt)
}

func (v lenValidatorCreator) makeValidator(constraint string) (validator, error) {
	c, err := strconv.Atoi(constraint)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidValidationTag, err.Error())
	}

	return lenValidator{constraint: c}, nil
}

func (v regexpValidatorCreator) makeValidator(constraint string) (validator, error) {
	constraint = strings.TrimSpace(constraint)
	// По условию задачи вся строка целиком должна удовлетворять регекспу (см. примеры написания валидаторов в задании),
	// поэтому добавим в регекс ограничители начала и конца.
	if !strings.HasPrefix(constraint, "^") {
		constraint = "^" + constraint
	}
	if !strings.HasSuffix(constraint, "$") {
		constraint += "$"
	}

	c, err := regexp.Compile(constraint)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidValidationTag, err.Error())
	}

	return regexpValidator{constraint: c}, nil
}

func (v inStrValidatorCreator) makeValidator(constraint string) (validator, error) {
	return inStrValidator{constraint: strings.Split(constraint, ",")}, nil
}

func (v inIntValidatorCreator) makeValidator(constraint string) (validator, error) {
	cs := strings.Split(constraint, ",")
	c := make([]int, len(cs))
	for i, s := range cs {
		ci, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidValidationTag, err.Error())
		}
		c[i] = ci
	}

	return inIntValidator{constraint: c}, nil
}

func (v minValidatorCreator) makeValidator(constraint string) (validator, error) {
	c, err := strconv.Atoi(constraint)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidValidationTag, err.Error())
	}

	return minValidator{constraint: c}, nil
}

func (v maxValidatorCreator) makeValidator(constraint string) (validator, error) {
	c, err := strconv.Atoi(constraint)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidValidationTag, err.Error())
	}

	return maxValidator{constraint: c}, nil
}

// Фабрика валидаторов.
// Создаёт валидатор на основании записи в тэге валидации (названия и значения)
// с учетом применимости этой валидации к типу поля.
func createValidatorForKind(name string, constraint string, fieldKind Kind) (validator, error) {
	// Проверим, что имя валидатора в тэге нам известно
	nameExists := func(name string, names []string) bool {
		for _, v := range names {
			if name == v {
				return true
			}
		}
		return false
	}
	if !nameExists(name, validatorNames) {
		return nil, fmt.Errorf(`%w: unknown validator "%s"`, ErrInvalidValidationTag, name)
	}

	var result validator
	var err error
	// Бежим по цепочке стратегий создания валидаторов, находим пригодную и пытаемся создать валидатор.
	for _, creator := range validatorCreatorsList {
		if creator.isSuitable(name, fieldKind) {
			if result, err = creator.makeValidator(constraint); err != nil {
				return nil, err
			}
		}
	}

	// Не удалось подобрать валидатор, устанавливаем ошибку.
	if result == nil {
		err = fmt.Errorf(`%w: unsupported validator "%s" for type "%s"`,
			ErrInvalidValidationTag, name, SupportedFieldKinds[fieldKind])
	}

	return result, err
}
