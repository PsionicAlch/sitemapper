package sitemapper

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"
)

func TestSiteMapperCrawling(t *testing.T) {
	mockServer := httptest.NewServer(createMockServer())
	defer mockServer.Close()

	options := DefaultOptions()

	if err := options.SetDomain(mockServer.URL); err != nil {
		t.Error(err)
	}

	if err := options.SetDurationBeforeFirstCrawl(0); err != nil {
		t.Error(err)
	}

	if err := options.SetCrawlInterval(time.Minute * 10); err != nil {
		t.Error(err)
	}

	if err := options.SetLinkAttributes("hx-get"); err != nil {
		t.Error(err)
	}

	options.SetInfoLogger(func(s string) {
		t.Log("INFO: ", s)
	})

	options.SetErrorLogger(func(err error) {
		t.Log("ERROR: ", err.Error())
	})

	mapper := NewSiteMapper(options)

	time.Sleep(time.Second * 1)

	linksFound := []string{}
	linksRequired := []string{mockServer.URL + "/page1", mockServer.URL + "/page2", mockServer.URL, mockServer.URL + "/htmx"}

	for _, link := range mapper.spider.getLinks() {
		linksFound = append(linksFound, link.link)
	}

	for _, link := range linksFound {
		if !slices.Contains(linksRequired, link) {
			t.Errorf("Crawler found %s when it should not have", link)
		}
	}

	for _, link := range linksRequired {
		if !slices.Contains(linksFound, link) {
			t.Errorf("Crawler did not find %s when it should have", link)
		}
	}

	if slices.Contains(linksFound, "https://example.com") {
		t.Error("Crawler found 'https://example.com' when it should not have")
	}

	mapper.RecrawlSite()

	time.Sleep(time.Second * 1)

	linksFound = []string{}
	linksRequired = []string{mockServer.URL + "/page1", mockServer.URL + "/page2", mockServer.URL, mockServer.URL + "/htmx"}

	for _, link := range mapper.spider.getLinks() {
		linksFound = append(linksFound, link.link)
	}

	for _, link := range linksFound {
		if !slices.Contains(linksRequired, link) {
			t.Errorf("Crawler found %s when it should not have", link)
		}
	}

	for _, link := range linksRequired {
		if !slices.Contains(linksFound, link) {
			t.Errorf("Crawler did not find %s when it should have", link)
		}
	}

	if slices.Contains(linksFound, "https://example.com") {
		t.Error("Crawler found 'https://example.com' when it should not have")
	}
}

func TestSiteMapperSitemapGeneration(t *testing.T) {
	mockServer := httptest.NewServer(createMockServer())
	defer mockServer.Close()

	options := DefaultOptions()

	if err := options.SetDomain(mockServer.URL); err != nil {
		t.Error(err)
	}

	if err := options.SetDurationBeforeFirstCrawl(0); err != nil {
		t.Error(err)
	}

	if err := options.SetCrawlInterval(time.Minute * 10); err != nil {
		t.Error(err)
	}

	if err := options.SetLinkAttributes("hx-get"); err != nil {
		t.Error(err)
	}

	mapper := NewSiteMapper(options)

	time.Sleep(time.Second * 1)

	sitemap, err := mapper.GenerateSitemap(mockServer.URL, "/htmx")
	if err != nil {
		t.Error(err)
	}

	if sitemap == mapper.EmptySitemapXML(mockServer.URL) {
		t.Error("Failed to generate proper sitemap")
	}

	sitemapURLs, err := extractURLsFromSitemap(sitemap)
	if err != nil {
		t.Errorf("Failed to extract urls from sitemap: %s", err)
	}

	requiredURLs := []string{
		mockServer.URL,
		mockServer.URL + "/page1",
		mockServer.URL + "/page2",
	}

	for _, url := range sitemapURLs {
		if !slices.Contains(requiredURLs, url) {
			t.Errorf("Found '%s' but should not have", url)
		}
	}

	for _, url := range requiredURLs {
		if !slices.Contains(sitemapURLs, url) {
			t.Errorf("Expected to find '%s' but could not", url)
		}
	}
}

func createMockServer() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		resp := `
		<html>
			<head><title>Test Site</title></head>
			<body>
				<a href="/page1">Page 1</a>
				<a href="/page2">Page 2</a>
				<a href="/nonexistent-link">Should Not Be Found</a>
				<a href="/redirect-url">Redirect</a>
				<a href="https://example.com">Another Site</a>
			</body>
		</html>
		`

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(resp))
	})

	mux.HandleFunc("GET /page1", func(w http.ResponseWriter, r *http.Request) {
		resp := `
		<html>
			<head><title>Page 1</title></head>
			<body>
				<div hx-get="/htmx">HTMX Secret</div>
			</body>
		</html>
		`

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(resp))
	})

	mux.HandleFunc("GET /page2", func(w http.ResponseWriter, r *http.Request) {
		resp := `
		<html>
			<head><title>Page 2</title></head>
			<body>
				<div hx-get="/htmx">HTMX Secret</div>
			</body>
		</html>
		`

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(resp))
	})

	mux.HandleFunc("GET /redirect-url", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	})

	mux.HandleFunc("GET /htmx", func(w http.ResponseWriter, r *http.Request) {
		resp := "<h1>I am an HTMX element</h1>"

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(resp))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	return mux
}

func extractURLsFromSitemap(sitemap string) ([]string, error) {
	var urlSet struct {
		XMLName xml.Name `xml:"urlset"`
		Urls    []struct {
			Loc string `xml:"loc"`
		} `xml:"url"`
	}

	if err := xml.Unmarshal([]byte(sitemap), &urlSet); err != nil {
		return nil, err
	}

	urls := make([]string, 0, len(urlSet.Urls))
	for _, url := range urlSet.Urls {
		urls = append(urls, url.Loc)
	}

	return urls, nil
}
