package aliasgen_test

import (
	"httpServer_project/lib/aliasgen"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateAlias_BasicCases(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
	}{
		{"Simple URL", "https://example.com"},
		{"URL with path", "https://example.com/products/123"},
		{"URL with query", "https://example.com/search?q=test"},
		{"Invalid URL", "://bad_url"},
		{"Empty URL", ""},
		{"Short path", "https://ex.com/a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alias := aliasgen.GenerateAlias(tt.rawURL)
			assert.NotEmpty(t, alias, "alias не должен быть пустым")
			assert.LessOrEqual(t, len(alias), aliasgen.AliasMaxLen, "alias должен быть <= 15 символов")
		})
	}
}

func TestGenerateAlias_CleanCharacters(t *testing.T) {
	rawURL := "https://Example.COM/Some_Path--123/?q=test"
	alias := aliasgen.GenerateAlias(rawURL)
	assert.NotContains(t, alias, "_", "alias не должен содержать подчеркиваний")
	assert.NotContains(t, alias, "?", "alias не должен содержать знаков вопроса")
	assert.NotContains(t, alias, "&", "alias не должен содержать амперсандов")
}

func TestGenerateAlias_MinLengthFallback(t *testing.T) {
	// короткий, некорректный URL, чтобы вызвать fallback randomAlias
	rawURL := "a://"
	alias := aliasgen.GenerateAlias(rawURL)
	assert.NotEmpty(t, alias, "alias не должен быть пустым даже при коротком URL")
	assert.LessOrEqual(t, len(alias), aliasgen.AliasMaxLen, "alias должен быть <= 15 символов")
}

func TestGenerateAlias_MaxLengthTruncation(t *testing.T) {
	rawURL := "https://verylonghostname.example.com/path/to/resource"
	alias := aliasgen.GenerateAlias(rawURL)
	assert.LessOrEqual(t, len(alias), aliasgen.AliasMaxLen, "alias должен быть <= 15 символов")
}
