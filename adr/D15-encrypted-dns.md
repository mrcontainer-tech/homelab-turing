# D15: Encrypted DNS for Cluster External Queries

**Date**: 2026-01-20
**Status**: Accepted
**Decision**: Option 2 (dnscrypt-proxy) - Multi-provider encrypted DNS
**Context**: DNS queries from the cluster to external services are currently unencrypted

---

## Problem Statement

The cluster's CoreDNS currently forwards external DNS queries using plain DNS (UDP/53) to:
- Local router (192.168.68.1)
- Google DNS (8.8.8.8)
- Cloudflare DNS (1.1.1.1)

This means DNS queries leave the cluster unencrypted, exposing:
- Which domains the cluster is resolving
- Potential for DNS spoofing/manipulation
- Privacy concerns on shared networks

## Goals

- Encrypt DNS queries from the cluster to upstream resolvers
- Maintain low latency for DNS resolution
- Keep resource usage minimal (Raspberry Pi constraints)
- Follow GitOps/Kubernetes-native patterns

---

## Options Considered

### Option 1: cloudflared (Cloudflare Tunnel DNS Proxy)

Deploy cloudflared as a Kubernetes Deployment/DaemonSet that acts as a DNS-over-HTTPS (DoH) proxy.

**Architecture**:
```
Pod → CoreDNS → cloudflared Service → Cloudflare DoH (1.1.1.1)
```

**Pros**:
- Official Cloudflare-maintained image
- Minimal resource usage (~10MB RAM)
- Simple configuration
- Well-documented for Kubernetes use
- ARM64 support available

**Cons**:
- Locked to Cloudflare as upstream provider
- Depends on external Cloudflare infrastructure

**Implementation**:
```yaml
# Deployment in kube-system or dedicated namespace
# Expose as ClusterIP Service on port 53
# Update CoreDNS forward directive to point to cloudflared service
```

---

### Option 2: dnscrypt-proxy

Deploy dnscrypt-proxy as a DNS proxy supporting DoH, DoT, and DNSCrypt protocols.

**Architecture**:
```
Pod → CoreDNS → dnscrypt-proxy Service → Multiple DoH/DoT providers
```

**Pros**:
- Multi-provider support (not locked to single vendor)
- Supports DoH, DoT, and DNSCrypt
- Built-in server health checking and failover
- Can use multiple upstream servers simultaneously
- ARM64 support

**Cons**:
- More complex configuration
- Slightly higher resource usage than cloudflared
- Less Kubernetes-specific documentation

---

### Option 3: CoreDNS with DoT Forward Plugin

Use CoreDNS's forward plugin with TLS to upstream DoT servers.

**Architecture**:
```
Pod → CoreDNS → DoT upstream (e.g., 1.1.1.1:853)
```

**Pros**:
- No additional components needed
- Native CoreDNS solution

**Cons**:
- DoT support in CoreDNS forward plugin is limited
- Less flexible than dedicated proxy
- May have issues with certain DoT implementations

**Implementation**:
```
forward . tls://1.1.1.1 tls://8.8.8.8 {
    tls_servername cloudflare-dns.com
    health_check 5s
}
```

---

### Option 4: AdGuard Home

Deploy AdGuard Home as a full-featured DNS server with DoH/DoT support and ad-blocking.

**Architecture**:
```
Pod → CoreDNS → AdGuard Home → DoH/DoT upstream
```

**Pros**:
- Web UI for management
- Ad-blocking and filtering capabilities
- DoH and DoT support
- Query logging and statistics

**Cons**:
- Higher resource usage (not ideal for Pi)
- Overkill if only encryption is needed
- Another UI/system to manage

---

### Option 5: Node-Level Encrypted DNS (systemd-resolved or stubby)

Configure encrypted DNS at the node level rather than in Kubernetes.

**Pros**:
- Protects all node traffic, not just cluster DNS
- Independent of Kubernetes

**Cons**:
- Conflicts with k3s CoreDNS pattern
- Harder to manage via GitOps
- May cause issues with Tailscale DNS (existing problem)
- Not Kubernetes-native

---

## Comparison Matrix

| Criteria | cloudflared | dnscrypt-proxy | CoreDNS DoT | AdGuard Home |
|----------|-------------|----------------|-------------|--------------|
| Resource Usage | Low | Low-Medium | Lowest | Medium-High |
| Setup Complexity | Simple | Medium | Simple | Medium |
| Provider Flexibility | Cloudflare only | Multiple | Multiple | Multiple |
| ARM64 Support | Yes | Yes | Yes | Yes |
| GitOps Friendly | Yes | Yes | Yes | Yes |
| Kubernetes Native | Yes | Yes | Yes | Yes |
| Additional Features | Minimal | Health checks | None | Ad-blocking, UI |

---

## Decision

**Option 2: dnscrypt-proxy** was chosen for this homelab for the following reasons:

1. **Multi-Provider Flexibility**: Not locked to a single DNS provider; uses Cloudflare, Google, Quad9, and others
2. **Resilience**: Automatic failover between providers if one becomes unavailable
3. **Privacy**: Configured to only use resolvers that don't log queries (`require_nolog = true`)
4. **Resource Efficiency**: Suitable for Raspberry Pi CM4 nodes (32-128Mi memory)
5. **ARM64 Support**: `klutchell/dnscrypt-proxy` image supports ARM64 natively
6. **GitOps Compatible**: Deployed as Kubernetes manifests in `core-components/dnscrypt-proxy/`

While cloudflared (Option 1) would have been simpler, the requirement to avoid single-provider dependency made dnscrypt-proxy the better choice.

---

## Implementation Plan

If this decision is accepted:

1. Create `core-components/cloudflared/` directory structure
2. Deploy cloudflared as a Deployment with 2 replicas for HA
3. Create ClusterIP Service exposing DNS (UDP/TCP 53)
4. Update CoreDNS ConfigMap to forward to cloudflared service
5. Test DNS resolution and verify encryption (packet capture or DoH logs)

### Example cloudflared Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudflared
  namespace: kube-system
spec:
  replicas: 2
  selector:
    matchLabels:
      app: cloudflared
  template:
    metadata:
      labels:
        app: cloudflared
    spec:
      containers:
      - name: cloudflared
        image: cloudflare/cloudflared:latest
        args:
        - proxy-dns
        - --port=5053
        - --upstream=https://1.1.1.1/dns-query
        - --upstream=https://1.0.0.1/dns-query
        ports:
        - containerPort: 5053
          protocol: UDP
        - containerPort: 5053
          protocol: TCP
        resources:
          requests:
            memory: "32Mi"
            cpu: "10m"
          limits:
            memory: "64Mi"
            cpu: "100m"
---
apiVersion: v1
kind: Service
metadata:
  name: cloudflared
  namespace: kube-system
spec:
  selector:
    app: cloudflared
  ports:
  - name: dns-udp
    port: 53
    targetPort: 5053
    protocol: UDP
  - name: dns-tcp
    port: 53
    targetPort: 5053
    protocol: TCP
```

### Updated CoreDNS Forward

```
forward . cloudflared.kube-system.svc.cluster.local
```

---

## References

- [Cloudflare DNS over HTTPS](https://developers.cloudflare.com/1.1.1.1/encryption/dns-over-https/)
- [cloudflared Docker image](https://hub.docker.com/r/cloudflare/cloudflared)
- [dnscrypt-proxy](https://github.com/DNSCrypt/dnscrypt-proxy)
- [CoreDNS forward plugin](https://coredns.io/plugins/forward/)
- [AdGuard Home](https://github.com/AdguardTeam/AdGuardHome)
