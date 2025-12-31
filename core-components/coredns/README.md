# CoreDNS Custom Configuration

This directory contains custom CoreDNS configuration for k3s.

## DNS Configuration

The custom ConfigMap overrides the default k3s CoreDNS forward directive to:
- Use the local network router (192.168.68.1)
- Fall back to Google DNS (8.8.8.8) and Cloudflare DNS (1.1.1.1)
- Avoid Tailscale DNS (100.100.100.100) which causes timeouts in the cluster

## Why this is needed

When Tailscale is running on k3s nodes, it modifies `/etc/resolv.conf` to use `100.100.100.100` (Tailscale MagicDNS). By default, k3s CoreDNS uses `forward . /etc/resolv.conf`, which inherits this configuration and causes DNS timeouts for external domains.

## How it works

K3s will automatically merge this custom ConfigMap with the base CoreDNS configuration.
