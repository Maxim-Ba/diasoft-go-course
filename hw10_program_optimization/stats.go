package hw10programoptimization

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
)

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
	u, err := getUsers(r)
	if err != nil {
		return nil, fmt.Errorf("get users error: %w", err)
	}
	return countDomains(u, domain)
}

type users [100_000]User

func getUsers(r io.Reader) (result users, err error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		var user User
		if err = json.Unmarshal([]byte(line), &user); err != nil {
			return
		}
		result[i] = user
	}
	return
}

func countDomains(u users, domain string) (DomainStat, error) {
	result := make(DomainStat)
	// 1 компиляция regexp
	re, err := regexp.Compile(`\.` + domain)
	if err != nil {
		return nil, err
	}
	for _, user := range u {
		matched := re.MatchString(user.Email)

		if matched {
			domainPart := strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])
			result[domainPart]++ // Вместо двух вызовов
		}
	}
	return result, nil
}
