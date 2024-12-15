# Multipass Hosts

Small utility that generates or updates the hosts file with your multipass vm
names and ip addresses. This allows accessing the vm locally using its name
instead of its dynamically assigned ip address. Additionally its possible to
define aliasses that allow a vm to be accessible using multiple names. When
multipass assigns new ip addresses to vms you can rerun the tool to update the
hosts configuration.

## Installation

The utility can be installed when Go is available using:

```bash
go install github.com/sanderhahn/multipass-hosts@latest
```

On Linux the tool needs to be executed as root so that it is able to overwrite
the `/etc/hosts` file:

```bash
# always run multipass-hosts as root
sudo chmod +s `which multipass-hosts`
sudo chown root:root `which multipass-hosts`
```

On macOS/Darwin the
[System Integrity Protection (SIP)](https://developer.apple.com/documentation/security/disabling-and-enabling-system-integrity-protection)
which will prevent the tool from updating the `/etc/hosts` file, however it
prints out how the file can be updated manually.

On Windows it updates the `$Env:SystemRoot\System32\drivers\etc\hosts` file when
executed as Administrator.

## Aliasses

This is possible by using aliasses that are read from the
`$HOME/.multipass-hosts.json` file using the format:

```json
{
  "aliasses": {
    "gitlab": [
      "gitlab.example.com",
      "example.io",
      "root.example.io",
      "repository.example.com"
    ]
  }
}
```

## Implementation

The tool executes `multipass list --format json` and extracts the name and ipv4
fields. These ip addresses are added into the `hosts` file surrounded with
comments that mark the start and end of the block.
