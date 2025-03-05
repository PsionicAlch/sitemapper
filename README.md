# SiteMapper

SiteMapper is a Go library designed to periodically crawl your website to find all public links available and then generate a working sitemap of your website. SiteMapper can be used as part of a monolithic application, a micro service or as a standalone application, granted that you will need to do some wrapping in cases where you plan to use it as a micro services or a standalone application.

SiteMapper was originally created for, and during the development of, [PsionicAlch](https://www.psionicalch.com) as a way of automatically generating a correct and viable sitemap without the need to hardcode any of the URLs.

## Install SiteMapper:

```bash
go get -u "github.com/PsionicAlch/sitemapper"
```

## How to get SiteMapper up and running:

Before you can create an instance of SiteMapper you'll first need to create a configuration for how you want it to work.

```golang
// You can either create a new instance of the SiteMapperOptions struct
mapperOptions := new(sitemapper.SiteMapperOptions)

// or you can use the DefaultOptions function that will construct a new instance of
// SiteMapperOptions for you as well as set some defaults. The defaults are as follows:
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
// - Logging functions are nil by default and can be set later.
//
// - Callback function is empty by default and can be set later.
mapperOptions := sitemapper.DefaultOptions()
```

Once you have an instance of SiteMapperOptions you can use the setters to change any of the options you might want to change for your specific use case. All of the setters, except SetInfoLogger and SetErrorLogger, do some validation and as such will return an error if the data you gave it was invalid.

```golang
// You can set the domain that the crawler will use for it's HTTP GET requests. This is not
// the domain that will be used when generating the sitemap. So if you're using SiteMapper
// as part of a monolithic application that will be behind a reverse proxy you can just use
// the localhost URL of your application.
if err := mapperOptions.SetDomain("http://localhost:8000"); err != nil {
    // Handle error...
}

// If you want SiteMapper to wait a moment before it's initial crawl you can pass any
// non-negative duration. If you want it to start immediately you can just pass 0.
if err := mapperOptions.SetDurationBeforeFirstCrawl(time.Second * 5); err != nil {
    // Handle error...
}

// You can set how often you want SiteMapper to recrawl your site. In this example we
// set it to crawl the website once a week. You can still manually ask it to recrawl
// the site in case any of the data has changed.
if err := mapperOptions.SetCrawlInterval(time.Hour * 24 * 7); err != nil {
    // Handle error...
}

// SiteMapper will start with just one URL and then crawl the site based off any other
// URLs it finds on that first page.
if err := mapperOptions.SetStartingURL("/"); err != nil {
    // Handle error...
}

// SiteMapper by default will crawl any URLs it finds inside of anchor tags but if
// your site uses libraries like HTMX you can tell SiteMapper to also look inside of
// the accompanying HTML attributes like hx-get for HTMX.
if err := mapperOptions.SetLinkAttributes("hx-get"); err != nil {
    // Handle error...
}

// If you want to receive the information logs that come with SiteMapper you can give
// it a mapping function that will be called whenever it needs to log some information.
// If you don't care about logging you can just pass it nil.
mapperOptions.SetInfoLogger(func (msg string) {
    // Log the info message.
})

// If you want to receive the error logs that come with SiteMapper you can give
// it a mapping function that will be called whenever it needs to log some errors.
// If you don't care about logging you can just pass it nil.
mapperOptions.SetErrorLogger(func (err error) {
    // Log the error message.
})

// If you want to run some custom logic after each crawl you can set a callback function.
// The callback function takes one argument, a pointer to SiteMapper. This allows you to
// have full access to the SiteMapper functionality.
mapperOptions.SetCallbackFunction(func (mapper *SiteMapper) {
    // Define your sitemap URL
	sitemapURL := "https://example.com/sitemap.xml"

	// Construct the Google ping URL
	googlePingURL := "https://www.google.com/ping?sitemap=" + url.QueryEscape(sitemapURL)

	// Send the GET request
	resp, err := http.Get(googlePingURL)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Print response status
	fmt.Println("Google Sitemap Ping Response:", resp.Status)
})
```

Once you have all of the options set up, you can create a new instance of SiteMapper.

```golang
// SiteMapper will create a separate goroutine for the crawler so that it doesn't block
// the main application. You will need to hold onto the instance of *SiteMapper so that
// you can request a recrawl or generate the sitemap.
mapper := sitemapper.NewSiteMapper(mapperOptions)
```

If you want to recrawl your website outside of the normal crawl interval it's as easy as calling RecrawlSite:

```golang
// RecrawlSite will use an internal channel to tell the goroutine to recrawl your website.
mapper.RecrawlSite()
```

Once you need to access the sitemap it's as easy as calling GenerateSitemap():

```golang
// GenerateSitemap takes two arguments:
// - baseDomain:    This is the domain name of your website. SiteMapper will make sure
//                  to replace the domain name that you passed earlier, when creating
//                  an instance of SiteMapper, with this domain so that your sitemap
//                  is based off your actual domain.
// - filterPattern: This is a regex pattern that tells SiteMapper which URLs to not include
//                  when generating the sitemap. A use case for this might be the HTMX specific
//                  URLs which were needed to be mapped but are not needed in the sitemap.
sitemap, err := mapper.GenerateSitemap("http://example.com", "/htmx")
if err != nil {
    // If an error does occur an empty sitemap will be returned. The empty sitemap is
    // a valid sitemap that only points to the home page of your website.
}
```

## License

This project is licensed under the MIT License. See the [LICENSE](https://github.com/PsionicAlch/SiteMapper/blob/main/LICENSE) file for details.

## Acknowledgments

This project uses the following open-source libraries:

[Go net/html](https://golang.org/x/net) for HTML parsing.
