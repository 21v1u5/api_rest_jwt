package httpx

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSON(t *testing.T) {
	type input struct {
		Name string `json:"name"`
	}

	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"valid", `{"name":"penha"}`, false},
		{"empty body", ``, true},
		{"malformed", `{"name":`, true},
		{"wrong type", `{"name":123}`, true},
		{"unknown field", `{"name":"a","is_admin":true}`, true},
		{"two objects", `{"name":"a"}{"name":"b"}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			var dst input
			err := DecodeJSON(w, r, &dst)

			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
