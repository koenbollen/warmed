package warmed_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/koenbollen/warmed"
)

func TestNew(t *testing.T) {
	t.Parallel()

	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Error("server called with header that was not HEAD")
		}
		serverCalled = true
	}))
	defer server.Close()

	client := warmed.New(server.URL + "/path")
	defer client.Close()

	targets := client.Targets()
	if len(targets) != 1 {
		t.Fatalf("expected one target, got %d", len(targets))
	}
	if targets[0] != server.URL+"/" {
		t.Errorf("expected target to be %v/, got %v", server.URL, targets[0])
	}

	if !serverCalled {
		t.Error("expected server to be called")
	}
}

func TestClient_Target(t *testing.T) {
	tests := []struct {
		name   string
		target string
		want   string
	}{
		{"simple", "https://example.org/path/file.ext", "https://example.org/"},
		{"host only", "example.org", ""},
		{"extra port", "http://example.org:8080/path", "http://example.org:8080/"},
		{"normalize", "HTTP://weirDuRL.com:80/path ", "http://weirDuRL.com:80/"},
		{"no trailing slash", "http://example.org", "http://example.org/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := warmed.New()
			defer c.Close()
			c.Target(tt.target)
			results := c.Targets()
			if tt.want == "" {
				if len(results) > 0 {
					t.Errorf("expected no result, got %v (%d)", results[0], len(results))
				}
			} else {
				if len(results) == 0 {
					t.Errorf("targeting %q has no results, expected %v", tt.target, tt.want)
				} else if results[0] != tt.want {
					t.Errorf("targeting %q resulted in %v, want %v", tt.target, results[0], tt.want)
				}
			}
		})
	}

	c := warmed.New()
	c.Target("https://example.org/", "https://example.org/path")
	result := c.Targets()
	c.Close()
	if len(result) != 1 {
		t.Errorf("same base urls results in more then one target: %v", result)
	}
}
