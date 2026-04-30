package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"httpServer_project/internal/http-server/handlers/alias/resolve"
	"httpServer_project/internal/http-server/handlers/url/save"
	"httpServer_project/lib/api/response"
)

func TestURLShortener_HappyPath(t *testing.T) {
	router := newTestRouter(newInMemoryStorage())

	createResp := performJSONRequest(router, http.MethodPost, "/url/", save.Request{
		URL:   "https://example.com/path",
		Alias: "my-alias",
	}, &authCreds{user: testUser, pass: testPass})

	require.Equal(t, http.StatusOK, createResp.Code)
	created := decodeJSON[save.Response](t, createResp.Body.Bytes())
	require.Equal(t, response.StatusOK, created.Status)
	require.Equal(t, "my-alias", created.Alias)

	redirectResp := performRequest(router, http.MethodGet, "/my-alias", nil, nil)
	require.Equal(t, http.StatusFound, redirectResp.Code)
	require.Equal(t, "https://example.com/path", redirectResp.Header().Get("Location"))
}

func TestURLShortener_CreateResolveDeleteFlow(t *testing.T) {
	router := newTestRouter(newInMemoryStorage())

	const rawURL = "https://example.com/some/path"
	const alias = "flow-alias"

	createResp := performJSONRequest(router, http.MethodPost, "/url/", save.Request{
		URL:   rawURL,
		Alias: alias,
	}, &authCreds{user: testUser, pass: testPass})
	require.Equal(t, http.StatusOK, createResp.Code)

	aliasesResp := performRequest(router, http.MethodGet, "/url/aliases?url="+rawURL, nil, &authCreds{user: testUser, pass: testPass})
	require.Equal(t, http.StatusOK, aliasesResp.Code)

	aliases := decodeJSON[resolve.Response](t, aliasesResp.Body.Bytes())
	require.Equal(t, response.StatusOK, aliases.Status)
	require.Contains(t, aliases.Alias, alias)

	deleteResp := performRequest(router, http.MethodDelete, "/url/"+alias, nil, &authCreds{user: testUser, pass: testPass})
	require.Equal(t, http.StatusOK, deleteResp.Code)

	redirectAfterDelete := performRequest(router, http.MethodGet, "/"+alias, nil, nil)
	require.Equal(t, http.StatusNotFound, redirectAfterDelete.Code)
}

func TestURLShortener_AliasesAndDeleteEndpoints(t *testing.T) {
	router := newTestRouter(newInMemoryStorage())

	const rawURL = "https://example.org/resource"
	const alias = "alias-endpoint"

	createResp := performJSONRequest(router, http.MethodPost, "/url/", save.Request{
		URL:   rawURL,
		Alias: alias,
	}, &authCreds{user: testUser, pass: testPass})
	require.Equal(t, http.StatusOK, createResp.Code)

	aliasesResp := performRequest(router, http.MethodGet, "/url/aliases?url="+rawURL, nil, &authCreds{user: testUser, pass: testPass})
	require.Equal(t, http.StatusOK, aliasesResp.Code)
	aliases := decodeJSON[resolve.Response](t, aliasesResp.Body.Bytes())
	require.Contains(t, aliases.Alias, alias)

	deleteResp := performRequest(router, http.MethodDelete, "/url/"+alias, nil, &authCreds{user: testUser, pass: testPass})
	require.Equal(t, http.StatusOK, deleteResp.Code)
}
