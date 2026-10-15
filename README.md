# service-discovery-lab — Service Registry

[![Go 1.22](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go&logoColor=white)](https://go.dev/)
![HTTP API](https://img.shields.io/badge/HTTP-API-005C84)
[![MIT License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A small HTTP registry for registering services, renewing their TTL with heartbeats, and finding instances that pass periodic health checks.

Install with go install github.com/ulrichheringer/service-discovery-lab@latest, then run service-discovery-lab.
Register with POST /services using name, url, and ttl_seconds. The registry checks /healthz by default; GET /services lists healthy instances.
