package aliasgen

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/url"
	"regexp"
	"strings"
)

const AliasMaxLen = 15

func GenerateAlias(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		log.Printf("Ошибка: не удалось разобрать URL (%s): %v", rawURL, err)
		return randomAlias("link")
	}

	host := cleanToken(u.Hostname())
	if host == "" {
		log.Printf("Предупреждение: у URL %s отсутствует корректное имя хоста", rawURL)
		host = "link"
	}

	segments := strings.Split(strings.Trim(u.Path, "/"), "/")
	var part string

	for _, s := range segments {
		s = cleanToken(s)
		if len(s) > 2 {
			part = s
			break
		}
	}

	alias := host
	if part != "" {
		alias = host + "-" + part
	}

	alias = cleanToken(alias)

	if len(alias) > AliasMaxLen {
		log.Printf("Предупреждение: алиас '%s' слишком длинный, обрезаем до %d символов", alias, AliasMaxLen)
		alias = alias[:AliasMaxLen]
	}

	if len(alias) < 4 {
		log.Printf("Предупреждение: алиас '%s' слишком короткий, генерируем случайный", alias)
		alias = randomAlias(host)
	}

	return alias
}

var cleanRe = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func cleanToken(s string) string {
	s = strings.ToLower(s)
	s = cleanRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func randomAlias(prefix string) string {
	b := make([]byte, 2)
	_, err := rand.Read(b)
	if err != nil {
		log.Printf("Ошибка генерации случайных байтов: %v", err)
		// fallback: просто timestamp или фиксированный набор
		return prefix + "-xx"
	}
	return prefix + "-" + hex.EncodeToString(b)
}
