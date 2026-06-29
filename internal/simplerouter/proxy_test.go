package simplerouter

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInjectProviderRouting(t *testing.T) {
	out := injectProviderRouting([]byte(`{"model":"z-ai/glm-5.2","max_tokens":8}`), "deepinfra/fp4")
	var obj map[string]any
	if err := json.Unmarshal(out, &obj); err != nil {
		t.Fatalf("result not JSON: %v", err)
	}
	if obj["model"] != "z-ai/glm-5.2" {
		t.Fatalf("model not preserved: %v", obj["model"])
	}
	prov, ok := obj["provider"].(map[string]any)
	if !ok {
		t.Fatalf("provider field missing: %v", obj["provider"])
	}
	only, ok := prov["only"].([]any)
	if !ok || len(only) != 1 || only[0] != "deepinfra/fp4" {
		t.Fatalf("provider.only = %v", prov["only"])
	}
	if prov["allow_fallbacks"] != false {
		t.Fatalf("allow_fallbacks = %v, want false", prov["allow_fallbacks"])
	}

	// Non-JSON bodies pass through untouched.
	if got := injectProviderRouting([]byte("not json"), "x"); string(got) != "not json" {
		t.Fatalf("non-JSON body altered: %q", got)
	}
}

func TestProviderProxyForwardsAndInjects(t *testing.T) {
	var gotPath, gotAuth string
	var gotBody map[string]any
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer upstream.Close()

	baseURL, stop, err := startProviderProxy(upstream.URL+"/api", "novita/fp8")
	if err != nil {
		t.Fatal(err)
	}
	defer stop()

	req, _ := http.NewRequest(http.MethodPost, baseURL+"/v1/messages", strings.NewReader(`{"model":"z-ai/glm-5.2","messages":[]}`))
	req.Header.Set("Authorization", "Bearer sk-or-test")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if gotPath != "/api/v1/messages" {
		t.Fatalf("upstream path = %q, want /api/v1/messages", gotPath)
	}
	if gotAuth != "Bearer sk-or-test" {
		t.Fatalf("auth header not forwarded: %q", gotAuth)
	}
	if gotBody["model"] != "z-ai/glm-5.2" {
		t.Fatalf("model not forwarded: %v", gotBody["model"])
	}
	prov, ok := gotBody["provider"].(map[string]any)
	if !ok || prov["only"].([]any)[0] != "novita/fp8" {
		t.Fatalf("provider routing not injected: %v", gotBody["provider"])
	}
}
