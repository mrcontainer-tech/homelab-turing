# Longhorn 

To get Longhorn to work on a Turing Pi with K3s:

- Vendor the helm chart in the chart directory.
- Disable the pre upgrade check job, or else it will fail as there is no service account present.
- Have a node that has a disk:
  - I have a SSD on node4 mounted under /mnt/ssd
  
If this happens:

```
Defaulted container "longhorn-manager" out of: longhorn-manager, pre-pull-share-manager-image
time="2025-06-27T07:45:37Z" level=fatal msg="Error starting manager: failed to check environment, please make sure you have iscsiadm/open-iscsi installed on the host: failed to execute: /usr/bin/nsenter [nsenter --mount=/host/proc/3919/ns/mnt --net=/host/proc/3919/ns/net iscsiadm --version], output , stderr nsenter: failed to execute iscsiadm: No such file or directory\n: exit status 127" func=main.main.DaemonCmd.func3 file="daemon.go:105"
```

install open-iscsi on the node. In this case I will install it on all the nodes. 

```
sudo apt-get update
sudo apt-get install open-iscsi
sudo systemctl enable --now iscsid
```


change in the values.yaml

```
createDefaultDiskLabeledNodes: false
```
