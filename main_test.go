package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestVirtualItemsAuthenticationAndPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		for key, want := range map[string]string{"api_key": "test-key", "ParentId": "real-id", "StartIndex": "2", "Limit": "5"} {
			if got := q.Get(key); got != want {
				t.Errorf("%s = %q, want %q", key, got, want)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"Items":[],"TotalRecordCount":0}`))
	}))
	defer server.Close()
	oldConfig, oldMap := config, libraryMap
	defer func() { config, libraryMap = oldConfig, oldMap }()
	config = Config{EmbyServer: server.URL}
	libraryMap = map[string]Library{"virtual-id": {ResourceID: "real-id", ResourceType: "collection"}}
	req := httptest.NewRequest("GET", "/emby/Users/user/Items?ParentId=virtual-id&api_key=test-key&StartIndex=2&Limit=5", nil)
	resp := &http.Response{Request: req, Header: http.Header{}}
	if err := hookDetails(resp); err != nil {
		t.Fatal(err)
	}
}

func TestVirtualItemsUpstreamFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"Items":[]}`))
	}))
	defer server.Close()
	oldConfig, oldMap := config, libraryMap
	defer func() { config, libraryMap = oldConfig, oldMap }()
	config = Config{EmbyServer: server.URL}
	libraryMap = map[string]Library{"virtual-id": {ResourceID: "real-id", ResourceType: "collection"}}
	req := httptest.NewRequest("GET", "/emby/Users/user/Items?ParentId=virtual-id", nil)
	if err := hookDetails(&http.Response{Request: req, Header: http.Header{}}); err == nil {
		t.Fatal("expected upstream failure instead of null response")
	}
	_, err := doGetJSON(server.URL, url.Values{}, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected HTTP 401 error, got %v", err)
	}
}
