# Vision

Here I jot down some ideas on the vision I have for my Homelab running on top of my Turing PI and why the hell I am doing what I am doing.

## Why?

First of all I like to tinker and play around with the hardware and the software of the Turing PI and Kubernetes and in extension the CNCF ecosystem. It gives me a super nice opportunity to play around with tools and technologies that I either have used before or I want to learn more about.

This cluster is meant to be a playground and a nice way todo things I normally wouldn't or can't do in any customer environments. For example how many times do I get to play around with Raspberry Pi's, MetalLB, Treafik or physical hardware.

Also its a great showcase of what I know about Kubernetes and how I deliver work to customers and in this case the customer is myself.

## What?

This Kubernetes cluster on top of my Turing PI is off course over-engineered and pretty complex but the following core-components (needed to run the applications or workloads in a proper manner) and applications or workloads are or will be included.


### Kubernetes distribution

### Talos

The Kubernetes distribution I chose is Talos. Talos is built on top of the Linux kernel and uses a container runtime to run Kubernetes components. It only exposes the Kubernetes API over HTTPS, making it immutable and minimal by design — no SSH, no package manager, no shell. This makes it incredibly secure and reproducible, which aligns well with the GitOps philosophy of this cluster. Talos runs great on Raspberry Pi's, making it a natural fit for the Turing Pi.

#### Alternatives considered

**K3s** was an option I considered early on due to its lightweight footprint and familiarity, but I moved away from it in favor of the stricter, more production-like operational model that Talos provides.

**MicroK8s** is another lightweight Kubernetes distribution, but it didn't offer the same security posture or declarative configuration model as Talos.

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

For Serverless & Functions:

- Knative Serving (CNCF graduated serverless platform)

### Applications

Applications are the end-user workloads and services running on the cluster. These showcase practical use cases and development workflows.
