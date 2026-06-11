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

## Backups (runbook)

Backup config lives in `values.yaml` (`defaultSettings.backupTarget`),
`manifests/backup-external-secret.yaml`, and `manifests/recurring-backup.yaml`.
The CNPG databases (linkding/mealie/miniflux) reuse the same bucket and
credentials via `backup-external-secret.yaml` in their own manifests.

**Run these BEFORE merging the backup config** — CNPG WAL archiving starts
failing (and accumulating WAL on the 2Gi volumes) if the bucket or secret
is missing:

```bash
# 1. Bucket (name is referenced in values.yaml and the pg-cluster.yaml files;
#    if you pick another name, update those references in the same commit)
aws s3 mb s3://mrcontainer-homelab-backups --region eu-west-1

# 2. Dedicated IAM user, scoped to the bucket
aws iam create-user --user-name homelab-backup
aws iam put-user-policy --user-name homelab-backup --policy-name backup-bucket-rw \
  --policy-document '{
    "Version": "2012-10-17",
    "Statement": [
      {"Effect": "Allow",
       "Action": ["s3:PutObject","s3:GetObject","s3:DeleteObject","s3:ListBucket","s3:GetBucketLocation"],
       "Resource": ["arn:aws:s3:::mrcontainer-homelab-backups","arn:aws:s3:::mrcontainer-homelab-backups/*"]}
    ]}'
aws iam create-access-key --user-name homelab-backup
# → note AccessKeyId / SecretAccessKey from the output

# 3. Store the keys in Secrets Manager (key names match the ExternalSecrets)
aws secretsmanager create-secret --region eu-west-1 --name homelab/backup-s3 \
  --secret-string '{"access-key-id":"<AccessKeyId>","secret-access-key":"<SecretAccessKey>"}'
```

**Verify after ArgoCD syncs:**

```bash
kubectl get backuptarget -n longhorn-system          # AVAILABLE should be true
kubectl get recurringjob -n longhorn-system
kubectl get scheduledbackup,backup -A                # CNPG backups after 03:00
kubectl get cluster -A -o wide                       # ContinuousArchiving condition
```

Notes:
- Longhorn ≥1.8 manages the target through the `default` BackupTarget CR,
  fed by the Helm `defaultSettings`. If the URL doesn't take effect after a
  chart upgrade (known Longhorn quirk for settings changed post-install),
  set it once via the Longhorn UI — the values stay in Git as the record.
- CNPG's in-tree `barmanObjectStore` is deprecated upstream in favour of the
  Barman Cloud plugin; it still works on the current operator (chart 0.28).
  Revisit when upgrading CNPG.
- Restore procedures: Longhorn → restore from backup in UI (new PVC);
  CNPG → `bootstrap.recovery` from the object store into a new Cluster.
