package main

import (
	"html"
	"strings"
)

//ConvertHTMLEntities converts scraped html entities and returns clean string
func ConvertHTMLEntities(htmlString string) (clean string) {
	clean = html.UnescapeString(htmlString)

	replacer := strings.NewReplacer(
		//set with pairs
		"\u0026", "&",
		"&amp;", "&",
	)

	clean = replacer.Replace(clean)
	return
}
