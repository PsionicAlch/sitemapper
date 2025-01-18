package sitemapper

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type sitemapURL struct {
	XMLName      xml.Name `xml:"url"`
	Location     string   `xml:"loc"`
	LastModified string   `xml:"lastmod,omitempty"`
}

type sitemapURLSet struct {
	XMLName      xml.Name     `xml:"urlset"`
	Xmlns        string       `xml:"xmlns,attr"`
	XmlnsXsi     string       `xml:"xmlns:xsi,attr"`
	XsiSchemaLoc string       `xml:"xsi:schemaLocation,attr"`
	URLS         []sitemapURL `xml:"url"`
}

func (mapper *SiteMapper) GenerateSitemap(baseDomain string, filterPattern string) (string, error) {
	urlSet := sitemapURLSet{
		Xmlns:        "http://www.sitemaps.org/schemas/sitemap/0.9",
		XmlnsXsi:     "http://www.w3.org/2001/XMLSchema-instance",
		XsiSchemaLoc: "http://www.sitemaps.org/schemas/sitemap/0.9 http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd",
	}

	links := mapper.spider.getLinks()

	var urls []sitemapURL

	filter, err := regexp.Compile(filterPattern)
	if err != nil {
		return mapper.EmptySitemapXML(baseDomain), fmt.Errorf("invalid filter pattern provided: %w", err)
	}

	for _, link := range links {
		if filter.MatchString(link.link) {
			continue
		}

		url := sitemapURL{
			Location:     replaceDomain(link.link, mapper.domain, baseDomain),
			LastModified: link.lastChanged.Format("2006-01-02"),
		}

		urls = append(urls, url)
	}

	urlSet.URLS = urls

	xmlBytes, err := xml.MarshalIndent(urlSet, "", "	")
	if err != nil {
		return mapper.EmptySitemapXML(baseDomain), fmt.Errorf("failed to generate xml: %w", err)
	}

	xmlHeader := []byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	return string(append(xmlHeader, xmlBytes...)), nil
}

func (mapper *SiteMapper) EmptySitemapXML(baseDomain string) string {
	emptySiteMap := fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8'?>
		<urlset xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.sitemaps.org/schemas/sitemap/0.9 http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd" xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
			<url>
				<loc>%s/</loc>
				<lastmod>%s</lastmod>
			</url>
		</urlset>
		`, baseDomain, time.Now().Format("2006-01-02"))

	return emptySiteMap
}

func replaceDomain(link, oldDomain, newDomain string) string {
	if strings.HasPrefix(link, oldDomain) {
		return strings.Replace(link, oldDomain, newDomain, 1)
	}

	return link
}
