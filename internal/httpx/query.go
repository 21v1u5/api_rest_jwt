package httpx

import (
	"net/url"
	"strconv"

	"github.com/21v1u5/api_rest_jwt/internal/validator"
)

func QueryInt(qs url.Values, key string, fallback int, v *validator.Validator) int {
	s := qs.Get(key)
	if s == "" {
		return fallback
	}

	n, err := strconv.Atoi(s)
	if err != nil {
		v.Check(false, key, "must be an integer")
		return fallback
	}
	return n
}

func QueryBool(qs url.Values, key string, v *validator.Validator) *bool {
	s := qs.Get(key)
	if s == "" {
		return nil
	}

	b, err := strconv.ParseBool(s)
	if err != nil {
		v.Check(false, key, "must be true or false")
		return nil
	}
	return &b
}
