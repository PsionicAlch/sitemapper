package sitemapper

import (
	"errors"
	"net/url"
	"strings"
	"time"
)

// SiteMapperOptions defines the configuration options for SiteMapper.
//
// These options allow customization of crawling behavior, logging, and site-specific details.
type SiteMapperOptions struct {
	// domain specifies the domain of the site for which the sitemap will be generated.
	//
	// Example: "https://example.com" or "http://localhost:8080"
	domain string

	// durationBeforeFirstCrawl is the delay before the crawler performs its first crawl.
	//
	// This can be used to avoid immediate crawling after initialization.
	durationBeforeFirstCrawl time.Duration

	// crawlInterval specifies the frequency at which the site is recrawled and the sitemap updated.
	//
	// Example: `time.Hour * 24` for daily crawling.
	crawlInterval time.Duration

	// startingURL is the URL where the crawler will begin crawling.
	//
	// If empty, it defaults to the root path ("/").
	startingURL string

	// linkAttributes are the HTML attributes (like href, src, etc.) that the crawler will parse to find links.
	// Example:
	//	[]string{"hx-get", "src"}
	linkAttributes []string

	// infoLogger is a function for logging informational messages. Example:
	//	func(msg string) { fmt.Println("INFO:", msg) }
	infoLogger func(string)

	// errorLogger is a function for logging errors that may occur during crawling. Example:
	//	func(err error) { fmt.Println("ERROR:", err.Error()) }
	errorLogger func(error)

	// callbackFunc is a function that will be called after crawling has finished. Since it needs
	// to be set before an instance of SiteMapper has been created we will pass the instance to
	// the callback function so that users have access to functions like GenerateSitemap if they
	// need it.
	callbackFunc func(*SiteMapper)
}

// DefaultOptions creates an instance of SiteMapperOptions with pre-defined default values.
//
// - Domain defaults to "http://localhost:8080".
//
// - Duration Before First Crawl defaults to 3 seconds.
//
// - Crawl Interval defaults to one week.
//
// - Starting URL defaults to "/".
//
// - Link Attributes defaults to an empty list.
//
// - Logging functions are empty by default and can be set later.
//
// - Callback function is empty by default and can be set later.
func DefaultOptions() *SiteMapperOptions {
	return &SiteMapperOptions{
		domain:                   "http://localhost:8080",
		durationBeforeFirstCrawl: time.Second * 3,
		crawlInterval:            time.Hour * 24 * 7,
		startingURL:              "/",
		linkAttributes:           []string{},
		infoLogger:               func(msg string) {},
		errorLogger:              func(err error) {},
		callbackFunc:             func(mapper *SiteMapper) {},
	}
}

// SetDomain updates the domain name of the site to crawl.
//
// Only domains with "http" or "https" schemes are allowed, and no relative path should be included.
func (options *SiteMapperOptions) SetDomain(domain string) error {
	parsedURL, err := url.Parse(domain)
	if err != nil {
		return errors.New("invalid domain: must be a valid URL")
	}

	// Ensure the scheme is "http" or "https".
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("invalid domain: scheme must be 'http' or 'https'")
	}

	// Ensure no relative path is present (path should be empty or "/").
	if parsedURL.Path != "" && parsedURL.Path != "/" {
		return errors.New("invalid domain: must not include a relative path")
	}

	// Ensure host is not empty.
	if parsedURL.Hostname() == "" {
		return errors.New("invalid domain: must include a host")
	}

	options.domain = strings.TrimRight(domain, "/")
	return nil
}

// SetDurationBeforeFirstCrawl updates the time delay before the initial crawl occurs.
// This is useful to control when the first crawl starts after initialization.
func (options *SiteMapperOptions) SetDurationBeforeFirstCrawl(duration time.Duration) error {
	if duration < 0 {
		return errors.New("invalid duration: cannot be negative")
	}

	options.durationBeforeFirstCrawl = duration

	return nil
}

// SetCrawlInterval sets the interval for recrawling the site and updating the sitemap.
// Example:
//
//	options.SetCrawlInterval(time.Hour * 24) // for daily crawling.
func (options *SiteMapperOptions) SetCrawlInterval(interval time.Duration) error {
	if interval < 0 {
		return errors.New("invalid interval: cannot be negative")
	}

	options.crawlInterval = interval

	return nil
}

// SetStartingURL sets the URL where the crawler begins its process.
//
// Only relative paths (e.g., "/path") are allowed.
func (options *SiteMapperOptions) SetStartingURL(urlPath string) error {
	parsedURL, err := url.Parse(urlPath)
	if err != nil || parsedURL.IsAbs() || !strings.HasPrefix(urlPath, "/") {
		return errors.New("invalid starting URL: must be a valid relative path")
	}

	options.startingURL = urlPath

	return nil
}

// SetLinkAttributes specifies which HTML attributes the crawler should inspect for URLs.
// For example:
//
//	options.SetLinkAttributes("hx-get", "src")
func (options *SiteMapperOptions) SetLinkAttributes(attributes ...string) error {
	if len(attributes) == 0 {
		return errors.New("invalid link attributes: must provide at least one attribute")
	}

	options.linkAttributes = attributes

	return nil
}

// SetInfoLogger assigns a logging function to handle informational messages. Example:
//
//	options.SetInfoLogger(func(msg string) {
//		log.Println(msg)
//	})
func (options *SiteMapperOptions) SetInfoLogger(logger func(string)) {
	options.infoLogger = func(msg string) {
		if logger != nil {
			logger(msg)
		}
	}
}

// SetErrorLogger assigns a logging function to handle error messages. Example:
//
//	options.SetErrorLogger(func(err error) {
//		log.Println(err.Error())
//	})
func (options *SiteMapperOptions) SetErrorLogger(logger func(error)) {
	options.errorLogger = func(err error) {
		if logger != nil {
			logger(err)
		}
	}
}

// SetCallbackFunction assigns a callback function that will be called after each
// website crawl.
//
//	options.SetCallbackFunction(func(mapper *SiteMapper) {
//			sitemapURL := "https://example.com/sitemap.xml"
//			googlePingURL := "https://www.google.com/ping?sitemap=" + url.QueryEscape(sitemapURL)
//
//			resp, err := http.Get(googlePingURL)
//			if err != nil {
//				fmt.Println("Error sending request:", err)
//				return
//			}
//			defer resp.Body.Close()
//
//			fmt.Println("Google Sitemap Ping Response:", resp.Status)
//	})
func (options *SiteMapperOptions) SetCallbackFunction(callback func(*SiteMapper)) {
	options.callbackFunc = func(mapper *SiteMapper) {
		if callback != nil {
			callback(mapper)
		}
	}
}
