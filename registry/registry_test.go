package registry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHeartbeatRefreshesTTLAndExpiryRemovesRecord(t *testing.T) {
	r := New(nil)
	now := time.Unix(100, 0)
	s := Service{Name: "catalog", URL: "http://localhost:9000", TTL: 5}
	if err := r.Register(s, now); err != nil {
		t.Fatal(err)
	}
	if err := r.Heartbeat("catalog", now.Add(4*time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, ok := r.Get("catalog", now.Add(8*time.Second)); !ok {
		t.Fatal("heartbeat did not refresh TTL")
	}
	if _, ok := r.Get("catalog", now.Add(10*time.Second)); ok {
		t.Fatal("expired service is visible")
	}
}

func TestCheckMarksUnhealthyEndpoint(t *testing.T) {
	unhealthy := httptest.NewServer(http.NotFoundHandler())
	defer unhealthy.Close()
	r := New(&http.Client{Timeout: time.Second})
	now := time.Now()
	if err := r.Register(Service{Name: "bad", URL: unhealthy.URL, HealthURL: unhealthy.URL, TTL: 30}, now); err != nil {
		t.Fatal(err)
	}
	r.Check(context.Background(), now)
	if _, ok := r.Get("bad", now); ok {
		t.Fatal("unhealthy service was returned")
	}
}
