package tests

import (
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"testing"

	"httpServer_project/internal/http-server/handlers/url/save"
)

func BenchmarkAPI_POSTURL(b *testing.B) {
	router := newTestRouter(newInMemoryStorage())
	creds := &authCreds{user: testUser, pass: testPass}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		payload := save.Request{
			URL:   "https://example.com/" + strconv.Itoa(i),
			Alias: "alias-" + strconv.Itoa(i),
		}
		resp := performJSONRequest(router, http.MethodPost, "/url/", payload, creds)
		if resp.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", resp.Code)
		}
	}
}

func BenchmarkAPI_GETRedirect(b *testing.B) {
	router := newTestRouter(newInMemoryStorage())
	creds := &authCreds{user: testUser, pass: testPass}

	seed := performJSONRequest(router, http.MethodPost, "/url/", save.Request{
		URL:   "https://example.com/bench",
		Alias: "bench-redirect",
	}, creds)
	if seed.Code != http.StatusOK {
		b.Fatalf("failed to seed: %d", seed.Code)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp := performRequest(router, http.MethodGet, "/bench-redirect", nil, nil)
		if resp.Code != http.StatusFound {
			b.Fatalf("unexpected status: %d", resp.Code)
		}
	}
}

func BenchmarkAPI_GETAliases(b *testing.B) {
	router := newTestRouter(newInMemoryStorage())
	creds := &authCreds{user: testUser, pass: testPass}
	url := "https://example.com/aliases-bench"

	seed := performJSONRequest(router, http.MethodPost, "/url/", save.Request{
		URL:   url,
		Alias: "bench-alias",
	}, creds)
	if seed.Code != http.StatusOK {
		b.Fatalf("failed to seed: %d", seed.Code)
	}

	path := "/url/aliases?url=" + url

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp := performRequest(router, http.MethodGet, path, nil, creds)
		if resp.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", resp.Code)
		}
	}
}

func BenchmarkAPI_DELETEAlias(b *testing.B) {
	router := newTestRouter(newInMemoryStorage())
	creds := &authCreds{user: testUser, pass: testPass}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		alias := fmt.Sprintf("delete-%d", i)
		create := performJSONRequest(router, http.MethodPost, "/url/", save.Request{
			URL:   "https://example.com/delete/" + strconv.Itoa(i),
			Alias: alias,
		}, creds)
		if create.Code != http.StatusOK {
			b.Fatalf("failed to create alias, status: %d", create.Code)
		}

		resp := performRequest(router, http.MethodDelete, "/url/"+alias, nil, creds)
		if resp.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", resp.Code)
		}
	}
}

func BenchmarkAPI_POSTURL_Parallel(b *testing.B) {
	router := newTestRouter(newInMemoryStorage())
	creds := &authCreds{user: testUser, pass: testPass}
	var counter uint64

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := atomic.AddUint64(&counter, 1)
			payload := save.Request{
				URL:   "https://example.com/p/" + strconv.FormatUint(i, 10),
				Alias: "p-alias-" + strconv.FormatUint(i, 10),
			}
			resp := performJSONRequest(router, http.MethodPost, "/url/", payload, creds)
			if resp.Code != http.StatusOK {
				b.Fatalf("unexpected status: %d", resp.Code)
			}
		}
	})
}
