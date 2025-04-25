# Terraform Provider Porkbun

This Terraform provider lets you automate the management of Porkbun domains, DNS records, and other related resources.

## Contributing

Contributions are welcome! If you have suggestions or improvements, please open an issue or a pull request.

### Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.23

### Building the Provider

1. Clone the repository.
2. Navigate into the repository directory.
3. Build the provider using the Go `install` command:

```shell
go install
```

### Acceptance Testing

To run acceptance tests, you need:

- A valid Porkbun API key and secret
- A registered domain for testing

Set the required environment variables:

```bash
export PORKBUN_API_KEY="your_api_key"
export PORKBUN_SECRET_API_KEY="your_secret_api_key"
export PORKBUN_ACCTEST_DOMAIN="example.com"
```

Run the tests with:

```bash
make acctest
```

> ⚠️ During testing, the provider will create and destroy resources in the domain specified by `PORKBUN_ACCTEST_DOMAIN`.
> Use a test domain or a domain you can safely modify.

## Using the Provider Locally

To test the provider locally, configure Terraform to use your local build by adding the following to your
`~/.terraformrc` file.
Replace `<GOPATH>` with your actual Go path, which you can find by running `go env GOPATH`:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/marcfrederick/porkbun" = "<GOPATH>/bin"
  }
  direct {}
}
```

Then, create a new Terraform configuration file (e.g., `main.tf`) with the following:

```hcl
terraform {
  required_providers {
    porkbun = {
      source  = "marcfrederick/porkbun"
      version = ">= 0.1.0"
    }
  }
}

provider "porkbun" {}
```

## Related Projects

Other existing Terraform providers for Porkbun support different subsets of the API. These providers rely on less
feature-rich or custom API clients, making them harder to extend compared to implementing a new provider.

- [cullenmcdermott/porkbun](https://registry.terraform.io/providers/cullenmcdermott/porkbun)
  - `porkbun_dns_record` (Resource)
- [kyswtn/porkbun](https://registry.terraform.io/providers/kyswtn/porkbun)
  - `porkbun_dns_record` (Resource)
  - `porkbun_nameservers` (Resource)
  - `porkbun_nameservers` (Data Source)

## Acknowledgements

* [tuzzmaniandevil](github.com/tuzzmaniandevil) for the [Porkbun API client](github.com/tuzzmaniandevil/porkbun-go).
* [HashiCorp](https://www.hashicorp.com) for
  the [Terraform Plugin Framework](github.com/hashicorp/terraform-plugin-framework)
  and [Terraform Provider Development Guide](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider).
