package sitemapper

import (
	"errors"
	"testing"
	"time"
)

// TestDefaultOptions ensures that the defaults aren't changed as users might rely on it.
func TestDefaultOptions(t *testing.T) {
	options := DefaultOptions()

	if options.domain != "http://localhost:8080" {
		t.Errorf("Expected default domain to be 'http://localhost:8080', got '%s' instead", options.domain)
	}

	if options.durationBeforeFirstCrawl != time.Second*3 {
		t.Errorf("Expected default durationBeforeFirstCrawl to be 3 seconds, got %v", options.durationBeforeFirstCrawl)
	}

	if options.crawlInterval != time.Hour*24*7 {
		t.Errorf("Expected default crawlInterval to be 1 week, got %v", options.crawlInterval)
	}

	if options.startingURL != "/" {
		t.Errorf("Expected default startingURL to be '/', got %s", options.startingURL)
	}

	if len(options.linkAttributes) != 0 {
		t.Errorf("Expected default linkAttributes to be empty, got %v", options.linkAttributes)
	}
}

func TestSetDomain(t *testing.T) {
	options := DefaultOptions()

	tests := []struct {
		input    string
		expected error
	}{
		{"http://example.com", nil},
		{"https://valid-domain.org", nil},
		{"http://localhost:8080", nil},
		{"ftp://invalid-scheme.com", errors.New("invalid domain: scheme must be 'http' or 'https'")},
		{"http://example.com/home", errors.New("invalid domain: must not include a relative path")},
		{"/home", errors.New("invalid domain: scheme must be 'http' or 'https'")},
		{"http://:8080", errors.New("invalid domain: must include a host")},
		{"invalid-url", errors.New("invalid domain: scheme must be 'http' or 'https'")},
		{"", errors.New("invalid domain: scheme must be 'http' or 'https'")},
	}

	for _, test := range tests {
		err := options.SetDomain(test.input)
		if (err == nil && test.expected != nil) || (err != nil && err.Error() != test.expected.Error()) {
			t.Errorf("SetDomain(%q) = %v, want %v", test.input, err, test.expected)
		}
	}
}

func TestSetDurationBeforeFirstCrawl(t *testing.T) {
	options := DefaultOptions()

	tests := []struct {
		input    time.Duration
		expected error
	}{
		{0, nil},               // Immediate crawling allowed
		{time.Second * 5, nil}, // Valid positive duration
		{-time.Second * 5, errors.New("invalid duration: cannot be negative")}, // Negative duration
	}

	for _, test := range tests {
		err := options.SetDurationBeforeFirstCrawl(test.input)
		if (err == nil && test.expected != nil) || (err != nil && err.Error() != test.expected.Error()) {
			t.Errorf("SetDurationBeforeFirstCrawl(%v) = %v, want %v", test.input, err, test.expected)
		}
	}
}

func TestSetCrawlInterval(t *testing.T) {
	options := DefaultOptions()

	tests := []struct {
		input    time.Duration
		expected error
	}{
		{time.Minute, nil},
		{time.Hour, nil},
		{time.Second * 59, nil},
		{-time.Minute, errors.New("invalid interval: cannot be negative")},
	}

	for _, test := range tests {
		err := options.SetCrawlInterval(test.input)
		if (err == nil && test.expected != nil) || (err != nil && err.Error() != test.expected.Error()) {
			t.Errorf("SetCrawlInterval(%v) = %v, want %v", test.input, err, test.expected)
		}
	}
}

func TestSetStartingURL(t *testing.T) {
	options := DefaultOptions()
	err := errors.New("invalid starting URL: must be a valid relative path")

	tests := []struct {
		input    string
		expected error
	}{
		{"/", nil},
		{"/path/to/start", nil},
		{"/path/to/start/", nil},
		{"path/to/start", err},
		{"invalid path", err},
		{"http://example.com", err},
		{"", err},
	}

	for _, test := range tests {
		err := options.SetStartingURL(test.input)
		if (err == nil && test.expected != nil) || (err != nil && err.Error() != test.expected.Error()) {
			t.Errorf("SetStartingURL(%q) = %v, want %v", test.input, err, test.expected)
		}
	}
}

func TestSetLinkAttributes(t *testing.T) {
	options := DefaultOptions()

	tests := []struct {
		input    []string
		expected error
	}{
		{[]string{"href"}, nil},          // Valid single attribute
		{[]string{"src", "hx-get"}, nil}, // Valid multiple attributes
		{[]string{}, errors.New("invalid link attributes: must provide at least one attribute")}, // Empty slice
	}

	for _, test := range tests {
		err := options.SetLinkAttributes(test.input...)
		if (err == nil && test.expected != nil) || (err != nil && err.Error() != test.expected.Error()) {
			t.Errorf("SetLinkAttributes(%v) = %v, want %v", test.input, err, test.expected)
		}
	}
}

func TestSetInfoLogger(t *testing.T) {
	options := DefaultOptions()

	var message string

	options.SetInfoLogger(func(msg string) {
		message = msg
	})

	testMessage := "Hello, World!"

	options.infoLogger(testMessage)

	if message != testMessage {
		t.Errorf("Expected message to be '%s', got '%s'", testMessage, message)
	}
}

func TestSetErrorLogger(t *testing.T) {
	options := DefaultOptions()

	var errMessage error

	options.SetErrorLogger(func(err error) {
		errMessage = err
	})

	testError := errors.New("hello, world!")

	options.errorLogger(testError)

	if errMessage != testError {
		t.Errorf("Expected errMessage to be '%s', got '%s'", testError, errMessage)
	}
}
