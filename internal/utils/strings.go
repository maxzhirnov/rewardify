package utils

import (
	"regexp"
)

// HidePassword принимает строку подключения к Postgres и возвращает
// строку, в которой пароль скрыт.
func HidePassword(connStr string) string {
	re := regexp.MustCompile(`assword='[^']+'`)
	return re.ReplaceAllString(connStr, "assword=*******")
}
