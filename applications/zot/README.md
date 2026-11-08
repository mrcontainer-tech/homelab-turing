# Zot

OCI pull-through cache (D24). Kyverno rewrites all app images to
`zot.mrcontainer.nl/...`, so when Zot is down every new image pull fails.

## Chart Info

- Chart: `project-zot/zot`, see `appsets/applications-infrastructure-chart-application-set.yaml`
- Values: `values.yaml`. Ingress, TLS and Tailscale ingress: `manifests/`

## Node requirement: inotify instance limit

Zot creates an inotify watcher at startup. If that fails it exits with
`too many open files` / `failed to create a new hot reloader` and crash-loops.

Cause: per-UID `fs.inotify.max_user_instances` (Debian default 128). Zot runs
as root and shares that budget with kubelet, containerd and every root pod.

Fix on every node (host setting, not settable from a Pod):

```bash
sudo tee /etc/sysctl.d/99-inotify.conf <<'SYSCTL'
fs.inotify.max_user_instances = 1024
fs.inotify.max_user_watches   = 1048576
SYSCTL
sudo sysctl --system
```

Diagnose: `sysctl fs.inotify.max_user_instances` and
`find /proc/*/fd -lname 'anon_inode:inotify' 2>/dev/null | wc -l`.

On Talos (D16) put the same keys under `machine.sysctls`.

## Operators that own their Pod image

Operators that compare the Pod image with their CR spec (CNPG) see the rewrite
as drift. Their CRs must use the zot path; see `core-components/cloudnativepg/README.md`.

## Related

- [D24: Zot registry](../../adr/D24-zot-registry.md)
- [Kyverno](../kyverno/): `mutate-image-registry`, `require-zot-registry`
