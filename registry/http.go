package registry

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func (r *Registry) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("POST /services", r.registerHTTP)
	mux.HandleFunc("GET /services", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, r.List(time.Now())) })
	mux.HandleFunc("/services/", r.serviceHTTP)
	return mux
}

func (r *Registry) registerHTTP(w http.ResponseWriter, req *http.Request) {
	var body map[string]json.RawMessage
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	var s Service
	_ = json.Unmarshal(body["name"], &s.Name)
	_ = json.Unmarshal(body["url"], &s.URL)
	_ = json.Unmarshal(body["health_url"], &s.HealthURL)
	_ = json.Unmarshal(body["ttl_seconds"], &s.TTL)
	if err := r.Register(s, time.Now()); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, s)
}

func (r *Registry) serviceHTTP(w http.ResponseWriter, req *http.Request) {
	path := strings.Trim(strings.TrimPrefix(req.URL.Path, "/services/"), "/")
	parts := strings.Split(path, "/")
	if len(parts) == 2 && parts[1] == "heartbeat" && req.Method == http.MethodPut {
		if err := r.Heartbeat(parts[0], time.Now()); err != nil {
			http.NotFound(w, req)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if len(parts) == 1 && req.Method == http.MethodDelete {
		r.mu.Lock()
		delete(r.services, parts[0])
		r.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if len(parts) == 1 && req.Method == http.MethodGet {
		s, ok := r.Get(parts[0], time.Now())
		if !ok {
			http.NotFound(w, req)
			return
		}
		writeJSON(w, s)
		return
	}
	http.NotFound(w, req)
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
