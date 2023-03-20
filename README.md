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

- [Amazon Route 53](https://aws.amazon.com/route53/)
- [DNSimple](https://dnsimple.com/)

## Deployment

### Helm

Proclaim can be deployed using the Helm chart in this repository.
The default values are suitable for most deployments.

```bash
# configure the "proclaim" secret (see below)
helm pull https://github.com/dogmatiq/proclaim/tree/main/charts dogmatiq/proclaim
helm install proclaim --values values.yaml dogmatiq/proclaim
```

Please note that pulling the Helm chart from the `main` branch will always
install the latest version. `main` can be replaced with a version number (e.g.
`v0.3.0`) to install that version.

## Configuration

Multiple DNS providers can be enabled at once by setting the `enabled` value to
`true` in the relevant `providers` section of the Helm chart [values file].

Each provider needs its own credentials which are stored in a Kubernetes secret
named `proclaim`. This secret is **NOT** created by the Helm chart.

### Amazon Route 53

Set the `proclaim.providers.route53.enabled` value to `true` in the Helm chart
[values file].

[IRSA] is recommended when running under EKS. Proclaim creates a service account
which can be annotated with IAM-specific annotations by setting the
`proclaim.serviceAccount.annotations` value in the Helm chart [values file].

Otherwise, the standard AWS environment variables (`AWS_ACCESS_KEY_ID`, etc) can
be added to the `proclaim` secret.

The [example IAM policy] illustrates the minimum set of permissions required for
Proclaim to function.

### DNSSimple

Set the `proclaim.providers.dnsimple.enabled` value to `true` in the Helm chart
[values file].

Add a `DNSIMPLE_TOKEN` key to the `proclaim` secret. The token can be either a
"user" token or an "account" token.

<!-- references -->

[dns-sd]: https://www.rfc-editor.org/rfc/rfc6763
[amazon route53]: https://aws.amazon.com/route53/
[dnsimple.com]: https://dnsimple.com/
[irsa]: https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html
[values file]: charts/values.yaml
[example iam policy]: examples/iam/policy.json
