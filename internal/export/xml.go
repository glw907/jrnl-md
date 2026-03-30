package export

import (
	"encoding/xml"
	"fmt"
	"strconv"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/journal"
)

type xmlJournal struct {
	XMLName xml.Name   `xml:"journal"`
	Entries xmlEntries `xml:"entries"`
	Tags    xmlTags    `xml:"tags"`
}

type xmlEntries struct {
	Entries []xmlEntry `xml:"entry"`
}

type xmlEntry struct {
	Date    string   `xml:"date,attr"`
	Starred string   `xml:"starred,attr"`
	Tags    []xmlTag `xml:"tag"`
	Body    string   `xml:",chardata"`
}

type xmlTags struct {
	Tags []xmlTagCount `xml:"tag"`
}

type xmlTag struct {
	Name string `xml:"name,attr"`
}

type xmlTagCount struct {
	Name  string `xml:"name,attr"`
	Count int    `xml:",chardata"`
}

// XML formats entries as an XML document.
func XML(entries []journal.Entry, cfg config.Config) (string, error) {
	var doc xmlJournal

	for _, e := range entries {
		xe := xmlEntry{
			Date:    e.Date.Format("2006-01-02T15:04:05"),
			Starred: strconv.FormatBool(e.Starred),
			Body:    trimBody(e.Body),
		}
		for _, tag := range e.Tags {
			xe.Tags = append(xe.Tags, xmlTag{Name: tag})
		}
		doc.Entries.Entries = append(doc.Entries.Entries, xe)
	}

	counts := TagCounts(entries)
	for _, tag := range sortedTagKeys(counts) {
		doc.Tags.Tags = append(doc.Tags.Tags, xmlTagCount{Name: tag, Count: counts[tag]})
	}

	data, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling XML: %w", err)
	}

	return xml.Header + string(data) + "\n", nil
}
