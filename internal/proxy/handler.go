package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

type Handler struct {
	state    *State
	proxy    *httputil.ReverseProxy
	upstream string
}

func NewHandler(state *State, upstream string) http.Handler {
	target, err := url.Parse(upstream)
	if err != nil {
		panic(err)
	}
	return &Handler{
		state: state,
		proxy: httputil.NewSingleHostReverseProxy(target),
		upstream: upstream,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("upstream error: %v", err)
		http.Error(w, "upstream error", http.StatusBadGateway)
	}

	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err == nil {
			r.Body = io.NopCloser(bytes.NewReader(body))
			model := extractModel(body)
			if model != "" && h.state.ShouldLimit(model) {
				wait, ok := h.state.ReserveAndWait()
				if !ok {
					http.Error(w, "nvidia request queue is full, please retry later, remaining wait: "+wait.Truncate(time.Second).String(), http.StatusTooManyRequests)
					return
				}
			}
		} else {
			log.Printf("request read error: %v", err)
		}
	}
	h.proxy.ServeHTTP(w, r)
}

func extractModel(body []byte) string {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	if v, ok := payload["model"].(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}
