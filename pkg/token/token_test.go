package token

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	token, err := Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(token, "rp_") {
		t.Errorf("token = %q, want prefix %q", token, "rp_")
	}

	if len(token) != 3+64 {
		t.Errorf("len(token) = %d, want %d", len(token), 3+64)
	}

	other, _ := Generate()
	if token == other {
		t.Errorf("two calls returned same token: %s", token)
	}
}

func TestGenerate_Length(t *testing.T) {
	for i := 0; i < 10; i++ {
		token, err := Generate()
		if err != nil {
			t.Fatal(err)
		}
		if len(token) != 3+64 {
			t.Errorf("len(token) = %d, want %d", len(token), 3+64)
		}
	}
}
