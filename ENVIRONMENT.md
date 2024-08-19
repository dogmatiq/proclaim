# Environment Variables

This document describes the environment variables used by `proclaim`.

| Name                 | Usage                                  | Description                      |
| -------------------- | -------------------------------------- | -------------------------------- |
| [`DNSIMPLE_API_URL`] | defaults to `https://api.dnsimple.com` | the URL of the DNSimple API      |
| [`DNSIMPLE_ENABLED`] | defaults to `false`                    | enable the DNSimple provider     |
| [`DNSIMPLE_TOKEN`]   | conditional                            | enable the DNSimple provider     |
| [`ROUTE53_ENABLED`]  | defaults to `false`                    | enable the AWS Route 53 provider |

> [!TIP]
> If an environment variable is set to an empty value, `proclaim` behaves as if
> that variable is left undefined.

## `DNSIMPLE_API_URL`

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

### See Also

- [`DNSIMPLE_ENABLED`] — enable the DNSimple provider

## `DNSIMPLE_ENABLED`

> enable the DNSimple provider

The `DNSIMPLE_ENABLED` variable **MAY** be left undefined, in which case the
default value of `false` is used. Otherwise, the value **MUST** be either `true`
or `false`.

```bash
export DNSIMPLE_ENABLED=true
export DNSIMPLE_ENABLED=false # (default)
```

## `DNSIMPLE_TOKEN`

> enable the DNSimple provider

The `DNSIMPLE_TOKEN` variable **MAY** be left undefined if and only if
[`DNSIMPLE_ENABLED`] is `false`.

⚠️ This variable is **sensitive**; its value may contain private information.

### See Also

- [`DNSIMPLE_ENABLED`] — enable the DNSimple provider

## `ROUTE53_ENABLED`

> enable the AWS Route 53 provider

The `ROUTE53_ENABLED` variable **MAY** be left undefined, in which case the
default value of `false` is used. Otherwise, the value **MUST** be either `true`
or `false`.

```bash
export ROUTE53_ENABLED=true
export ROUTE53_ENABLED=false # (default)
```

---

> [!NOTE]
> This document only describes environment variables declared using [Ferrite].
> `proclaim` may consume other undocumented environment variables.

> [!IMPORTANT]
> Some of the example values given in this document are **non-normative**.
> Although these values are syntactically valid, they may not be meaningful to
> `proclaim`.

<!-- references -->

[`dnsimple_api_url`]: #DNSIMPLE_API_URL
[`dnsimple_enabled`]: #DNSIMPLE_ENABLED
[`dnsimple_token`]: #DNSIMPLE_TOKEN
[ferrite]: https://github.com/dogmatiq/ferrite
[`route53_enabled`]: #ROUTE53_ENABLED
