# service-discovery-lab

Tiny HTTP registry with heartbeat TTLs and periodic HTTP health checks.

Install with go install github.com/ulrichheringer/service-discovery-lab@latest, then run service-discovery-lab. POST /services with name, url, health_url, and ttl_seconds to register; GET /services lists healthy live records.
