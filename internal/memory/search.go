package memory

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var (
	asciiWordsRe = regexp.MustCompile(`[A-Za-z0-9_]+`)
	hanCharsRe   = regexp.MustCompile(`[\x{4e00}-\x{9fff}\x{3400}-\x{4dbf}]`)
)

// Find returns memory headers whose metadata and content overlap with the query.
// It searches both personal and project memory layers.
// Metadata matches are weighted 2x over body matches.
func Find(query, cwd string, maxResults int) ([]*Header, error) {
	tokens := tokenize(query)
	if len(tokens) == 0 {
		return nil, nil
	}

	// Scan both layers. A failure in one layer does not prevent searching the other.
	var allHeaders []*Header

	projectHeaders, pErr := Scan(cwd, 100)
	if pErr == nil {
		for _, h := range projectHeaders {
			h.Layer = "project"
		}
		allHeaders = append(allHeaders, projectHeaders...)
	}

	personalHeaders, prErr := ScanPersonal(100)
	if prErr == nil {
		for _, h := range personalHeaders {
			h.Layer = "personal"
		}
		allHeaders = append(allHeaders, personalHeaders...)
	}

	if pErr != nil && prErr != nil {
		return nil, fmt.Errorf("scan both layers failed: project: %v, personal: %v", pErr, prErr)
	}

	type scored struct {
		score  float64
		header *Header
	}

	var results []scored
	for _, h := range allHeaders {
		meta := strings.ToLower(h.Title + " " + h.Description)
		body := strings.ToLower(h.BodyPreview)

		var metaHits, bodyHits float64
		for t := range tokens {
			if strings.Contains(meta, t) {
				metaHits++
			}
			if strings.Contains(body, t) {
				bodyHits++
			}
		}

		score := metaHits*2.0 + bodyHits
		if score > 0 {
			results = append(results, scored{score, h})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].score != results[j].score {
			return results[i].score > results[j].score
		}
		return results[i].header.ModifiedAt.After(results[j].header.ModifiedAt)
	})

	if maxResults <= 0 {
		maxResults = 5
	}
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	out := make([]*Header, len(results))
	for i, r := range results {
		out[i] = r.header
	}
	return out, nil
}

// ScanPersonal reads memory files from the personal memory directory.
func ScanPersonal(maxFiles int) ([]*Header, error) {
	dir, err := PersonalDir()
	if err != nil {
		return nil, err
	}
	return scanDir(dir, maxFiles)
}

// tokenize extracts search tokens from text, handling ASCII and Han ideographs.
func tokenize(text string) map[string]bool {
	tokens := make(map[string]bool)

	for _, m := range asciiWordsRe.FindAllString(text, -1) {
		lower := strings.ToLower(m)
		if len(lower) >= 3 {
			tokens[lower] = true
		}
	}

	for _, ch := range hanCharsRe.FindAllString(text, -1) {
		tokens[ch] = true
	}

	return tokens
}
