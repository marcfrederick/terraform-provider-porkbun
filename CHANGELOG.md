## 1.1.2 (unreleased)

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
