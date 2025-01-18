package sitemapper

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

// crawlerURL represents a URL with its metadata.
type crawlerURL struct {
	// link is the URL of the page.
	link string

	// checksum is a hash of the page content to detect changes between crawls.
	checksum string

	// lastChanged is timestamp of the last detected change.
	lastChanged time.Time
}

// crawler manages the crawling process within a specific domain.
type crawler struct {
	// mutex ensures thread-safe access to shared resources.
	mutex sync.Mutex

	// domain represents the domain which should be crawled whilst also ensuring
	// that links outside of this domain don't get indexed.
	domain string

	// linkAttributes are the HTML attributes the crawler should consider as links.
	linkAttributes []string

	// visited is the URLs that have been found in the latest crawl.
	visited map[string]crawlerURL

	// links represents all the links that have been discovered throughout the lifespan
	// of the running application. Useful for keeping track of which links have changed.
	links map[string]crawlerURL

	// infoLogger is used for logging informational messages. No messages will be logged
	// if an infoLogger was not passed to SiteMapper.
	infoLogger func(string)

	// errorLogger is used for logging error messages. No messages will be logged
	// if an infoLogger was not passed to SiteMapper.
	errorLogger func(error)
}

// newCrawler creates a new crawler instance.
func newCrawler(domain string, linkAttributes []string, infoLogger func(string), errorLogger func(error)) *crawler {
	return &crawler{
		domain:         domain,
		linkAttributes: linkAttributes,
		visited:        make(map[string]crawlerURL),
		links:          make(map[string]crawlerURL),
		infoLogger:     infoLogger,
		errorLogger:    errorLogger,
	}
}

// crawl starts crawling from the given URL.
func (crawler *crawler) crawl(url string) {
	// Ensure only one goroutine modifies shared state at a time.
	crawler.mutex.Lock()
	defer crawler.mutex.Unlock()

	// Reset the visited map for a new crawl.
	crawler.visited = make(map[string]crawlerURL)

	// Normalize the starting URL.
	normalizedURL, ok := crawler.normalizeURL(url)
	if !ok {
		return
	}

	// Initialize the queue with the starting URL.
	queue := []string{normalizedURL}

	// Process the queue until it's empty.
	for len(queue) > 0 {
		currentURL := queue[0]

		// Dequeue the first URL.
		queue = queue[1:]

		// Skip the URL if it has already been visited.
		if _, has := crawler.visited[currentURL]; has {
			continue
		}

		// Create an HTTP client that will error on redirects.
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return errors.New("redirects not allowed")
			},
		}

		// Fetch the HTML data for the currentURL.
		resp, err := client.Get(currentURL)
		if err != nil {
			crawler.errorLogger(fmt.Errorf("error fetching \"%s\": %w", currentURL, err))
			continue
		}

		// Ensure that the response was successful.
		if resp.StatusCode != http.StatusOK {
			crawler.errorLogger(fmt.Errorf("\"%s\" did not return status code 200: %d", currentURL, resp.StatusCode))
			resp.Body.Close()
			continue
		}

		// Read the body of the response.
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			crawler.errorLogger(fmt.Errorf("error reading response body: %w", err))
			resp.Body.Close()
			continue
		}

		// Since the body has been read into a variable we can safely close it.
		resp.Body.Close()

		// Info log which site we are currently crawling.
		crawler.infoLogger(fmt.Sprintf("Crawling '%s'", currentURL))

		// Extract all the links from the page and add unvisited links to the queue.
		links := crawler.extractLinks(bytes.NewReader(bodyBytes))
		for _, link := range links {
			if _, has := crawler.visited[link]; !has {
				queue = append(queue, link)
			}
		}

		// Compute a hash of the page content for change detection.
		hasher := sha256.New()
		hasher.Write(bodyBytes)

		// Store metadata for the current URL.
		url := crawlerURL{
			link:        currentURL,
			checksum:    hex.EncodeToString(hasher.Sum(nil)),
			lastChanged: time.Now(),
		}

		crawler.visited[currentURL] = url
	}

	// Update the list of known links.
	newLinks := make(map[string]crawlerURL)
	for linkVisited, urlVisited := range crawler.visited {
		if oldUrl, has := crawler.links[linkVisited]; has {
			if urlVisited.checksum != oldUrl.checksum {
				newLinks[linkVisited] = urlVisited
			} else {
				newLinks[linkVisited] = oldUrl
			}
		} else {
			newLinks[linkVisited] = urlVisited
		}
	}

	crawler.links = newLinks
}

// getLinks retrieves all discovered links as a slice of crawlerURL.
func (crawler *crawler) getLinks() []crawlerURL {
	crawler.mutex.Lock()
	defer crawler.mutex.Unlock()

	return slices.Collect(maps.Values(crawler.links))
}

// extractLinks parses HTML content and extracts links based on the specified attributes.
func (crawler *crawler) extractLinks(r io.Reader) []string {
	links := []string{}
	tokenizer := html.NewTokenizer(r)

	for {
		tt := tokenizer.Next()

		switch tt {
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()

			// Handle <a> tags specifically. This is necessary because <link> tags also use
			// href attributes and there is no reason why we'd ever want to crawl a <link>.
			if token.Data == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						link := attr.Val
						if normalized, ok := crawler.normalizeURL(link); ok {
							links = append(links, normalized)
						}
					}
				}
			} else {
				// Handle the other tags based on what the user has configured.
				for _, attr := range token.Attr {
					if slices.Contains(crawler.linkAttributes, attr.Key) {
						link := attr.Val
						if normalized, ok := crawler.normalizeURL(link); ok {
							links = append(links, normalized)
						}
					}
				}
			}
		case html.ErrorToken:
			// End of the document or an error.
			return links
		}
	}
}

// normalizeURL normalizes a URL and ensures it belongs to the specified domain.
func (crawler *crawler) normalizeURL(href string) (string, bool) {
	// Explicitly handle empty strings
	if strings.TrimSpace(href) == "" {
		return "", false
	}

	parsedURL, err := url.Parse(href)
	if err != nil || parsedURL.Scheme == "javascript" {
		return "", false
	}

	// Resolve relative URLs against the base domain.
	if !parsedURL.IsAbs() {
		baseURL, err := url.Parse(crawler.domain)
		if err != nil {
			return "", false
		}
		parsedURL = baseURL.ResolveReference(parsedURL)
	}

	// Remove URL fragments and trailing slashes.
	parsedURL.Fragment = ""
	normalized := strings.TrimRight(parsedURL.String(), "/")

	// Ensure the URL belongs to the specified domain.
	if strings.HasPrefix(normalized, crawler.domain) {
		return normalized, true
	}

	return "", false
}

// ensureTrailingSlash appends a trailing slash to URLs without file extensions or paths.
func ensureTrailingSlash(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	if parsedURL.Path == "" || !strings.HasSuffix(parsedURL.Path, "/") {
		if !strings.Contains(parsedURL.Path, ".") {
			parsedURL.Path += "/"
		}
	}

	return parsedURL.String()
}
