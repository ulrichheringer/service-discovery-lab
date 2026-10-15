# service-discovery-lab

A small HTTP registry for registering services, renewing their TTL with heartbeats, and finding instances that pass periodic health checks.

Install with go install github.com/ulrichheringer/service-discovery-lab@latest, then run service-discovery-lab.
Register with POST /services using name, url, and ttl_seconds. The registry checks /healthz by default; GET /services lists healthy instances.
