package emailverifier

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var emailRegex = regexp.MustCompile(emailRegexString)

// Syntax stores all information about an email Syntax
type Syntax struct {
	Username      string `json:"username"`
	Domain        string `json:"domain"`
	Valid         bool   `json:"valid"`
	InvalidReason string `json:"invalidReason,omitempty"`
}

// ParseAddress attempts to parse an email address and return it in the form of an Syntax
func (v *Verifier) ParseAddress(email string) Syntax {
	if v.asciiOnlyCheckEnabled {
		if i := strings.IndexFunc(email, func(r rune) bool { return r > unicode.MaxASCII }); i >= 0 {
			r, _ := utf8.DecodeRuneInString(email[i:])
			return Syntax{Valid: false, InvalidReason: fmt.Sprintf("email not valid as it contains non-ASCII character: %q", r)}
		}
	}

	isAddressValid := IsAddressValid(email)
	if !isAddressValid {
		return Syntax{Valid: false, InvalidReason: "email not valid, regex"}
	}

	index := strings.LastIndex(email, "@")
	username := email[:index]
	// The regex accepts a trailing dot (fully qualified form, "example.com.");
	// drop it so the domain matches its usual spelling in every later check.
	domain := strings.TrimSuffix(strings.ToLower(email[index+1:]), ".")

	return Syntax{
		Username: username,
		Domain:   domain,
		Valid:    isAddressValid,
	}
}

// IsAddressValid checks if email address is formatted correctly by using regex
func IsAddressValid(email string) bool {
	return emailRegex.MatchString(email)
}
