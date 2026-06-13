package auth

import (
	"testing"
	"time"
	"net/http"

	"github.com/google/uuid"
)

func TestGetBearerToken(t *testing.T) {
	tests := []struct{
		name     string
		input    http.Header
		expected string
		wantErr  bool
	}{
		{
			name:     "Valid Authorization header",
			input:    http.Header{
				"Authorization": {"Bearer TOKEN_STRING"},
			},
			expected: "TOKEN_STRING",
			wantErr:  false,
		},
		{
			name:     "Missing Authorization header",
			input:    http.Header{
				"no Auth": {"Bearer TOKEN_STRING"},
			},
			expected: "",
			wantErr:  true,
		},
		{
			name:     "Malformed Authorization header",
			input:    http.Header{
				"Authorization": {"Malformed Authorization header"},
			},
			expected: "",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := GetBearerToken(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBearerToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if output != tt.expected {
				t.Errorf("GetBearerToken() output = %v, want %v", output, tt.expected)
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	validToken, _ := MakeJWT(userID, "secret", time.Hour)

	tests := []struct{
		name        string
		tokenString string
		tokenSecret string
		wantUserID  uuid.UUID
		wantErr     bool
	}{
		{
			name:        "Valid token",
			tokenString: validToken,
			tokenSecret: "secret",
			wantUserID:  userID,
			wantErr:     false,
		},
		{
			name:        "Invalid token",
			tokenString: "invalid.token.string",
			tokenSecret: "secret",
			wantUserID:  uuid.Nil,
			wantErr:     true,
		},
		{
			name:        "Wrong secret",
			tokenString: validToken,
			tokenSecret: "wrong_secret",
			wantUserID:  uuid.Nil,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUserID, err := ValidateJWT(tt.tokenString, tt.tokenSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUserID != tt.wantUserID {
				t.Errorf("ValidateJWT() gotUserID = %v, want %v", gotUserID, tt.wantUserID)
			}
		})
	}
}
