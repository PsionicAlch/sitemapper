package sitemapper

import "time"

// SiteMapper is responsible for managing the crawling and sitemap generation for a specified domain.
// It schedules periodic crawls and allows manual recrawling.
//
// SiteMapper uses a crawler to traverse the site and build the sitemap based on the configuration options.
type SiteMapper struct {
	// spider is the internal crawler instance responsible for the actual crawling process.
	spider *crawler

	// recrawlSignal is a channel used to trigger manual recrawling.
	recrawlSignal chan bool

	// domain is the domain name of the site being crawled.
	domain string
}

// NewSiteMapper initializes and returns a new SiteMapper instance configured with the provided options.
//
// The SiteMapper starts its first crawl after the delay specified in `options.durationBeforeFirstCrawl`
// and subsequently recrawls the site at intervals defined by `options.crawlInterval`.
//
// Parameters:
//
//	options *sitemapper.SiteMapperOptions // Configuration options for the SiteMapper instance.
//
// Returns:
//
//	*sitemapper.SiteMapper // A new SiteMapper instance.
func NewSiteMapper(options *SiteMapperOptions) *SiteMapper {
	mapper := &SiteMapper{
		spider:        newCrawler(options.domain, options.linkAttributes, options.infoLogger, options.errorLogger),
		recrawlSignal: make(chan bool),
		domain:        options.domain,
	}

	// Start the crawling process in a separate goroutine.
	go func() {
		if options.durationBeforeFirstCrawl > 0 {
			// Wait for the initial delay before the first crawl.
			time.Sleep(options.durationBeforeFirstCrawl)
		}

		// Perform the first crawl.
		mapper.spider.crawl(options.startingURL)

		// Schedule periodic crawls using a ticker.
		ticker := time.NewTicker(options.crawlInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Perform a scheduled crawl.
				mapper.spider.crawl(options.startingURL)
				options.callbackFunc(mapper)
			case <-mapper.recrawlSignal:
				// Perform a manual recrawl triggered by the RecrawlSite method.
				mapper.spider.crawl(options.startingURL)
				options.callbackFunc(mapper)
			}
		}
	}()

	return mapper
}

// RecrawlSite triggers a manual recrawl of the site, bypassing the scheduled interval.
func (mapper *SiteMapper) RecrawlSite() {
	mapper.recrawlSignal <- true
}
