package sitemapper

import (
	"bytes"
	"slices"
	"testing"
)

func TestExtractLinks(t *testing.T) {
	c := newCrawler("http://example.com", []string{"hx-get"}, nil, nil)

	htmlContent := `
	<html>
		<body>
			<a href="http://example.com/page1">Page 1</a>
			<a href="/page2">Page 2</a>
			<a href="javascript:void(0)">Invalid</a>
			<div>
				<div>
					<div>
						<div>
							<button hx-get="/htmx">Deeply Nested Button</button>
						</div>
					</div>
				</div>
			</div>
		</body>
	</html>
	`

	expectedLinks := []string{
		"http://example.com/page1",
		"http://example.com/page2",
		"http://example.com/htmx",
	}

	reader := bytes.NewReader([]byte(htmlContent))
	links := c.extractLinks(reader)

	for _, link := range expectedLinks {
		if !slices.Contains(links, link) {
			t.Errorf("Expected to find '%s' but did not", link)
		}
	}

	for _, link := range links {
		if !slices.Contains(expectedLinks, link) {
			t.Errorf("Found '%s' but should not have", link)
		}
	}
}

func TestNormalizeURL(t *testing.T) {
	c := newCrawler("http://example.com", nil, nil, nil)

	tests := []struct {
		input    string
		expected string
		valid    bool
	}{
		{"http://example.com/page", "http://example.com/page", true},
		{"/page", "http://example.com/page", true},
		{"javascript:void(0)", "", false},
		{"http://otherdomain.com", "", false},
		{"http://example.com/page#fragment", "http://example.com/page", true},
		{"", "", false},
	}

	for _, test := range tests {
		normalized, ok := c.normalizeURL(test.input)
		if ok != test.valid {
			t.Errorf("Expected validity '%v' for URL '%s', got '%v'", test.valid, test.input, ok)
		}
		if normalized != test.expected {
			t.Errorf("Expected normalized URL '%s' for input '%s', got '%s'", test.expected, test.input, normalized)
		}
	}
}

func TestEnsureTrailingSlash(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"http://example.com", "http://example.com/"},
		{"http://example.com/page", "http://example.com/page/"},
		{"http://example.com/page/", "http://example.com/page/"},
		{"http://example.com/image.jpg", "http://example.com/image.jpg"},
		{"http://example.com/folder", "http://example.com/folder/"},
	}

	for _, test := range tests {
		result := ensureTrailingSlash(test.input)
		if result != test.expected {
			t.Errorf("Expected %s for input %s, got %s", test.expected, test.input, result)
		}
	}
}
