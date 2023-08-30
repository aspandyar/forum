package validator

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*(?:\\.[a-zA-Z]{2,})$")

type Validator struct {
	NonFieldErrors []string
	FieldErrors    map[string]string
}

// j
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0 && len(v.NonFieldErrors) == 0
}

func (v *Validator) AddFieldError(key, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

func PermittedInt(value int, permittedValues ...int) bool {
	for i := range permittedValues {
		if value == permittedValues[i] {
			return true
		}
	}

	return false
}

func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func IncorrectInput(value string) bool {
	return checkTags(value) == nil
}

func checkTags(tagStr string) error {
	if strings.TrimSpace(tagStr) == "" {
		return nil
	}
	// Check for valid characters
	validCharsRegex := regexp.MustCompile(`^[A-Za-z0-9, ]+$`)
	if !validCharsRegex.MatchString(tagStr) {
		return errors.New("invalid characters in tags")
	}

	// Remove additional spaces
	tagStr = strings.TrimSpace(tagStr)

	// Check for valid start and end
	if len(tagStr) > 0 && !isLetter(tagStr[0]) {
		return errors.New("tag should start with a letter")
	}

	if len(tagStr) > 0 && !isLetter(tagStr[len(tagStr)-1]) {
		return errors.New("tag should end with a letter")
	}

	return nil
}

func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
