package resolve

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"httpServer_project/internal/storage"
	"httpServer_project/lib/api/response"
)

type resolveMock struct {
	getFn func(ctx context.Context, url string) ([]string, error)
}

func (m resolveMock) GetAliasesByURL(ctx context.Context, url string) ([]string, error) {
	return m.getFn(ctx, url)
}

func resolveTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestResolveHandler_EmptyURL(t *testing.T) {
	h := New(resolveTestLogger(), resolveMock{getFn: func(context.Context, string) ([]string, error) {
		return nil, nil
	}})

	req := httptest.NewRequest(http.MethodGet, "/url/aliases", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "url cannot be empty", resp.Error)
}

func TestResolveHandler_InvalidURL(t *testing.T) {
	h := New(resolveTestLogger(), resolveMock{getFn: func(context.Context, string) ([]string, error) {
		return nil, nil
	}})

	req := httptest.NewRequest(http.MethodGet, "/url/aliases?url=invalid", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "invalid URL", resp.Error)
}

func TestResolveHandler_NotFound(t *testing.T) {
	h := New(resolveTestLogger(), resolveMock{getFn: func(context.Context, string) ([]string, error) {
		return nil, storage.ErrAliasNotFound
	}})

	req := httptest.NewRequest(http.MethodGet, "/url/aliases?url=https://example.com", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)

	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "not found", resp.Error)
}

func TestResolveHandler_InternalError(t *testing.T) {
	h := New(resolveTestLogger(), resolveMock{getFn: func(context.Context, string) ([]string, error) {
		return nil, errors.New("unexpected")
	}})

	req := httptest.NewRequest(http.MethodGet, "/url/aliases?url=https://example.com", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)

	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "internal error", resp.Error)
}

func TestResolveHandler_Success(t *testing.T) {
	h := New(resolveTestLogger(), resolveMock{getFn: func(_ context.Context, url string) ([]string, error) {
		require.Equal(t, "https://example.com", url)
		return []string{"a1", "a2"}, nil
	}})

	req := httptest.NewRequest(http.MethodGet, "/url/aliases?url=https://example.com", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, response.StatusOK, resp.Status)
	require.ElementsMatch(t, []string{"a1", "a2"}, resp.Alias)
}
