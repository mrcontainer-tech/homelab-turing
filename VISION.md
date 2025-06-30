# Vision

Here I jot down some ideas on the vision I have for my Homelab running on top of my Turing PI and why the hell I am doing what I am doing.

## Why?

First of all I like to tinker and play around with the hardware and the software of the Turing PI and Kubernetes and in extension the CNCF ecosystem. It gives me a super nice opportunity to play around with tools and technologies that I either have used before or I want to learn more about.

This cluster is meant to be a playground and a nice way todo things I normally wouldn't or can't do in any customer environments. For example how many times do I get to play around with Raspberry Pi's, MetalLB, Treafik or physical hardware.

Also its a great showcase of what I know about Kubernetes and how I deliver work to customers and in this case the customer is myself.

## What?

This Kubernetes cluster on top of my Turing PI is off course over-engineered and pretty complex but the following core-components (needed to run the applications or workloads in a proper manner) and applications or workloads are or will be included.


### Kubernetes distribution

To start with Kubernetes I choose to use K3s.

### K3s

The Kubernetes distrubution I choose is K3s, as I wanted to still have the ability to SSH into the nodes and troubleshoot or experiment with the cluster. Also I have experience with K3s and I feel its a great lightweight Kubernetes distribution. Other distributions I have considered are Talos and MicroK8s.

#### Talos

Talos is built on top of the Linux kernel and uses a container runtime to run Kubernetes components. Talos only exposes the Kubernetes API over HTTPS. The more I read about it and learn about it, the more I like it actually. In the future I might consider moving towards it, as Talos can also run on Raspberry Pi's.

#### MicroK8s

MicroK8s just like K3s is a super lightweight Kube distribution. Reason I didn't went for this one is mostly because I wanted to quickly spin up a Kubernetes cluster and the Turing PI docs recommended K3s.

### Core Components

Core components are essential building blocks on the Homelab, that make sure it is possible to run applications or workloads in a proper and secure manner.

For GitOps:

- ArgoCD

For networking and ingress the following will be used:
 
- MetalLB
- Treafik (Default of K3s)

For Security:

- cert-manager (Route53)
- Kyverno
- Harbor
- external-secrets

For monitoring:

- Prometheus
- Grafana

For block storage and Disaster Recovery:

- Longhorn (S3)

For Databases:

- Enterprise Postgres Operator

For DNS:

- CoreDNS
- ExternalDNS (Route53)

### Applications

Applicatio
