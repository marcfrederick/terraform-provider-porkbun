resource "porkbun_nameservers" "example" {
  domain      = "example.com"
  nameservers = ["ns1.example.com", "ns2.example.com"]
}
