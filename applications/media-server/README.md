# Media Server (K3s) — GitOps (Argo CD)

Plain-manifests home-media stack (Jellyfin, Jellyseerr, Sonarr, Radarr, Prowlarr, qBittorrent, Bazarr) delivered via **Argo CD ApplicationSet**.
Large RWX data lives on the SSD of `node4` (`/mnt/ssd`); per-app configs use Longhorn PVCs (RWO).

## Repository layout

```
applications/
└── media-server/
    ├── README.md                 # this file
    └── manifests/                # plain Kubernetes manifests
```

Your cluster has an **ApplicationSet** that discovers any `applications/*/manifests` path and creates an App named `<dir>-manifest`.
For this folder, Argo CD renders and applies everything under `applications/media-server/manifests`.

> Note: Manifests include a `Namespace` object (typically `media`). `syncOptions: CreateNamespace=false` is set at the ApplicationSet level, so the namespace is created by the manifest itself.

## Deploy / sync (Argo CD)

Push to `main` and Argo CD will reconcile automatically (Automated, Prune, SelfHeal).
To force a sync:

```bash
argocd app sync media-server-manifest
```

## Storage prerequisites (on `node4`)

Create the SSD folders and set ownership/permissions for `uid:gid = 1000:1000` (match container user):

```bash
# Mountpoint (if not present)
sudo mkdir -p /mnt/ssd

# Media & downloads layout
sudo mkdir -p /mnt/ssd/media/movies
sudo mkdir -p /mnt/ssd/media/tv
sudo mkdir -p /mnt/ssd/downloads

# Ownership and permissions
sudo chown -R 1000:1000 /mnt/ssd/media /mnt/ssd/downloads
sudo find /mnt/ssd/media /mnt/ssd/downloads -type d -exec chmod 775 {} +
sudo find /mnt/ssd/media /mnt/ssd/downloads -type f -exec chmod 664 {} +

# Optional: default ACLs so new files/dirs inherit group write
sudo setfacl -R -m d:u:1000:rwx,d:g:1000:rwx /mnt/ssd/media /mnt/ssd/downloads
```

## Internal service endpoints (ClusterIP)

Use these URLs for intra-cluster configuration:

* Jellyfin — `http://jellyfin.media.svc.cluster.local:8096`
* Jellyseerr — `http://jellyseerr.media.svc.cluster.local:5055`
* Sonarr — `http://sonarr.media.svc.cluster.local:8989`
* Radarr — `http://radarr.media.svc.cluster.local:7878`
* Prowlarr — `http://prowlarr.media.svc.cluster.local:9696`
* Bazarr — `http://bazarr.media.svc.cluster.local:6767`
* qBittorrent (WebUI/API) — `http://qbittorrent.media.svc.cluster.local:8080`
  (BitTorrent peers: TCP/UDP `6881` via the same Service)

> From pods in the same namespace, you can shorten to `http://<service>:<port>`.

## Public URLs (Ingress via Traefik)

* `https://jellyfin.mrcontainer.nl`
* `https://jellyseerr.mrcontainer.nl`
* `https://sonarr.mrcontainer.nl`
* `https://radarr.mrcontainer.nl`
* `https://prowlarr.mrcontainer.nl`
* `https://qb.mrcontainer.nl`
* `https://bazarr.mrcontainer.nl`

TLS secrets are provided per-app (`*-cert.yaml`). Traefik terminates TLS and forwards to Services.

## Wiring overview

1. **qBittorrent**

   * Enable WebUI (8080). Create categories: `movies`, `tv`.

2. **Radarr/Sonarr → qBittorrent**

   * Add Download Client: qBittorrent (`qbittorrent.media.svc.cluster.local:8080`).
   * Set category: `movies` (Radarr), `tv` (Sonarr).

3. **Prowlarr**

   * Add indexers.
   * **Apps** → add Radarr (`http://radarr.media.svc.cluster.local:7878`) and Sonarr (`http://sonarr.media.svc.cluster.local:8989`), paste API keys, **sync indexers**.

4. **Jellyseerr**

   * Services → add Radarr and Sonarr (URLs above).
   * Jellyfin URL: `https://jellyfin.mrcontainer.nl` (no port; SSL on) *or* `http://jellyfin.media.svc.cluster.local:8096` (no SSL).
   * Leave **URL Base** empty.

## Notes

* Data PVs are hostPath-backed and pinned to `node4` for SSD I/O; config PVCs use the `longhorn` StorageClass.
* Prefer pinning container image tags instead of `:latest`.
* Back up config PVCs via Longhorn (snapshots/backups); media/downloads are outside Longhorn and should be backed up separately.
