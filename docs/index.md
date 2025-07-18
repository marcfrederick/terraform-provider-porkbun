---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "porkbun Provider"
description: |-
  Provider for managing domains, DNS records, URL forwarding, and nameserver configurations for domains registered with Porkbun.
---

# porkbun Provider

Provider for managing domains, DNS records, URL forwarding, and nameserver configurations for domains registered with Porkbun.

## Example Usage

```terraform
provider "porkbun" {
  api_key        = "pk1_********"
  secret_api_key = "sk1_********"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `api_key` (String, Sensitive) API key for authentication. Can also be set using the `PORKBUN_API_KEY` environment variable.
- `ipv4_only` (Boolean) Use IPv4 only for API requests. Defaults to false.
- `max_retries` (Number) Maximum number of retries for API requests. Defaults to 3.
- `secret_api_key` (String, Sensitive) Secret API key for authentication. Can also be set using the `PORKBUN_SECRET_API_KEY` environment variable.
