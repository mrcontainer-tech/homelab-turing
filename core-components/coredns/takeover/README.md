# CoreDNS takeover — staged, NOT active

These manifests give this repo full ownership of CoreDNS (2 replicas, PDB,
node spread), ending the dual-ownership fight between ArgoCD and the k3s
addon controller (root cause amplifier of the June 2026 DNS outage).

**They are deliberately outside `manifests/`**, so the ApplicationSet does
not deploy them. Activation is a manual, ordered procedure because it
involves node-level k3s changes.

## Design choices

- **Same resource names as the k3s addon** (`coredns` Deployment/SA,
  `system:coredns` ClusterRole, `kube-dns` Service). ArgoCD adopts the live
  objects in place via ServerSideApply — no delete/recreate, no DNS gap.
- **`.skip` file, not `--disable`**: a `coredns.yaml.skip` marker makes k3s
  stop reconciling the addon but **leaves the live resources untouched**.
  `--disable` would *uninstall* the addon (deleting the kube-dns Service —
  an instant DNS gap that ArgoCD, which itself needs DNS, may not recover
  from unattended).
- The `coredns` ConfigMap stays where it already lives:
  `../manifests/coredns-custom.yaml`.

## Procedure (run only when the cluster is otherwise healthy)

1. **Skip the k3s addon on every server node** (node1–node3):

   ```bash
   for n in node1 node2 node3; do
     ssh $n 'sudo touch /var/lib/rancher/k3s/server/manifests/coredns.yaml.skip'
   done
   ```

   No k3s restart needed for the skip to prevent future re-applies; existing
   resources keep running.

2. **Activate in Git** — move the manifests into the scanned directory:

   ```bash
   git mv core-components/coredns/takeover/0*.yaml core-components/coredns/manifests/
   git commit -m "feat: take GitOps ownership of CoreDNS (2 replicas, PDB)"
   git push
   ```

3. **Watch the sync**: `coredns-manifest` should adopt the existing objects
   and roll the Deployment to 2 replicas spread over 2 nodes:

   ```bash
   kubectl get pods -n kube-system -l k8s-app=kube-dns -o wide -w
   ```

4. **Verify DNS from a throwaway pod**:

   ```bash
   kubectl run dnstest --rm -it --image=busybox:1.36 --restart=Never -- \
     nslookup github.com
   ```

## Caveats

- **NodeHosts**: k3s patches the `NodeHosts` key of the `coredns` ConfigMap
  when nodes join/leave. The ApplicationSet has a scoped `ignoreDifferences`
  for that key so ArgoCD and k3s don't fight over it; when adding a node,
  also update `NodeHosts` in Git so the repo stays truthful.
- **k3s upgrades**: the `.skip` file persists across upgrades, but verify it
  still exists after upgrading k3s (`ls /var/lib/rancher/k3s/server/manifests/`).
- **CoreDNS version**: now pinned in Git (`rancher/mirrored-coredns-coredns:1.12.1`).
  k3s upgrades will no longer bump it for you — review CoreDNS releases when
  upgrading k3s.
