package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"httpServer_project/internal/http-server/handlers/url/save"
)

func TestURLRoutes_RequireBasicAuth(t *testing.T) {
	router := newTestRouter(newInMemoryStorage())

	testCases := []struct {
		name   string
		method string
		path   string
		body   any
	}{
		{
			name:   "POST /url/ without auth",
			method: http.MethodPost,
			path:   "/url/",
			body: save.Request{
				URL:   "https://example.com",
				Alias: "a1",
			},
		},
		{
			name:   "GET /url/aliases without auth",
			method: http.MethodGet,
			path:   "/url/aliases?url=https://example.com",
		},
		{
			name:   "DELETE /url/{alias} without auth",
			method: http.MethodDelete,
			path:   "/url/a1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			noAuth := performJSONRequest(router, tc.method, tc.path, tc.body, nil)
			require.Equal(t, http.StatusUnauthorized, noAuth.Code)

			wrongAuth := performJSONRequest(router, tc.method, tc.path, tc.body, &authCreds{user: "wrong", pass: "wrong"})
			require.Equal(t, http.StatusUnauthorized, wrongAuth.Code)
		})
	}
}

func TestURLRoutes_AllowValidBasicAuth(t *testing.T) {
	router := newTestRouter(newInMemoryStorage())

	resp := performJSONRequest(router, http.MethodPost, "/url/", save.Request{
		URL:   "https://example.com",
		Alias: "a1",
	}, &authCreds{user: testUser, pass: testPass})

	require.NotEqual(t, http.StatusUnauthorized, resp.Code)
	require.Equal(t, http.StatusOK, resp.Code)
}
