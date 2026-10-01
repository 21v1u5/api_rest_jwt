package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var EmailRX = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) Check(ok bool, field, message string) {
	if ok {
		return
	}
	if _, exists := v.Errors[field]; !exists {
		v.Errors[field] = message
	}
}

func NotBlank(s string) bool {
	return strings.TrimSpace(s) != ""
}

func MinChars(s string, n int) bool {
	return utf8.RuneCountInString(s) >= n
}

func MaxChars(s string, n int) bool {
	return utf8.RuneCountInString(s) <= n
}

func Matches(s string, rx *regexp.Regexp) bool {
	return rx.MatchString(s)
}

type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return "validation failed"
}

func (v *Validator) Err() error {
	if v.Valid() {
		return nil
	}
	return &ValidationError{Fields: v.Errors}
}
