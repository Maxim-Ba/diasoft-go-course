package hw02unpackstring

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type ArchivedItem struct {
	amount int
	char   string
}

func (ai ArchivedItem) ToString() string {
	return strings.Repeat(ai.char, ai.amount)
}

var ErrInvalidString = errors.New("invalid string")

func Unpack(basestr string) (string, error) {
	var strb strings.Builder
	var possibleNextIdx int
	for i := range basestr {
		if i >= possibleNextIdx {
			ai, next, err := getNextItem(basestr, i)
			if err != nil {
				fmt.Printf(err.Error() + "\n")
				return "", ErrInvalidString
			}
			possibleNextIdx = next
			strb.WriteString(ai.ToString())
		}
	}
	res := strb.String()

	return res, nil
}

// Получает следующий элемент [символ, число], индекс следующего необработаного символа или ошибку.
func getNextItem(reststr string, from int) (ArchivedItem, int, error) {
	// получаем руну из строки
	r, _ := utf8.DecodeRuneInString(reststr[from:])
	from += utf8.RuneLen(r)
	// если не экранированый символ цифра
	if unicode.IsDigit(r) {
		return ArchivedItem{}, from, errors.New("разрешено использование цифр, но не чисел")
	}
	// если экранированый символ
	if string(r) == `\` {
		nextSubstr := reststr[from:]
		// значит текущий символ следующий
		r, _ = utf8.DecodeRuneInString(nextSubstr)
		if !unicode.IsDigit(r) && string(r) != `\` {
			return ArchivedItem{}, from, errors.New("заэкранировать можно только цифру или слэш")
		}
		// еще сдвигаем указатель на следующий за экранированым символом символ
		from += utf8.RuneLen(r)
	}
	amount, from := getAmount(reststr, from)
	return ArchivedItem{
		amount: amount,
		char:   string(r),
	}, from, nil
}

// получаем сколько раз повторить символ, если подстрока начинается не с числа, то 1.
func getAmount(str string, from int) (int, int) {
	var amount int
	amount = 1
	// получаем руну из строки
	r, _ := utf8.DecodeRuneInString(str[from:])
	if unicode.IsDigit(r) {
		amount, _ = strconv.Atoi(string(r))
		from += utf8.RuneLen(r)
	}

	return amount, from
}
