package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

type Service struct {
	Name      string
	URL       string
	HealthURL string
	TTL       int
	ExpiresAt time.Time
	Healthy   bool
}

func (s Service) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"name": s.Name, "url": s.URL, "health_url": s.HealthURL,
		"ttl_seconds": s.TTL, "expires_at": s.ExpiresAt, "healthy": s.Healthy,
	})
}

type Registry struct {
	mu       sync.RWMutex
	services map[string]Service
	client   *http.Client
}

func New(client *http.Client) *Registry {
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Second}
	}
	return &Registry{services: make(map[string]Service), client: client}
}

func (r *Registry) Register(s Service, now time.Time) error {
	if strings.TrimSpace(s.Name) == "" {
		return errors.New("name is required")
	}
	u, err := url.ParseRequestURI(s.URL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return errors.New("url must be an absolute http or https URL")
	}
	if s.TTL < 1 {
		return errors.New("ttl_seconds must be positive")
	}
	if s.HealthURL == "" {
		s.HealthURL = strings.TrimRight(s.URL, "/") + "/healthz"
	}
	if health, err := url.ParseRequestURI(s.HealthURL); err != nil || (health.Scheme != "http" && health.Scheme != "https") || health.Host == "" {
		return errors.New("health_url must be an absolute http or https URL")
	}
	s.ExpiresAt = now.Add(time.Duration(s.TTL) * time.Second)
	s.Healthy = true
	r.mu.Lock()
	r.services[s.Name] = s
	r.mu.Unlock()
	return nil
}

func (r *Registry) Heartbeat(name string, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.services[name]
	if !ok || !now.Before(s.ExpiresAt) {
		delete(r.services, name)
		return fmt.Errorf("service %q not found or expired", name)
	}
	s.ExpiresAt = now.Add(time.Duration(s.TTL) * time.Second)
	r.services[name] = s
	return nil
}

func (r *Registry) Get(name string, now time.Time) (Service, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.services[name]
	if ok && !now.Before(s.ExpiresAt) {
		delete(r.services, name)
		return Service{}, false
	}
	return s, ok && s.Healthy
}

func (r *Registry) List(now time.Time) []Service {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]Service, 0, len(r.services))
	for name, s := range r.services {
		if !now.Before(s.ExpiresAt) {
			delete(r.services, name)
			continue
		}
		if s.Healthy {
			result = append(result, s)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

func (r *Registry) Check(ctx context.Context, now time.Time) {
	r.mu.Lock()
	services := make([]Service, 0, len(r.services))
	for name, s := range r.services {
		if !now.Before(s.ExpiresAt) {
			delete(r.services, name)
			continue
		}
		services = append(services, s)
	}
	r.mu.Unlock()
	for _, s := range services {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.HealthURL, nil)
		healthy := false
		if err == nil {
			resp, requestErr := r.client.Do(req)
			if requestErr == nil {
				healthy = resp.StatusCode >= 200 && resp.StatusCode < 300
				resp.Body.Close()
			}
		}
		r.mu.Lock()
		current, ok := r.services[s.Name]
		if ok && current.URL == s.URL && now.Before(current.ExpiresAt) {
			current.Healthy = healthy
			r.services[s.Name] = current
		}
		r.mu.Unlock()
	}
}

func (r *Registry) RunChecks(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case now := <-ticker.C:
			r.Check(ctx, now)
		case <-ctx.Done():
			return
		}
	}
}
