package hw03frequencyanalysis

import (
	"fmt"
	"sort"
	"unicode"
)

type Word struct {
	word  string
	count int
}

const TEN = 10

func Top10(input string) []string {
	counter := make(map[string]int)
	wordstart := -1
	for i, r := range input {
		if unicode.IsSpace(r) {
			if wordstart != -1 {
				word := input[wordstart:i]
				counter[word]++
				wordstart = -1
			}
		} else {
			if wordstart == -1 {
				wordstart = i
			}
		}
		if i == len(input)-1 && wordstart != -1 {
			word := input[wordstart:]
			counter[word]++
		}
	}

	return createdSortedKeys(counter)
}

func createdSortedKeys(counter map[string]int) []string {
	words := make([]Word, 0, len(counter))
	for k, v := range counter {
		words = append(words, Word{count: v, word: k})
	}
	sort.Slice(words, func(i, j int) bool {
		if words[i].count == words[j].count {
			return words[i].word < words[j].word
		}
		return words[i].count > words[j].count
	})
	result := make([]string, 0, TEN)
	for i := 0; i < TEN && i < len(words); i++ {
		result = append(result, words[i].word)
	}
	fmt.Printf("result = %v \n", result)

	return result
}
