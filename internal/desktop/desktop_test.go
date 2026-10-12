package desktop

import "testing"

func TestBrowserURLPolicy(t *testing.T) {
	for _, raw := range []string{"https://example.com/docs?q=go", "http://localhost:8080"} {
		if err := ValidateURL(raw); err != nil {
			t.Errorf("rejected %q: %v", raw, err)
		}
	}
	for _, raw := range []string{"file:///etc/passwd", "javascript:alert(1)", "myapp://run", "https://", "https://user:secret@example.com", "//example.com"} {
		if ValidateURL(raw) == nil {
			t.Errorf("accepted %q", raw)
		}
	}
}
