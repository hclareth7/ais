package main

import "testing"

func TestValidateExternalURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "valid http", url: "http://example.com", wantErr: false},
		{name: "valid https", url: "https://example.com", wantErr: false},
		{name: "https with path", url: "https://example.com/page?q=1", wantErr: false},
		{name: "https with port", url: "https://example.com:8080/path", wantErr: false},
		{name: "javascript scheme", url: "javascript:alert(1)", wantErr: true},
		{name: "data scheme", url: "data:text/html,<h1>test</h1>", wantErr: true},
		{name: "file scheme", url: "file:///etc/passwd", wantErr: true},
		{name: "ftp scheme", url: "ftp://example.com/file", wantErr: true},
		{name: "empty string", url: "", wantErr: true},
		{name: "just text", url: "not-a-url", wantErr: true},
		{name: "missing scheme", url: "example.com", wantErr: true},
		{name: "scheme only", url: "https://", wantErr: true},
		{name: "relative path", url: "/path/to/page", wantErr: true},
		{name: "mailto scheme", url: "mailto:user@example.com", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateExternalURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateExternalURL(%q) error = %v, wantErr = %v", tt.url, err, tt.wantErr)
			}
		})
	}
}
