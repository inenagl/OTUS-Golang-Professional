package hw10programoptimization

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/tidwall/gjson"
)

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	ds, err := countDomains(r, domain)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	return ds, nil
}

func countDomains(r io.Reader, domain string) (DomainStat, error) {
	result := make(DomainStat)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		if !gjson.Valid(line) {
			return nil, fmt.Errorf("invalid json: '%s'", line)
		}
		email := gjson.Get(line, "Email")
		if !email.Exists() {
			continue
		}

		fullDomain, ok := extractDomain(email.String(), domain)
		if ok {
			result[fullDomain]++
		}
	}

	err := scanner.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func extractDomain(email string, domain string) (fullDomain string, ok bool) {
	_, fullDomain, found := strings.Cut(email, "@")
	if !found {
		return "", false
	}

	domainParts := strings.Split(fullDomain, ".")
	if len(domainParts) < 2 || domain != domainParts[len(domainParts)-1] {
		return "", false
	}

	return strings.ToLower(fullDomain), true
}
