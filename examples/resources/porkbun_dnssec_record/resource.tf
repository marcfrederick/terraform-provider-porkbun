resource "porkbun_dnssec_record" "example" {
  domain       = "example.com"
  max_sig_life = 86400

  ds_data = {
    key_tag     = "64087"
    algorithm   = 13
    digest_type = 2
    digest      = "15E445BD08128BDC213E25F1C8227DF4CB35186CAC701C1C335B2C406D5530DC"
  }
}
