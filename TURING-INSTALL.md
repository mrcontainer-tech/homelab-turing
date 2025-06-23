# Homelab Turing Pi HA Kubernetes

This guide walks through installing a High‑Availability (HA) K3s Kubernetes cluster on a Turing Pi (v2.5) using four Raspberry Pi CM4 modules.

## 🛠 Prerequisites

* **Hardware**: Turing Pi with 4 CM4 slots, ITX case, power supply
* **BMC access**: `turingpi` CLI and `rpiboot` installed on the BMC
* **MicroSD**: Latest Raspberry Pi OS Lite image
* **Network**: DHCP or static LAN (we use static IPs in this guide)

## 1. Flash OS & Enable SSH/UART

These commands are run from the **BMC**. SSH to it using 

1. **Set the Raspberry to mass storage device and mount the boot partition**

   ```bash
   mount /dev/sda1 /mnt/bootfs
   ```
2. **Enable UART & SSH**

   ```bash
   # Edit config.txt
   vi /mnt/bootfs/config.txt
   # Add:
   enable_uart=1

   # Enable SSH on first boot
   touch /mnt/bootfs/ssh

   # Seed `pi` password via userconf
   echo 'pi:$6$c70VpvPsVNCG0YR5$l5vWWLsLko9Kj65gcQ8qvMkuOoRkEagI90qi3F/Y7rm8eNYZHW8CY6BOIKwMH7a3YYzZYL90zf304cAHLFaZE0' \
     > /mnt/bootfs/userconf
   ```
3. **Unmount & power‑cycle**

   ```bash
   umount /mnt/bootfs
   tpi power -n 1 off
   tpi power -n 1 on
   ```
4. **Verify serial console**

   ```bash
   tpi uart get -n 1
   ```

## 2. Enable cgroup memory

K3s requires the memory cgroup (v2). On each node:

```bash
sudo cp /boot/firmware/cmdline.txt /boot/firmware/cmdline.txt.bak
sudo sed -i 's/$/ cgroup_enable=memory cgroup_memory=1/' /boot/firmware/cmdline.txt
sudo reboot
```

## 3. Set Hostnames

On each node, replace `N` with the node number (1–4):

```bash
sudo hostnamectl set-hostname nodeN
```

## 4. Configure Static IPs (NetworkManager)

Why static IPs? K3s (written in Go) uses its own DNS resolver which does not honor mDNS/Avahi, so any .local hostnames advertised only via Avahi will appear unresolvable to K3s. By assigning fixed static IP addresses (or mapping names in /etc/hosts), you ensure the API server and peers can always be reached reliably.

On **each** Pi (node1–node4):

```bash
# Identify your connection name:
nmcli connection show

# Modify it (example for node1):
sudo nmcli connection modify "Wired connection 1" \
  ipv4.method manual \
  ipv4.addresses 192.168.68.10N/24 \
  ipv4.gateway 192.168.68.1 \
  ipv4.dns "192.168.68.1 8.8.8.8" \
  connection.autoconnect yes

# Bring it up:
sudo nmcli connection up "Wired connection 1"
```

Nodes:

* node1 → `.101`
* node2 → `.102`
* node3 → `.103`
* node4 → `.104`

## 5. Install K3s (HA control‑plane)

### Node 1 (first master)

```bash
curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="server \
  --cluster-init \
  --token <YOUR_TOKEN> \
  --disable servicelb \
  --disable local-storage \
  --disable cloud-controller \
  --tls-san 192.168.68.101 \
  --tls-san 192.168.68.102 \
  --tls-san 192.168.68.103 \
  --tls-san 192.168.68.104 \
  --tls-san node1.local \
  --tls-san node2.local \
  --tls-san node3.local \
  --tls-san node4.local" sh -
```

### Node 2 & 3 (additional masters)

Retrieve the **server token** on node1:

```bash
cat /var/lib/rancher/k3s/server/token
```

On **node2** and **node3**:

```bash
curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="server \
  --server https://192.168.68.101:6443 \
  --token <YOUR_TOKEN> \
  --disable servicelb \
  --disable local-storage \
  --disable cloud-controller" sh -
```

### Node 4 (agent)

The fourth node will function as a worker node, kubernetes works best in HA when there are at least three masters.

```bash
curl -sfL https://get.k3s.io | \
  K3S_URL="https://192.168.68.101:6443" \
  K3S_TOKEN="<YOUR_TOKEN>" \
  sh -
```

## 6. Label Nodes

We want each node to function as a worker node:

```bash
kubectl label node node1 node-role.kubernetes.io/worker=""
kubectl label node node2 node-role.kubernetes.io/worker=""
kubectl label node node3 node-role.kubernetes.io/worker=""
kubectl label node node4 node-role.kubernetes.io/worker=""
```

## 7. Install ArgoCD

```bash
kubectl create namespace argocd
kubectl apply -n argocd -f \
  https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```