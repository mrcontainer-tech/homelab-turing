# dnscrypt-proxy

Encrypted DNS proxy providing DNS-over-HTTPS (DoH) and DNSCrypt for cluster DNS queries.

## Overview

This component deploys dnscrypt-proxy as an in-cluster DNS proxy that encrypts all external DNS queries. CoreDNS forwards to this service instead of using plain DNS to upstream resolvers.

## Architecture

```
Pod → CoreDNS → dnscrypt-proxy → Multiple DoH/DoT providers
                                 (Cloudflare, Google, Quad9, etc.)
```

## Features

- **Multi-provider**: Automatically selects from multiple upstream DoH/DNSCrypt providers
- **Privacy-focused**: Only uses resolvers that don't log queries
- **High availability**: 2 replicas spread across nodes
- **Resource efficient**: Optimized for Raspberry Pi constraints

## Configuration

The proxy is configured via `00-configmap.yaml` with the following key settings:

| Setting | Value | Description |
|---------|-------|-------------|
| `require_nolog` | true | Only use resolvers that don't log |
| `require_nofilter` | true | Only use resolvers that don't filter |
| `lb_strategy` | p2 | Power-of-two load balancing |
| `cache_size` | 4096 | DNS cache entries |

## Verification

```bash
# Check pods are running
kubectl get pods -n kube-system -l app.kubernetes.io/name=dnscrypt-proxy

# Test DNS resolution
kubectl run -it --rm dns-test --image=busybox:1.36 --restart=Never -- \
  nslookup google.com dnscrypt-proxy.kube-system.svc.cluster.local
```

## Related

- [ADR D15: Encrypted DNS](../../adr/D15-encrypted-dns.md) - Decision record for this implementation
- [CoreDNS configuration](../coredns/) - Forwards to this service
