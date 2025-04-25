resource "porkbun_dns_record" "example" {
  domain    = "example.com"
  subdomain = "www"
  type      = "A"
  content   = "1.1.1.2"
  ttl       = 600
  prio      = 10
}
