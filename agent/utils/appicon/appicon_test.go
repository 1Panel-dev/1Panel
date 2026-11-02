package appicon

import "testing"

func TestRemoteURLPreservesValidSources(t *testing.T) {
	for _, value := range []string{
		"https://cdn.example/icon.png",
		"http://cdn.example:8080/icon.png",
		"HTTPS://cdn.example/icon.png?token=a%2Bb&v=42#image",
		"https://[::1]:8443/icon.png",
	} {
		t.Run(value, func(t *testing.T) {
			if !IsRemoteURL(value) || RemoteURL(value) != value {
				t.Fatalf("URL was rejected or rewritten: %q", value)
			}
		})
	}
}

func TestRemoteURLRejectsOtherIconFormatsAndAmbiguousURLs(t *testing.T) {
	for _, value := range []string{
		"", "app_php.png?etag=42", "iVBORw0KGgo=", "data:image/png;base64,iVBORw0KGgo=",
		"javascript:alert(1)", "file:///icon.png", "//cdn.example/icon.png", "https:///icon.png",
		"https://", "https://user:secret@cdn.example/icon.png", "https://@cdn.example/icon.png",
		" https://cdn.example/icon.png", "https://cdn.example/icon.png ", "https://cdn.example/icon image.png",
		"https://cdn.example/\nicon.png", "https://cdn.example/icon.png\r\nLocation: https://other.example",
		"https://cdn.example/\\other.png", "https://cdn.example:bad/icon.png", "https://cdn.example/%zz",
		"https://cdn.example/icon.png?token=%zz", "https://cdn.example/\u00a0icon.png", "https://cdn.example/\x7ficon.png",
		"https://cdn.example:65536/icon.png",
	} {
		t.Run(value, func(t *testing.T) {
			if IsRemoteURL(value) || RemoteURL(value) != "" {
				t.Fatalf("invalid or historical icon was classified as a remote URL: %q", value)
			}
		})
	}
}
