## 1.3.0 (unreleased)

FEATURES:

- **New Resource:** `porkbun_dnssec_record`

## 1.2.0 (2025-05-02)

> As of this release, the provider is
> now [available on the OpenTofu registry](https://search.opentofu.org/provider/marcfrederick/porkbun).

FEATURES:

- **New Data Source:** `porkbun_domains`
- **New Ephemeral Resource:** `porkbun_ssl`

## 1.1.1 (2025-04-29)

BUG FIXES:

- resource/porkbun_dns_record: Fix an issue where the `subdomain` field was sometimes set incorrectly when reading
  records.

## 1.1.0 (2025-04-28)

FEATURES:

- provider: Add `max_retries` config option for retrying requests. This should help with rate limiting issues.

## 1.0.1 (2025-04-27)

BUG FIXES:

- resource/porkbun_dns_record: Fix an issue where DNS records could not be imported.

## 1.0.0 (2025-04-26)

FEATURES:

- **New Data Source:** `porkbun_domain`
- **New Data Source:** `porkbun_nameservers`
- **New Data Source:** `porkbun_ssl`
- **New Resource:** `porkbun_dns_record`
- **New Resource:** `porkbun_nameservers`
- **New Resource:** `porkbun_url_forward`
