package extractor

import (
	"strings"
	"unicode"
)

// commonWords are everyday English tokens that carry no architectural meaning
// on their own. A multi-word name made ONLY of these is a phrase fragment
// (e.g. "map it step"), not a component.
var commonWords = map[string]bool{
	"a": true, "about": true, "after": true, "all": true, "also": true,
	"an": true, "and": true, "any": true, "are": true, "as": true, "at": true,
	"be": true, "before": true, "but": true, "by": true, "can": true,
	"check": true, "close": true, "core": true, "do": true, "done": true,
	"each": true, "end": true, "every": true, "first": true, "fix": true,
	"for": true, "from": true, "get": true, "has": true, "have": true,
	"here": true, "how": true, "if": true, "in": true, "into": true,
	"is": true, "it": true, "its": true, "last": true, "like": true,
	"list": true, "loop": true, "make": true, "map": true, "more": true,
	"most": true, "must": true, "new": true, "next": true, "no": true,
	"not": true, "now": true, "of": true, "off": true, "on": true,
	"one": true, "only": true, "or": true, "order": true, "other": true,
	"our": true, "out": true, "over": true, "part": true, "phase": true,
	"plan": true, "run": true, "same": true, "see": true, "set": true,
	"should": true, "show": true, "so": true, "some": true, "start": true,
	"step": true, "task": true, "test": true, "that": true, "the": true,
	"then": true, "this": true, "time": true, "to": true, "up": true,
	"use": true, "used": true, "using": true, "was": true, "we": true,
	"were": true, "what": true, "when": true, "where": true, "which": true,
	"will": true, "with": true, "work": true, "you": true, "your": true,
	// trunk branch names — never architectural entities on their own
	"main": true, "master": true, "develop": true, "branch": true,
}

// IsNoise reports whether an extracted entity name is junk that should never
// enter the graph: too short, purely numeric, malformed edges, or a phrase
// made only of common English words ("map it step"). Names with code-like
// separators (dots, dashes) and at least one uncommon token pass.
func IsNoise(name string) bool {
	name = strings.TrimSpace(name)
	if len(name) < 3 {
		return true
	}
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") ||
		strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") {
		return true
	}

	allDigits := true
	for _, r := range name {
		if !unicode.IsDigit(r) {
			allDigits = false
			break
		}
	}
	if allDigits {
		return true
	}

	// Phrase fragments: multi-word names where every token is a common word
	if strings.Contains(name, " ") {
		allCommon := true
		for _, tok := range strings.Fields(strings.ToLower(name)) {
			if !commonWords[tok] {
				allCommon = false
				break
			}
		}
		if allCommon {
			return true
		}
	} else if commonWords[strings.ToLower(name)] {
		// A bare common word ("master", "step", "map") is never a component
		return true
	}

	return false
}
