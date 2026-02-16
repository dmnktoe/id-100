package seo

import (
"encoding/xml"
"fmt"
"time"
)

// URL represents a single URL entry in the sitemap
type URL struct {
Loc        string  `xml:"loc"`
LastMod    string  `xml:"lastmod,omitempty"`
ChangeFreq string  `xml:"changefreq,omitempty"`
Priority   float64 `xml:"priority,omitempty"`
}

// URLSet represents the root element of the sitemap
type URLSet struct {
XMLName xml.Name `xml:"urlset"`
Xmlns   string   `xml:"xmlns,attr"`
URLs    []URL    `xml:"url"`
}

// GenerateSitemap generates a sitemap.xml for the application
func GenerateSitemap(baseURL string, idNumbers []int) ([]byte, error) {
config := GetConfig()
urlSet := URLSet{
Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
URLs:  []URL{},
}

currentDate := time.Now().Format("2006-01-02")

// Add static pages
for _, pageName := range config.GetStaticPages() {
page := config.Pages[pageName]
priority := 0.8
if pageName == "home" {
priority = 1.0
}

urlSet.URLs = append(urlSet.URLs, URL{
Loc:        baseURL + page.Path,
LastMod:    currentDate,
ChangeFreq: "weekly",
Priority:   priority,
})
}

// Add ID pages
for _, idNumber := range idNumbers {
urlSet.URLs = append(urlSet.URLs, URL{
Loc:        fmt.Sprintf("%s/id/%d", baseURL, idNumber),
LastMod:    currentDate,
ChangeFreq: "monthly",
Priority:   0.6,
})
}

// Marshal to XML with proper formatting
output, err := xml.MarshalIndent(urlSet, "", "  ")
if err != nil {
return nil, err
}

// Add XML header
xmlHeader := []byte(xml.Header)
return append(xmlHeader, output...), nil
}
