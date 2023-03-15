# Environment Variables

This document describes the environment variables used by `proclaim`.

If any of the environment variable values do not meet the requirements herein,
the application will print usage information to `STDERR` then exit with a
non-zero exit code. Please note that **undefined** variables and **empty**
values are considered equivalent.

⚠️ This document includes **non-normative** example values. While these values
are syntactically correct, they may not be meaningful to this application.

⚠️ The application may consume other undocumented environment variables; this
document only shows those variables declared using [Ferrite].

The key words **MUST**, **MUST NOT**, **REQUIRED**, **SHALL**, **SHALL NOT**,
**SHOULD**, **SHOULD NOT**, **RECOMMENDED**, **MAY**, and **OPTIONAL** in this
document are to be interpreted as described in [RFC 2119].

## Index

- [`DNSIMPLE_API_URL`] — the URL of the DNSimple API
- [`DNSIMPLE_ENABLED`] — enable the DNSimple provider
- [`DNSIMPLE_TOKEN`] — enable the DNSimple provider
- [`ROUTE53_ENABLED`] — enable the AWS Route 53 provider

## Specification

### `DNSIMPLE_API_URL`

> the URL of the DNSimple API

The `DNSIMPLE_API_URL` variable **MAY** be left undefined, in which case the
default value of `https://api.dnsimple.com` is used. Otherwise, the value
**MUST** be a fully-qualified URL. The value is not used when
[`DNSIMPLE_ENABLED`] is `false`.

```bash
export DNSIMPLE_API_URL=https://api.dnsimple.com # (default)
export DNSIMPLE_API_URL=https://example.org/path # (non-normative) a typical URL for a web page
```

<details>
<summary>URL syntax</summary>

A fully-qualified URL includes both a scheme (protocol) and a hostname. URLs are
not necessarily web addresses; `https://example.org` and
`mailto:contact@example.org` are both examples of fully-qualified URLs.

</details>

#### See Also

- [`DNSIMPLE_ENABLED`] — enable the DNSimple provider

### `DNSIMPLE_ENABLED`

> enable the DNSimple provider

The `DNSIMPLE_ENABLED` variable **MAY** be left undefined, in which case the
default value of `false` is used. Otherwise, the value **MUST** be either `true`
or `false`.

```bash
export DNSIMPLE_ENABLED=true
export DNSIMPLE_ENABLED=false # (default)
```

### `DNSIMPLE_TOKEN`

> enable the DNSimple provider

The `DNSIMPLE_TOKEN` variable **MAY** be left undefined if and only if
[`DNSIMPLE_ENABLED`] is `false`.

```bash
export DNSIMPLE_TOKEN=foo # (non-normative)
```

#### See Also

- [`DNSIMPLE_ENABLED`] — enable the DNSimple provider

### `ROUTE53_ENABLED`

> enable the AWS Route 53 provider

The `ROUTE53_ENABLED` variable **MAY** be left undefined, in which case the
default value of `false` is used. Otherwise, the value **MUST** be either `true`
or `false`.

```bash
export ROUTE53_ENABLED=true
export ROUTE53_ENABLED=false # (default)
```

## Usage Examples

<details>
<summary>Kubernetes</summary>

This example shows how to define the environment variables needed by `proclaim`
on a [Kubernetes container] within a Kubenetes deployment manifest.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-deployment
spec:
  template:
    spec:
      containers:
        - name: example-container
          env:
            - name: DNSIMPLE_API_URL # the URL of the DNSimple API (defaults to https://api.dnsimple.com)
              value: https://api.dnsimple.com
            - name: DNSIMPLE_ENABLED # enable the DNSimple provider (defaults to false)
              value: "false"
            - name: DNSIMPLE_TOKEN # enable the DNSimple provider
              value: foo
            - name: ROUTE53_ENABLED # enable the AWS Route 53 provider (defaults to false)
              value: "false"
```

Alternatively, the environment variables can be defined within a [config map][kubernetes config map]
then referenced from a deployment manifest using `configMapRef`.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: example-config-map
data:
  DNSIMPLE_API_URL: https://api.dnsimple.com # the URL of the DNSimple API (defaults to https://api.dnsimple.com)
  DNSIMPLE_ENABLED: "false" # enable the DNSimple provider (defaults to false)
  DNSIMPLE_TOKEN: foo # enable the DNSimple provider
  ROUTE53_ENABLED: "false" # enable the AWS Route 53 provider (defaults to false)
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-deployment
spec:
  template:
    spec:
      containers:
        - name: example-container
          envFrom:
            - configMapRef:
                name: example-config-map
```

</details>

<details>
<summary>Docker</summary>

This example shows how to define the environment variables needed by `proclaim`
when running as a [Docker service] defined in a Docker compose file.

```yaml
service:
  example-service:
    environment:
      DNSIMPLE_API_URL: https://api.dnsimple.com # the URL of the DNSimple API (defaults to https://api.dnsimple.com)
      DNSIMPLE_ENABLED: "false" # enable the DNSimple provider (defaults to false)
      DNSIMPLE_TOKEN: foo # enable the DNSimple provider
      ROUTE53_ENABLED: "false" # enable the AWS Route 53 provider (defaults to false)
```

</details>

<!-- references -->

[`dnsimple_api_url`]: #DNSIMPLE_API_URL
[`dnsimple_enabled`]: #DNSIMPLE_ENABLED
[`dnsimple_token`]: #DNSIMPLE_TOKEN
[docker service]: https://docs.docker.com/compose/environment-variables/#set-environment-variables-in-containers
[ferrite]: https://github.com/dogmatiq/ferrite
[kubernetes config map]: https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/#configure-all-key-value-pairs-in-a-configmap-as-container-environment-variables
[kubernetes container]: https://kubernetes.io/docs/tasks/inject-data-application/define-environment-variable-container/#define-an-environment-variable-for-a-container
[rfc 2119]: https://www.rfc-editor.org/rfc/rfc2119.html
[`route53_enabled`]: #ROUTE53_ENABLED
