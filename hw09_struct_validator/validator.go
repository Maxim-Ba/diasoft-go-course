package hw09structvalidator

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("обнаружены ошибки валидации:")
	for _, err := range v {
		builder.WriteString("\n\t")
		builder.WriteString(err.Field)
		builder.WriteString(": ")
		builder.WriteString(err.Err.Error())
	}
	return builder.String()
}

func Validate(v interface{}) error {
	val := reflect.ValueOf(v)

	// Если указатель или интерфейс
	if val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()
	var allErrors ValidationErrors

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		if !field.IsExported() {
			continue
		}

		tag := field.Tag.Get("validate")
		if tag == "" || tag == "-" {
			continue
		}

		errs, progErr := validateField(field.Name, fieldValue, tag)
		if progErr != nil {
			return progErr
		}
		allErrors = append(allErrors, errs...)
	}

	if len(allErrors) > 0 {
		return allErrors
	}
	return nil
}

// проверяет одно правило для конкретного значения.
type ValidatorFunc func(fieldName string, value interface{}, param string) (ValidationErrors, error)

var stringValidators = map[string]ValidatorFunc{
	"len":    validateStringLen,
	"regexp": validateStringRegexp,
	"in":     validateStringIn,
}

var intValidators = map[string]ValidatorFunc{
	"min": validateIntMin,
	"max": validateIntMax,
	"in":  validateIntIn,
}

// разбивает тег на отдельные правила
func splitRules(tag string) []string {
	parts := strings.Split(tag, "|")
	var result []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		// Ищем запятые, которые являются разделителями правил
		var subParts []string
		start := 0
		for i := 0; i < len(part); i++ {
			if part[i] == ',' {
				// Пропускаем пробелы после запятой
				j := i + 1
				for j < len(part) && part[j] == ' ' {
					j++
				}
				// Ищем конец следующего слова (имени валидатора)
				k := j
				for k < len(part) && (part[k] >= 'a' && part[k] <= 'z' || part[k] >= 'A' && part[k] <= 'Z') {
					k++
				}
				// Если после слова идёт двоеточие, значит это новое правило
				if k < len(part) && part[k] == ':' {
					subParts = append(subParts, strings.TrimSpace(part[start:i]))
					start = i + 1
				}
				// Иначе запятая внутри параметра, ничего не делаем
			}
		}
		if start < len(part) {
			subParts = append(subParts, strings.TrimSpace(part[start:]))
		}
		result = append(result, subParts...)
	}
	return result
}

// validateField применяет все правила из тега к одному полю.
func validateField(fieldName string, value reflect.Value, tag string) (ValidationErrors, error) {
	var fieldErrors ValidationErrors

	rules := splitRules(tag)
	for _, rule := range rules {
		if rule == "" {
			continue
		}

		// определяем тип поля
		switch value.Kind() {
		case reflect.String:
			errs, progErr := applyStringValidator(fieldName, value.String(), rule)
			if progErr != nil {
				return nil, progErr
			}
			fieldErrors = append(fieldErrors, errs...)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			errs, progErr := applyIntValidator(fieldName, value.Int(), rule)
			if progErr != nil {
				return nil, progErr
			}
			fieldErrors = append(fieldErrors, errs...)

		case reflect.Slice:
			errs, progErr := validateSliceField(fieldName, value, rule)
			if progErr != nil {
				return nil, progErr
			}
			fieldErrors = append(fieldErrors, errs...)

		default:
			// остальное игнорируем
		}
	}

	return fieldErrors, nil
}

//  применяет один валидатор к строковому значению.
func applyStringValidator(fieldName, strValue, rule string) (ValidationErrors, error) {
	validatorName, param := parseRule(rule)
	if validatorName == "" {
		return nil, fmt.Errorf("пустое правило для поля %s", fieldName)
	}

	// Ищем валидатор
	validator, ok := stringValidators[validatorName]
	if !ok {
		return nil, fmt.Errorf("неизвестный валидатор для строки %q в поле %s", validatorName, fieldName)
	}

	return validator(fieldName, strValue, param)
}

//  применяет один валидатор к целочисленному значению.
func applyIntValidator(fieldName string, intValue int64, rule string) (ValidationErrors, error) {
	validatorName, param := parseRule(rule)
	if validatorName == "" {
		return nil, fmt.Errorf("пустое правило для поля %s", fieldName)
	}

	validator, ok := intValidators[validatorName]
	if !ok {
		return nil, fmt.Errorf("неизвестный валидатор для числа %q в поле %s", validatorName, fieldName)
	}

	return validator(fieldName, intValue, param)
}

