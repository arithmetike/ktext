package schema

import "regexp"

// qualityRule is a check applied to a text field.
type qualityRule struct {
	check   func(string) bool
	message string
}

var (
	reActionable = regexp.MustCompile(`(?i)\b(prohibit|require|never|always|must|shall|do not|don't|avoid|ensure|only|exclusively)\b`)
	reActionVerb = regexp.MustCompile(`(?i)\b(use|run|write|import|export|call|return|throw|log|test|deploy|format|name|prefix|suffix|wrap|extend|create|delete|update|add|remove|keep|put|set|check|validate|handle)\b`)
)

func wordCount(s string) int {
	n := 0
	inWord := false
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' {
			inWord = false
		} else if !inWord {
			inWord = true
			n++
		}
	}
	return n
}

var qualityRules = map[string][]qualityRule{
	"identity.purpose": {
		{func(v string) bool { return wordCount(v) >= 3 }, "expand — purpose needs at least 3 words"},
	},
	"constraints": {
		{func(v string) bool { return wordCount(v) >= 3 }, "expand — at least 3 words"},
		{func(v string) bool { return reActionable.MatchString(v) }, "add actionable language: must, never, always, require, etc."},
	},
	"decisions.rationale": {
		{func(v string) bool { return wordCount(v) >= 3 }, "expand the rationale — at least 3 words"},
	},
	"conventions": {
		{func(v string) bool { return wordCount(v) >= 3 }, "expand — at least 3 words"},
		{func(v string) bool { return reActionVerb.MatchString(v) }, "add an action verb: use, write, import, format, etc."},
	},
	"risks": {
		{func(v string) bool { return wordCount(v) >= 3 }, "describe the specific exposure, not just the category"},
	},
}

// CheckQuality returns the messages for failed quality rules for section.
func CheckQuality(section, value string) []string {
	rules := qualityRules[section]
	var out []string
	for _, r := range rules {
		if !r.check(value) {
			out = append(out, r.message)
		}
	}
	return out
}
