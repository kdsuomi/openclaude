package simplerouter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
)

// startProviderProxy launches a localhost reverse proxy that forwards requests
// to the OpenRouter Anthropic endpoint, injecting `provider.only` into each JSON
// body so Claude Code is pinned to a single provider/endpoint. It returns the
// base URL to use as ANTHROPIC_BASE_URL and a stop func to shut it down.
//
// This is needed because OpenRouter only honors provider routing in the request
// body, and Claude Code offers no way to add body fields — so simplerouter rewrites
// the body in transit. The proxy is bound to 127.0.0.1 and lives only for the
// session.
func startProviderProxy(target, providerTag string) (baseURL string, stop func(), err error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return "", nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.FlushInterval = -1 // stream SSE responses immediately

	baseDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		baseDirector(req)
		req.Host = targetURL.Host // correct Host/SNI for the upstream
		if req.Body == nil {
			return
		}
		body, readErr := io.ReadAll(req.Body)
		req.Body.Close()
		if readErr != nil {
			req.Body = io.NopCloser(bytes.NewReader(nil))
			return
		}
		injected := injectProviderRouting(body, providerTag)
		req.Body = io.NopCloser(bytes.NewReader(injected))
		req.ContentLength = int64(len(injected))
		req.Header.Set("Content-Length", strconv.Itoa(len(injected)))
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", nil, err
	}
	server := &http.Server{Handler: proxy}
	go server.Serve(listener)

	baseURL = fmt.Sprintf("http://%s", listener.Addr().String())
	stop = func() { _ = server.Close() }
	return baseURL, stop, nil
}

// injectProviderRouting sets the OpenRouter `provider` field on a JSON body to
// pin a single endpoint. If the body isn't valid JSON it is returned unchanged.
func injectProviderRouting(body []byte, providerTag string) []byte {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(body, &obj); err != nil {
		return body
	}
	provider, _ := json.Marshal(map[string]any{
		"only":            []string{providerTag},
		"allow_fallbacks": false,
	})
	obj["provider"] = provider
	out, err := json.Marshal(obj)
	if err != nil {
		return body
	}
	return out
}
