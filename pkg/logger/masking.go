package logger

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"
)

/*
MaskType definitions
*/
type MaskType string

const (
	MaskFull       MaskType = "full"
	MaskEmail      MaskType = "email"
	MaskPhone      MaskType = "phone"
	MaskFirst      MaskType = "first"
	MaskLast       MaskType = "last"
	MaskPartial    MaskType = "partial"
	MaskHash       MaskType = "hash"
	MaskIDCard     MaskType = "id_card"
	MaskCreditCard MaskType = "credit_card"
)

type Masker interface {
	Mask(value string, rule MaskRule) string
}

type masker struct {
	pattern string
}

/*
MaskRule structure
*/
type MaskRule struct {
	Field   string
	Type    MaskType
	IsArray bool

	// for partial
	Prefix int
	Suffix int
}

func NewMasker(pattern string) Masker {
	return &masker{
		pattern: pattern,
	}
}

func (m *masker) Mask(value string, rule MaskRule) string {

	if value == "" {
		return value
	}

	switch rule.Type {
	case MaskFull:
		return maskFull(value, m.pattern)

	case MaskEmail:
		return maskEmail(value)

	case MaskPhone:
		return maskPhone(value)

	case MaskFirst:
		return maskFirst(value)

	case MaskLast:
		return maskLast(value)

	case MaskPartial:
		return maskPartial(value, m.pattern, rule.Prefix, rule.Suffix)

	case MaskHash:
		return maskHash(value)

	case MaskIDCard:
		return maskIDCard(value, m.pattern)

	case MaskCreditCard:
		return maskCreditCard(value, m.pattern)

	default:
		return value
	}

}

/*
Implementations
*/

func maskFull(value, pattern string) string {
	if pattern != "" {
		return pattern
	}
	return strings.Repeat("*", len(value))
}

func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return maskFull(email, "")
	}

	name := parts[0]
	domain := parts[1]

	if len(name) <= 2 {
		return "**@" + domain
	}

	return name[:2] +
		strings.Repeat("*", len(name)-2) +
		"@" + domain
}

func maskPhone(phone string) string {
	if len(phone) < 7 {
		return maskFull(phone, "")
	}

	return phone[:3] +
		strings.Repeat("*", len(phone)-6) +
		phone[len(phone)-3:]
}

func maskFirst(value string) string {
	if len(value) <= 1 {
		return "*"
	}
	return value[:1] + strings.Repeat("*", len(value)-1)
}

func maskLast(value string) string {
	if len(value) <= 1 {
		return "*"
	}
	return strings.Repeat("*", len(value)-1) + value[len(value)-1:]
}

func maskPartial(value, pattern string, prefix, suffix int) string {
	l := len(value)

	if prefix+suffix >= l {
		return strings.Repeat("*", l)
	}

	maskChar := pattern
	if maskChar == "" {
		maskChar = "*"
	}

	return value[:prefix] + strings.Repeat(maskChar, l-prefix-suffix) + value[l-suffix:]
}

func maskHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:8]) // short hash for log
}

func maskIDCard(value, pattern string) string {
	digits := extractDigits(value)

	if len(digits) < 6 {
		return maskFull(value, pattern)
	}

	first := string(digits[0])
	last := string(digits[len(digits)-2:])

	return first +
		strings.Repeat(pattern, len(digits)-3) +
		last
}

func maskCreditCard(value, pattern string) string {
	digits := extractDigits(value)

	if len(digits) < 12 {
		return maskFull(value, pattern)
	}

	last4 := string(digits[len(digits)-4:])
	masked := strings.Repeat(pattern, len(digits)-4) + last4

	var result strings.Builder
	idx := 0

	for _, r := range value {
		if unicode.IsDigit(r) {
			result.WriteRune(rune(masked[idx]))
			idx++
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

/*
Helpers
*/

func extractDigits(value string) []rune {
	digits := make([]rune, 0, len(value))
	for _, r := range value {
		if unicode.IsDigit(r) {
			digits = append(digits, r)
		}
	}
	return digits
}

/*
Optional: regex helper if needed later
*/
func maskRegex(value, pattern string) string {
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(value, "*")
}
