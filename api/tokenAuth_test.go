package api

import (
	"net/http"
	"testing"
)

func TestBearerGetToken(t *testing.T) {
	tests := []struct {
		name      string
		header    http.Header
		want      string
		wantedErr bool
	}{
		{
			name: "valid bearer token",
			header: http.Header{
				"Authorization": []string{"Bearer validate123"},
			},
			want:      "validate123",
			wantedErr: false,
		}, {
			name:      "missing authorization header",
			header:    http.Header{},
			want:      "",
			wantedErr: true,
		},
		{
			name: "invalid prefix",
			header: http.Header{
				"Authorization": []string{"Token validate123"},
			},
			want:      "",
			wantedErr: true,
		},
		{
			name: "no space between bearer and token",
			header: http.Header{
				"Authorization": []string{"Bearervalidate123"},
			},
			want:      "",
			wantedErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := GetBearerToken(test.header)
			if (err != nil) != test.wantedErr {
				t.Fatalf("Expected error: %v, got: %v", test.wantedErr, err)
			}
			if got != test.want {
				t.Fatalf("Expected token: %s, got %s", test.want, got)
			}
		})
	}
}
