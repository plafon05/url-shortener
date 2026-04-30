package remove

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

type removeMock struct {
	deleteFn func(ctx context.Context, alias string) error
}

func (m removeMock) DeleteURL(ctx context.Context, alias string) error {
	return m.deleteFn(ctx, alias)
}

func removeTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func withAlias(req *http.Request, alias string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("alias", alias)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestRemoveHandler_EmptyAlias(t *testing.T) {
	h := New(removeTestLogger(), removeMock{deleteFn: func(context.Context, string) error { return nil }})

	req := httptest.NewRequest(http.MethodDelete, "/url/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "alias cannot be empty", resp.Error)
}

func TestRemoveHandler_NotFound(t *testing.T) {
	h := New(removeTestLogger(), removeMock{deleteFn: func(context.Context, string) error { return storage.ErrAliasNotFound }})

	req := withAlias(httptest.NewRequest(http.MethodDelete, "/url/a1", nil), "a1")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)

	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "not found", resp.Error)
}

func TestRemoveHandler_InternalError(t *testing.T) {
	h := New(removeTestLogger(), removeMock{deleteFn: func(context.Context, string) error { return errors.New("db") }})

	req := withAlias(httptest.NewRequest(http.MethodDelete, "/url/a1", nil), "a1")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)

	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "internal error", resp.Error)
}

func TestRemoveHandler_Success(t *testing.T) {
	h := New(removeTestLogger(), removeMock{deleteFn: func(_ context.Context, alias string) error {
		require.Equal(t, "a1", alias)
		return nil
	}})

	req := withAlias(httptest.NewRequest(http.MethodDelete, "/url/a1", nil), "a1")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, response.StatusOK, resp.Status)
}
