package sanitizer

import (
	"regexp"
	"strings"
	"unicode"
)

// scriptTagRegex removes full <script>...</script> blocks including enclosed javascript text.
var scriptTagRegex = regexp.MustCompile(`(?is)<script.*?>.*?</script>`)

// styleTagRegex removes full <style>...</style> blocks including enclosed css text.
var styleTagRegex = regexp.MustCompile(`(?is)<style.*?>.*?</style>`)

// htmlTagRegex matches generic HTML tags to prevent Stored XSS attacks in text inputs.
var htmlTagRegex = regexp.MustCompile(`(?i)<[^>]*>`)

// NormalizeEmail cleans and lowercases email strings.
// Why: Standardizes email representations across registration, login, and recovery lookups
// to prevent duplicate account creation with varying case and avoid authentication lookup failures.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// SanitizeName cleans display names by stripping HTML/script tags, removing non-printable control characters, and trimming whitespace.
// Why: Mitigates Stored Cross-Site Scripting (XSS) and unicode corruption risks while preserving legitimate internationalized name characters.
func SanitizeName(name string) string {
	// Strip entire script and style blocks first so script payloads are completely removed
	stripped := scriptTagRegex.ReplaceAllString(name, "")
	stripped = styleTagRegex.ReplaceAllString(stripped, "")

	// Strip any remaining standalone HTML tags
	stripped = htmlTagRegex.ReplaceAllString(stripped, "")

	// Remove non-printable control characters that could break terminal logs, UI renderers, or DB encodings
	var builder strings.Builder
	builder.Grow(len(stripped))
	for _, r := range stripped {
		if !unicode.IsControl(r) {
			builder.WriteRune(r)
		}
	}

	return strings.TrimSpace(builder.String())
}

// SanitizeCode removes surrounding whitespace and non-printable characters from verification tokens/OTPs.
// Why: Prevents OTP verification failures caused by user copy-paste artifacts like invisible control codes or whitespace.
func SanitizeCode(code string) string {
	trimmed := strings.TrimSpace(code)

	var builder strings.Builder
	builder.Grow(len(trimmed))
	for _, r := range trimmed {
		if !unicode.IsControl(r) {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}
