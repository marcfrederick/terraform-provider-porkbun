resource "porkbun_url_forward" "example" {
  domain       = "example.com"
  subdomain    = "www"
  location     = "test.com"
  type         = "temporary"
  include_path = false
  wildcard     = false
}
