package emailverifier

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const regexInvalid = "email not valid, regex"

func TestCheckAddressSyntax(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  Syntax
	}{
		// Valid addresses.
		{"simple", "example@domain.com", Syntax{Username: "example", Domain: "domain.com", Valid: true}},
		{"yahoo", "support@yahoo.com", Syntax{Username: "support", Domain: "yahoo.com", Valid: true}},
		{"digits in domain", "tool@163.com", Syntax{Username: "tool", Domain: "163.com", Valid: true}},
		{"digit inside domain label", "user@gma3il.com", Syntax{Username: "user", Domain: "gma3il.com", Valid: true}},
		{"underscore in local part", "a_b@github.com", Syntax{Username: "a_b", Domain: "github.com", Valid: true}},
		{"dot in local part", "first.last@example.com", Syntax{Username: "first.last", Domain: "example.com", Valid: true}},
		{"plus tag", "user+tag@gmail.com", Syntax{Username: "user+tag", Domain: "gmail.com", Valid: true}},
		{"apostrophe", "o'reilly@example.com", Syntax{Username: "o'reilly", Domain: "example.com", Valid: true}},
		{"numeric local part", "1234567890@example.com", Syntax{Username: "1234567890", Domain: "example.com", Valid: true}},
		{"all RFC 5322 atext specials", "!#$%&'*+-/=?^_`{|}~@example.com", Syntax{Username: "!#$%&'*+-/=?^_`{|}~", Domain: "example.com", Valid: true}},
		{"single-letter labels", "a@b.co", Syntax{Username: "a", Domain: "b.co", Valid: true}},
		{"subdomains", "user@mail.sub.example.co.uk", Syntax{Username: "user", Domain: "mail.sub.example.co.uk", Valid: true}},
		{"hyphen in domain", "user@my-domain.com", Syntax{Username: "user", Domain: "my-domain.com", Valid: true}},
		{"punycode domain", "user@xn--bcher-kva.ch", Syntax{Username: "user", Domain: "xn--bcher-kva.ch", Valid: true}},
		{"domain lowercased, username kept", "Example.User@Domain.COM", Syntax{Username: "Example.User", Domain: "domain.com", Valid: true}},
		{"IDN domain", "abc@доменное.com", Syntax{Username: "abc", Domain: "доменное.com", Valid: true}},
		{"IDN domain lowercased", "ПОЧТА@ДОМЕН.РФ", Syntax{Username: "ПОЧТА", Domain: "домен.рф", Valid: true}},
		{"non-ASCII local part", "ü@example.com", Syntax{Username: "ü", Domain: "example.com", Valid: true}},
		{"quoted local part with space", `"john doe"@example.com`, Syntax{Username: `"john doe"`, Domain: "example.com", Valid: true}},
		{"quoted local part with @ splits on last @", `"a@b"@example.com`, Syntax{Username: `"a@b"`, Domain: "example.com", Valid: true}},
		{"trailing dot (FQDN) is trimmed", "user@example.com.", Syntax{Username: "user", Domain: "example.com", Valid: true}},
		{"trailing dot trimmed after lowercasing", "User@Example.COM.", Syntax{Username: "User", Domain: "example.com", Valid: true}},
		{"double trailing dot", "user@example.com..", Syntax{InvalidReason: regexInvalid}},

		// Invalid addresses.
		{"empty", "", Syntax{InvalidReason: regexInvalid}},
		{"no @", "plainaddress", Syntax{InvalidReason: regexInvalid}},
		{"no local part", "@example.com", Syntax{InvalidReason: regexInvalid}},
		{"no domain", "user@", Syntax{InvalidReason: regexInvalid}},
		{"single-label domain", "user@localhost", Syntax{InvalidReason: regexInvalid}},
		{"two unquoted @", "a@b@example.com", Syntax{InvalidReason: regexInvalid}},
		{"leading space", " jerry@gmail.com", Syntax{InvalidReason: regexInvalid}},
		{"trailing space", "jerry@gmail.com ", Syntax{InvalidReason: regexInvalid}},
		{"trailing newline", "jerry@gmail.com\n", Syntax{InvalidReason: regexInvalid}},
		{"unquoted space", "john doe@example.com", Syntax{InvalidReason: regexInvalid}},
		{"leading dot in local part", ".user@example.com", Syntax{InvalidReason: regexInvalid}},
		{"trailing dot in local part", "user.@example.com", Syntax{InvalidReason: regexInvalid}},
		{"consecutive dots in local part", "user..name@example.com", Syntax{InvalidReason: regexInvalid}},
		{"leading dot in domain", "user@.example.com", Syntax{InvalidReason: regexInvalid}},
		{"label starts with hyphen", "user@-example.com", Syntax{InvalidReason: regexInvalid}},
		{"label ends with hyphen", "user@example-.com", Syntax{InvalidReason: regexInvalid}},
		{"underscore in domain", "user@exa_mple.com", Syntax{InvalidReason: regexInvalid}},
		{"numeric TLD", "user@example.123", Syntax{InvalidReason: regexInvalid}},
		{"IP literal", "user@[192.168.0.1]", Syntax{InvalidReason: regexInvalid}},
		{"emoji outside BMP", "😀@gmail.com", Syntax{InvalidReason: regexInvalid}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, verifier.ParseAddress(tt.email), "ParseAddress(%q)", tt.email)
		})
	}
}

func TestCheckAddressSyntax_ASCIIOnly(t *testing.T) {
	v := NewVerifier().EnableASCIIOnlyCheck()

	nonASCII := func(r rune) string {
		return fmt.Sprintf("email not valid as it contains non-ASCII character: %q", r)
	}

	tests := []struct {
		name  string
		email string
		want  Syntax
	}{
		{"ASCII address passes", "example@domain.com", Syntax{Username: "example", Domain: "domain.com", Valid: true}},
		{"domain still lowercased", "User@Example.COM", Syntax{Username: "User", Domain: "example.com", Valid: true}},
		{"punycode domain passes", "user@xn--bcher-kva.ch", Syntax{Username: "user", Domain: "xn--bcher-kva.ch", Valid: true}},
		{"IDN domain rejected", "abc@доменное.com", Syntax{InvalidReason: nonASCII('д')}},
		{"non-ASCII local part rejected", "ü@example.com", Syntax{InvalidReason: nonASCII('ü')}},
		{"reports first non-ASCII rune", "aé@exämple.com", Syntax{InvalidReason: nonASCII('é')}},
		{"emoji rejected before regex", "😀@gmail.com", Syntax{InvalidReason: nonASCII('😀')}},
		{"ASCII but malformed falls through to regex", " jerry@gmail.com", Syntax{InvalidReason: regexInvalid}},
		{"empty falls through to regex", "", Syntax{InvalidReason: regexInvalid}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, v.ParseAddress(tt.email), "ParseAddress(%q)", tt.email)
		})
	}

	t.Run("DisableASCIIOnlyCheck restores the default", func(t *testing.T) {
		v := NewVerifier().EnableASCIIOnlyCheck().DisableASCIIOnlyCheck()
		assert.Equal(t, Syntax{Username: "ü", Domain: "example.com", Valid: true}, v.ParseAddress("ü@example.com"))
	})
}
