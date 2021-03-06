# ddcloud\_networkdomain

A Network Domain is the fundamental building block for your MCP 2.0 cloud deployment. At least one Network Domain and one VLAN must be deployed before you can deploy your first Server in an MCP 2.0 data center.

Refer to the documentation for further details:
https://community.opsourcecloud.net/View.jsp?procId=994fa801956149b3861e428801f9888f

**Note:** Due to current platform limitations, organisations that use MCP 2.0 cannot perform more than one concurrent deployment operation for network domains or VLANs (all other operations can however be performed concurrently). If necessary, use the `depends_on` attribute to ensure that resources that relate to the same network domain are not run in parallel.

## Example Usage

```
resource "ddcloud_networkdomain" "my-domain" {
    name                    = "terraform-test-domain"
    description             = "This is my Terraform test network domain."
    datacenter              = "AU9" # The ID of the data centre in which to create your network domain.
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for the network domain.
* `description` - (Optional) A description for the network domain.
* `plan` - (Optional) The plan (service level) for the network domain (`ESSENTIALS` or `ADVANCED` default is `ESSENTIALS`).
* `datacenter` - (Required) The Id of the MCP 2.0 datacenter in which the network domain is created.

## Attribute Reference

The following attributes are exported:

* `nat_ipv4_address` - The IPv4 address for the network domain's IPv6->IPv4 Source Network Address Translation (SNAT). This is the IPv4 address of the network domain's IPv4 egress.
