# CoreDNS Custom Configuration

Overrides the k3s-bundled CoreDNS `Corefile` so cluster DNS forwards to
[dnscrypt-proxy](../dnscrypt-proxy/) (encrypted upstream DNS) instead of
`/etc/resolv.conf`, which Tailscale rewrites to `100.100.100.100` (MagicDNS)
causing timeouts for external domains.

## ⚠️ Namespace exception

This component deploys into **`kube-system`** (where k3s runs CoreDNS), not
into a `coredns` namespace. The ArgoCD Application's destination namespace
(`coredns`) is therefore misleading; the manifest's explicit
`namespace: kube-system` wins.

## ⚠️ Ownership conflict with k3s (known issue)

This ConfigMap **replaces an object the k3s addon controller also manages**
(it carries `objectset.rio.cattle.io/*` annotations). k3s re-applies its
bundled version on service restart/upgrade; ArgoCD self-heal then stomps it
back. The failure mode is non-deterministic — see the June 2026 incident in
`analysis-report.html`, where an invalid `forward` directive sat dormant for
months and took down cluster DNS when the pod restarted.

Rules for editing the Corefile here:

- The `forward` plugin accepts **only IP addresses** (or a resolv.conf path),
  never DNS names — CoreDNS cannot resolve names before it is running.
- Validate before committing: `docker run --rm -v $PWD:/cfg coredns/coredns
  -conf /cfg/Corefile -dns.port 1053` or at minimum review against the
  CoreDNS plugin docs.
- Long-term fix (staged, not active): take full GitOps ownership of CoreDNS
  (2 replicas, PDB, Service pinned to `10.43.0.10`) so there is exactly one
  owner. The manifests and the ordered activation runbook live in
  [takeover/](takeover/).
