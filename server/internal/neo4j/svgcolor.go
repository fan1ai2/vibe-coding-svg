package neo4j

import (
	"encoding/xml"
	"sort"
	"strings"
)

type ColorFreq struct {
	Color string
	Count int
}

// ExtractColors parses SVG content and extracts fill/stroke color values,
// filtering out none, currentColor, url(#...), transparent, and empty strings.
// Returns up to n most frequent colors.
func ExtractColors(svgContent string, n int) []string {
	decoder := xml.NewDecoder(strings.NewReader(svgContent))
	freq := make(map[string]int)

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		el, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		for _, attr := range el.Attr {
			if attr.Name.Local == "fill" || attr.Name.Local == "stroke" {
				c := strings.TrimSpace(attr.Value)
				if isValidColor(c) {
					freq[normalizeColor(c)]++
				}
			}
		}
	}

	if len(freq) == 0 {
		return nil
	}

	list := make([]ColorFreq, 0, len(freq))
	for color, count := range freq {
		list = append(list, ColorFreq{Color: color, Count: count})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Count > list[j].Count })

	if n > len(list) {
		n = len(list)
	}
	result := make([]string, n)
	for i := 0; i < n; i++ {
		result[i] = list[i].Color
	}
	return result
}

func isValidColor(c string) bool {
	if c == "" || c == "none" || c == "currentColor" || c == "transparent" {
		return false
	}
	if strings.HasPrefix(c, "url(") {
		return false
	}
	if strings.HasPrefix(c, "var(") {
		return false
	}
	return true
}

func normalizeColor(c string) string {
	c = strings.ToLower(c)
	if len(c) == 4 && c[0] == '#' {
		return "#" + string([]byte{c[1], c[1], c[2], c[2], c[3], c[3]})
	}
	return c
}