func validateSliceField(fieldName string, sliceValue reflect.Value, rule string) (ValidationErrors, error) {
	var allErrors ValidationErrors
	for i := 0; i < sliceValue.Len(); i++ {
		elem := sliceValue.Index(i)
		elemName := fmt.Sprintf("%s[%d]", fieldName, i)

		switch elem.Kind() {
		case reflect.String:
			errs, progErr := applyStringValidator(elemName, elem.String(), rule)
			if progErr != nil {
				return nil, progErr
			}
			allErrors = append(allErrors, errs...)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			errs, progErr := applyIntValidator(elemName, elem.Int(), rule)
			if progErr != nil {
				return nil, progErr
			}
			allErrors = append(allErrors, errs...)

		default:
			// Игнорируем
		}
	}
	return allErrors, nil
}

// разбирает строку в виде "имя:параметр" и возвращает имя и параметр.
func parseRule(rule string) (name, param string) {
	parts := strings.SplitN(rule, ":", 2)
	name = strings.TrimSpace(parts[0])
	if len(parts) == 2 {
		param = strings.TrimSpace(parts[1])
	}
	return
}

func validateStringLen(fieldName string, value interface{}, param string) (ValidationErrors, error) {
	str, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("внутренняя ошибка: ожидалась строка в validateStringLen")
	}
	length, err := strconv.Atoi(param)
	if err != nil {
		return nil, fmt.Errorf("некорректный параметр len %q для поля %s: %w", param, fieldName, err)
	}
	if len(str) != length {
		return ValidationErrors{{Field: fieldName, Err: fmt.Errorf("длина должна быть равна %d", length)}}, nil
	}
	return nil, nil
}

func validateStringRegexp(fieldName string, value interface{}, param string) (ValidationErrors, error) {
	str, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("внутренняя ошибка: ожидалась строка в validateStringRegexp")
	}
	re, err := regexp.Compile(param)
	if err != nil {
		return nil, fmt.Errorf("некорректное регулярное выражение %q для поля %s: %w", param, fieldName, err)
	}
	if !re.MatchString(str) {
		return ValidationErrors{{Field: fieldName, Err: fmt.Errorf("строка должна соответствовать регулярному выражению %q", param)}}, nil
	}
	return nil, nil
}

func validateStringIn(fieldName string, value interface{}, param string) (ValidationErrors, error) {
	str, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("внутренняя ошибка: ожидалась строка в validateStringIn")
	}
	allowed := strings.Split(param, ",")
	for i := range allowed {
		allowed[i] = strings.TrimSpace(allowed[i])
	}
	found := false
	for _, a := range allowed {
		if str == a {
			found = true
			break
		}
	}
	if !found {
		return ValidationErrors{{Field: fieldName, Err: fmt.Errorf("значение должно быть одним из [%s]", strings.Join(allowed, ", "))}}, nil
	}
	return nil, nil
}

func validateIntMin(fieldName string, value interface{}, param string) (ValidationErrors, error) {
	num, ok := value.(int64)
	if !ok {
		return nil, fmt.Errorf("внутренняя ошибка: ожидалось int64 в validateIntMin")
	}
	minVal, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("некорректный параметр min %q для поля %s: %w", param, fieldName, err)
	}
	if num < minVal {
		return ValidationErrors{{Field: fieldName, Err: fmt.Errorf("значение не может быть меньше %d", minVal)}}, nil
	}
	return nil, nil
}

func validateIntMax(fieldName string, value interface{}, param string) (ValidationErrors, error) {
	num, ok := value.(int64)
	if !ok {
		return nil, fmt.Errorf("внутренняя ошибка: ожидалось int64 в validateIntMax")
	}
	maxVal, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("некорректный параметр max %q для поля %s: %w", param, fieldName, err)
	}
	if num > maxVal {
		return ValidationErrors{{Field: fieldName, Err: fmt.Errorf("значение не может быть больше %d", maxVal)}}, nil
	}
	return nil, nil
}

func validateIntIn(fieldName string, value interface{}, param string) (ValidationErrors, error) {
	num, ok := value.(int64)
	if !ok {
		return nil, fmt.Errorf("внутренняя ошибка: ожидалось int64 в validateIntIn")
	}
	parts := strings.Split(param, ",")
	allowed := make([]int64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		val, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("некорректное число в списке in %q для поля %s: %w", param, fieldName, err)
		}
		allowed = append(allowed, val)
	}
	found := false
	for _, a := range allowed {
		if num == a {
			found = true
			break
		}
	}
	if !found {
		allowedStrs := make([]string, len(allowed))
		for i, a := range allowed {
			allowedStrs[i] = strconv.FormatInt(a, 10)
		}
		return ValidationErrors{{Field: fieldName, Err: fmt.Errorf("значение должно быть одним из [%s]", strings.Join(allowedStrs, ", "))}}, nil
	}
	return nil, nil
}
