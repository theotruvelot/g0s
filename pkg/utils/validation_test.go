package utils

import (
	"strings"
	"testing"
)

func TestValidateServerURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantError bool
		errorText string
	}{
		{
			name:      "empty URL",
			url:       "",
			wantError: true,
			errorText: "cannot be empty",
		},
		{
			name:      "invalid URL format",
			url:       "://invalid",
			wantError: true,
			errorText: "invalid URL format",
		},
		{
			name:      "no scheme",
			url:       "localhost:8080",
			wantError: true,
			errorText: "URL must use http or https scheme",
		},
		{
			name:      "invalid scheme - ftp",
			url:       "ftp://example.com",
			wantError: true,
			errorText: "URL must use http or https scheme",
		},
		{
			name:      "invalid scheme - file",
			url:       "file:///path/to/file",
			wantError: true,
			errorText: "URL must use http or https scheme",
		},
		{
			name:      "no host - http",
			url:       "http://",
			wantError: true,
			errorText: "URL must have a valid host",
		},
		{
			name:      "no host - https",
			url:       "https://",
			wantError: true,
			errorText: "URL must have a valid host",
		},
		{
			name:      "valid http URL",
			url:       "http://localhost:8080",
			wantError: false,
		},
		{
			name:      "valid https URL",
			url:       "https://api.example.com",
			wantError: false,
		},
		{
			name:      "valid http URL with path",
			url:       "http://localhost:8080/api/v1",
			wantError: false,
		},
		{
			name:      "valid https URL with port",
			url:       "https://example.com:443",
			wantError: false,
		},
		{
			name:      "valid http URL with IP",
			url:       "http://192.168.1.1:8080",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateServerURL(tt.url)

			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateServerURL() expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("ValidateServerURL() error = %v, want error containing %v", err, tt.errorText)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateServerURL() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestNormalizeServerURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "URL with trailing slash",
			url:  "http://localhost:8080/",
			want: "http://localhost:8080",
		},
		{
			name: "URL without trailing slash",
			url:  "http://localhost:8080",
			want: "http://localhost:8080",
		},
		{
			name: "URL with path and trailing slash",
			url:  "https://api.example.com/v1/",
			want: "https://api.example.com/v1",
		},
		{
			name: "URL with path without trailing slash",
			url:  "https://api.example.com/v1",
			want: "https://api.example.com/v1",
		},
		{
			name: "empty URL",
			url:  "",
			want: "",
		},
		{
			name: "URL with multiple trailing slashes",
			url:  "http://example.com///",
			want: "http://example.com//",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeServerURL(tt.url)
			if got != tt.want {
				t.Errorf("NormalizeServerURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
