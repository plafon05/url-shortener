package save

import (
	"bytes"
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

type saveMock struct {
	saveFn func(ctx context.Context, urlToSave, alias string) (int64, error)
}

func (m saveMock) SaveURL(ctx context.Context, urlToSave, alias string) (int64, error) {
	return m.saveFn(ctx, urlToSave, alias)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestSaveHandler_DecodeError(t *testing.T) {
	h := New(testLogger(), saveMock{saveFn: func(context.Context, string, string) (int64, error) {
		return 1, nil
	}})

	req := httptest.NewRequest(http.MethodPost, "/url/", bytes.NewBufferString("{"))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, response.StatusError, resp.Status)
	require.Equal(t, "failed to decode request", resp.Error)
}

func TestSaveHandler_ValidationError(t *testing.T) {
	h := New(testLogger(), saveMock{saveFn: func(context.Context, string, string) (int64, error) {
		return 1, nil
	}})

	body := `{"url":"invalid_url","alias":"a"}`
	req := httptest.NewRequest(http.MethodPost, "/url/", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, response.StatusError, resp.Status)
	require.Contains(t, resp.Error, "поле URL содержит недопустимый URL")
}

func TestSaveHandler_AliasExists(t *testing.T) {
	h := New(testLogger(), saveMock{saveFn: func(context.Context, string, string) (int64, error) {
		return 0, storage.ErrAliasExists
	}})

	body := `{"url":"https://example.com","alias":"known"}`
	req := httptest.NewRequest(http.MethodPost, "/url/", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	require.Equal(t, http.StatusConflict, w.Code)

	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "alias already exists", resp.Error)
}

func TestSaveHandler_InternalError(t *testing.T) {
	h := New(testLogger(), saveMock{saveFn: func(context.Context, string, string) (int64, error) {
		return 0, errors.New("db down")
	}})

	body := `{"url":"https://example.com","alias":"a"}`
	req := httptest.NewRequest(http.MethodPost, "/url/", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)

	var resp response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "internal error", resp.Error)
}

func TestSaveHandler_SuccessWithAlias(t *testing.T) {
	h := New(testLogger(), saveMock{saveFn: func(_ context.Context, urlToSave, alias string) (int64, error) {
		require.Equal(t, "https://example.com", urlToSave)
		require.Equal(t, "my-alias", alias)
		return 42, nil
	}})

	body := `{"url":"https://example.com","alias":"my-alias"}`
	req := httptest.NewRequest(http.MethodPost, "/url/", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, response.StatusOK, resp.Status)
	require.Equal(t, "my-alias", resp.Alias)
}

func TestSaveHandler_SuccessWithGeneratedAlias(t *testing.T) {
	h := New(testLogger(), saveMock{saveFn: func(_ context.Context, _ string, alias string) (int64, error) {
		require.NotEmpty(t, alias)
		return 1, nil
	}})

	body := `{"url":"https://example.com/some/path"}`
	req := httptest.NewRequest(http.MethodPost, "/url/", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, response.StatusOK, resp.Status)
	require.NotEmpty(t, resp.Alias)
}
