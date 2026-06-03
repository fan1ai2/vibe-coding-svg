package ai

import (
	"encoding/xml"
	"strings"
)

// ValidateCandidate checks a single SVG for basic correctness.
// Returns true if the SVG passes all checks.
func ValidateCandidate(svgContent string) bool {
	if len(svgContent) > 50*1024 {
		return false
	}

	decoder := xml.NewDecoder(strings.NewReader(svgContent))
	hasSvg := false
	hasViewBox := false

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		el, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if el.Name.Local == "svg" {
			hasSvg = true
			for _, attr := range el.Attr {
				if attr.Name.Local == "viewBox" && attr.Value != "" {
					hasViewBox = true
					break
				}
			}
			break
		}
	}

	return hasSvg && hasViewBox
}

// FilterCandidates validates each candidate and returns only valid ones.
func FilterCandidates(candidates []IconCandidate) []IconCandidate {
	result := make([]IconCandidate, 0, len(candidates))
	for _, c := range candidates {
		if ValidateCandidate(c.SvgContent) {
			result = append(result, c)
		}
	}
	return result
}

// DetectStyleMismatch performs lightweight check: if user requested "line"
// but most paths have fill (not "none"), flag it.
func DetectStyleMismatch(svgContent, requestedStyle string) bool {
	if requestedStyle != "line" {
		return false
	}
	decoder := xml.NewDecoder(strings.NewReader(svgContent))
	fillCount := 0
	noneCount := 0
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
			if attr.Name.Local == "fill" {
				fillCount++
				if attr.Value == "none" {
					noneCount++
				}
			}
		}
	}
	if fillCount == 0 {
		return false
	}
	return noneCount < fillCount/2
}
