//go:generate easyjson -all stats.go
package hw10programoptimization

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

//easyjson:json
type User struct {
	ID       int
	Name     string
	Username string
	Email    string
	Phone    string
	Password string
	Address  string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	scanner := bufio.NewScanner(r)
	return countDomains(scanner, domain)
}

func countDomains(scanner *bufio.Scanner, domain string) (DomainStat, error) {
	result := make(DomainStat)
	// 1 компиляция regexp
	re, err := regexp.Compile(`\.` + domain)
	if err != nil {
		return nil, err
	}
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var user User
		// easyjson вместо encoding/json
		if err := user.UnmarshalJSON(line); err != nil {
			return nil, fmt.Errorf("unmarshal error: %w", err)
		}
		matched := re.MatchString(user.Email)
		if matched {
			domainPart := strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])
			result[domainPart]++ // Вместо двух вызовов
		}
	}
	return result, nil
}
