package ai

import (
	"regexp"
	"strings"
)

// approxTokens returns an approximate token count for a piece of text.
// This is a heuristic (average 4 characters per token) and is used only for budgeting.
func approxTokens(s string) int {
	if s == "" {
		return 0
	}
	// collapse spaces to get better estimate
	normalized := strings.Join(strings.Fields(s), " ")
	return (len(normalized) + 3) / 4
}

var sentenceSplitRE = regexp.MustCompile(`(?m)([^.!?\n]+[.!?\n]?)`)

// chunkTextByTokens splits text into chunks each approximately under maxTokens.
// It splits on sentence boundaries and groups sentences until reaching the token limit.
func chunkTextByTokens(text string, maxTokens int) []string {
	if text == "" {
		return nil
	}
	// quick path
	if approxTokens(text) <= maxTokens {
		return []string{text}
	}

	parts := sentenceSplitRE.FindAllString(text, -1)
	var chunks []string
	var cur strings.Builder
	curTokens := 0

	flush := func() {
		s := strings.TrimSpace(cur.String())
		if s != "" {
			chunks = append(chunks, s)
		}
		cur.Reset()
		curTokens = 0
	}

	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t == "" {
			continue
		}
		tTokens := approxTokens(t)
		// If single sentence bigger than maxTokens, split by words
		if tTokens > maxTokens {
			words := strings.Fields(t)
			var wcur strings.Builder
			wTokens := 0
			for _, w := range words {
				wT := approxTokens(w + " ")
				if wTokens+wT > maxTokens {
					if wcur.Len() > 0 {
						if cur.Len() > 0 {
							cur.WriteString(" ")
						}
						cur.WriteString(strings.TrimSpace(wcur.String()))
						flush()
						wcur.Reset()
						wTokens = 0
					}
				}
				if wcur.Len() > 0 {
					wcur.WriteString(" ")
				}
				wcur.WriteString(w)
				wTokens += wT
			}
			if wcur.Len() > 0 {
				if cur.Len() > 0 {
					cur.WriteString(" ")
				}
				cur.WriteString(strings.TrimSpace(wcur.String()))
				flush()
			}
			continue
		}

		if curTokens+tTokens > maxTokens {
			flush()
		}
		if cur.Len() > 0 {
			cur.WriteString(" ")
		}
		cur.WriteString(t)
		curTokens += tTokens
	}
	// final
	flush()
	return chunks
}
