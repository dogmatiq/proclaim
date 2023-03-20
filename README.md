<div align="center">

# Proclaim

A Kubernetes controller and CRD that publishes DNS-SD records.

[![Latest Version](https://img.shields.io/github/tag/dogmatiq/proclaim.svg?&style=for-the-badge&label=semver)](https://github.com/dogmatiq/proclaim/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/dogmatiq/proclaim/ci.yml?style=for-the-badge&branch=main)](https://github.com/dogmatiq/proclaim/actions/workflows/ci.yml)
[![Code Coverage](https://img.shields.io/codecov/c/github/dogmatiq/proclaim/main.svg?style=for-the-badge)](https://codecov.io/github/dogmatiq/proclaim)

</div>

Proclaim defines a `DNSSDServiceInstance` Kubernetes custom resource that
describes a [DNS-SD] service instance to be advertised on one of the supported
DNS provider implementations:

- AWS Route53
- DNSimple.com

<!-- references -->

[dns-sd]: https://www.rfc-editor.org/rfc/rfc6763

## Deployment



#### Via a Local Helm Chart

```

helm pull https://github.com/dogmatiq/proclaim/tree/main/charts dogmatiq/proclaim
# Configure values.yaml with your preferences
helm install proclaim --values values.yaml domatiq/proclaim
```

#### Via Terraform

```
TODO

```

## Provider Authentication

#### Route53 Provider

- The best way is to use [IRSA](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html) and a AWS Managed Role
- A tested AWS role is: **arn:aws:iam::aws:policy/AmazonRoute53FullAccess**
- TODO: document a sample IAM Policy

#### DNSSimple Provider

- TODO: document