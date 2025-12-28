package paramutil

import (
	"net/http/httptest"
	"testing"
)

func TestReadIDParam(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		url        string
		expectedID int64
		expectErr  bool
	}{
		{
			name:       "reads valid positive integer",
			id:         "123",
			url:        "/todos/123",
			expectedID: 123,
		},
		{
			name:       "reads ID 1",
			id:         "1",
			url:        "/todos/1",
			expectedID: 1,
		},
		{
			name:       "reads large ID",
			id:         "9223372036854775807",
			url:        "/todos/9223372036854775807",
			expectedID: 9223372036854775807,
		},
		{
			name:      "invalid non-numeric id",
			id:        "abc",
			url:       "/todos/abc",
			expectErr: true,
		},
		{
			name:      "invalid empty string id",
			id:        "",
			url:       "/todos/",
			expectErr: true,
		},
		{
			name: "invalid negative id",
			id:   "-1",
			url:  "/todos/-1",
		},
		{
			name:      "invalid decimal id",
			id:        "1.5",
			url:       "/todos/1.5",
			expectErr: true,
		},
		{
			name: "zero id",
			id:   "0",
			url:  "/todos/0",
		},
		{
			name: "leading zero id",
			id:   "01",
			url:  "/todos/01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			req.SetPathValue("id", tt.id)

			id, err := ReadIDParam(req)
			if err != nil && !tt.expectErr {
				t.Errorf("Expected no error, got %v", err)
			} else if err == nil && tt.expectErr {
				t.Errorf("Expected error, got nil")
			}
			if tt.expectedID != 0 && id != tt.expectedID {
				t.Errorf("Expected ID %d, got %d", tt.expectedID, id)
			}
		})
	}
}
