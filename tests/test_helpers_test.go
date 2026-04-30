package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
	"testing"

	"httpServer_project/internal/http-server/handlers/alias/resolve"
	"httpServer_project/internal/http-server/handlers/redirect"
	"httpServer_project/internal/http-server/handlers/url/remove"
	"httpServer_project/internal/http-server/handlers/url/save"
	mwLogger "httpServer_project/internal/http-server/middleware/logger"
	"httpServer_project/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/require"
)

const (
	testUser = "myuser"
	testPass = "mypass"
)

type authCreds struct {
	user string
	pass string
}

type inMemoryStorage struct {
	mu           sync.RWMutex
	nextID       int64
	aliasToURL   map[string]string
	urlToAliases map[string]map[string]struct{}
}

func newInMemoryStorage() *inMemoryStorage {
	return &inMemoryStorage{
		nextID:       1,
		aliasToURL:   make(map[string]string),
		urlToAliases: make(map[string]map[string]struct{}),
	}
}

func (s *inMemoryStorage) SaveURL(_ context.Context, urlToSave, alias string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.aliasToURL[alias]; ok {
		return 0, storage.ErrAliasExists
	}

	id := s.nextID
	s.nextID++
	s.aliasToURL[alias] = urlToSave
	if _, ok := s.urlToAliases[urlToSave]; !ok {
		s.urlToAliases[urlToSave] = make(map[string]struct{})
	}
	s.urlToAliases[urlToSave][alias] = struct{}{}

	return id, nil
}

func (s *inMemoryStorage) GetURL(_ context.Context, alias string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.aliasToURL[alias]
	if !ok {
		return "", storage.ErrURLNotFound
	}

	return url, nil
}

func (s *inMemoryStorage) DeleteURL(_ context.Context, alias string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	url, ok := s.aliasToURL[alias]
	if !ok {
		return storage.ErrAliasNotFound
	}

	delete(s.aliasToURL, alias)
	if aliases, exists := s.urlToAliases[url]; exists {
		delete(aliases, alias)
		if len(aliases) == 0 {
			delete(s.urlToAliases, url)
		}
	}

	return nil
}

func (s *inMemoryStorage) GetAliasesByURL(_ context.Context, rawURL string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	aliasesSet, ok := s.urlToAliases[rawURL]
	if !ok || len(aliasesSet) == 0 {
		return nil, storage.ErrAliasNotFound
	}

	aliases := make([]string, 0, len(aliasesSet))
	for alias := range aliasesSet {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)

	return aliases, nil
}

func newTestRouter(store *inMemoryStorage) http.Handler {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			testUser: testPass,
		}))
		r.Post("/", save.New(log, store))
		r.Delete("/{alias}", remove.New(log, store))
		r.Get("/aliases", resolve.New(log, store))
	})

	router.Get("/{alias}", redirect.New(log, store))

	return router
}

func performRequest(router http.Handler, method, path string, body []byte, creds *authCreds) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	if creds != nil {
		req.SetBasicAuth(creds.user, creds.pass)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

func performJSONRequest(router http.Handler, method, path string, payload any, creds *authCreds) *httptest.ResponseRecorder {
	var body []byte
	if payload != nil {
		data, _ := json.Marshal(payload)
		body = data
	}
	return performRequest(router, method, path, body, creds)
}

func decodeJSON[T any](t *testing.T, body []byte) T {
	t.Helper()

	var out T
	require.NoError(t, json.Unmarshal(body, &out))
	return out
}
