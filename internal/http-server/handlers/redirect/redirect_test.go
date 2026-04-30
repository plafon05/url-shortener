package redirect

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"httpServer_project/internal/storage"
	"httpServer_project/lib/api/response"
)

type redirectMock struct {
	getFn func(ctx context.Context, alias string) (string, error)
}

func (m redirectMock) GetURL(ctx context.Context, alias string) (string, error) {
	return m.getFn(ctx, alias)
}

func redirectTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func withRedirectAlias(req *http.Request, alias string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("alias", alias)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestRedirectHandler_EmptyAlias(t *testing.T) {
	h := New(redirectTestLogger(), redirectMock{getFn: func(context.Context, string) (string, error) { return "", nil }})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "alias cannot be empty", resp.Error)
}

func TestRedirectHandler_NotFound(t *testing.T) {
	h := New(redirectTestLogger(), redirectMock{getFn: func(context.Context, string) (string, error) {
		return "", storage.ErrURLNotFound
	}})

	req := withRedirectAlias(httptest.NewRequest(http.MethodGet, "/a1", nil), "a1")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)

	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "not found", resp.Error)
}

func TestRedirectHandler_InternalError(t *testing.T) {
	h := New(redirectTestLogger(), redirectMock{getFn: func(context.Context, string) (string, error) {
		return "", errors.New("db")
	}})

	req := withRedirectAlias(httptest.NewRequest(http.MethodGet, "/a1", nil), "a1")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)

	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "internal error", resp.Error)
}

func TestRedirectHandler_Success(t *testing.T) {
	h := New(redirectTestLogger(), redirectMock{getFn: func(_ context.Context, alias string) (string, error) {
		require.Equal(t, "a1", alias)
		return "https://example.com", nil
	}})

	req := withRedirectAlias(httptest.NewRequest(http.MethodGet, "/a1", nil), "a1")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	require.Equal(t, http.StatusFound, w.Code)
	require.Equal(t, "https://example.com", w.Header().Get("Location"))
}
