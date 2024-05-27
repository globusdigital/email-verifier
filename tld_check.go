package emailverifier

import "strings"

// TopLevelDomainExists checks if the TLD exists.
func TopLevelDomainExists(email string) bool {
	email = strings.ToLower(email)
	lastDot := strings.LastIndex(email, ".")
	if lastDot == -1 {
		return false
	}
	tld := email[lastDot+1:]
	_, ok1 := GenericTLDs[tld]
	_, ok2 := CountryCodeTLDs[tld]
	return ok1 || ok2
}
