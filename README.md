# narcotk-cloudflare-dynamic-dns 
A simple Cloudflare Dynamic DNS updater.

It detects changes do your external IP and updates Cloudflare DNS.

Suitable for running upon Linux, OSX, RaspberryPi and Ubiquiti Edgerouters.

Can use pushover.net to notify you of any changes.

## Features

- detects changes to your external IP and updates your cloudflare DNS entries
- fully configurable
- can be configured by environment variables, configuration file or from the command line
- compatible with linux, osx, QNAP, raspberry pi, ubiquiti edge routers + more
- clean and lean: a single binary, no mess and no further dependancies
- simple to install and configure
- IPv4 and IPv6 compatible
- multiple external IP detection methods
- supports all cloudflare DNS record types: A, AAAA, CAA, CERT, CNAME, DNSKEY, DS, LOC, MX, NAPTR, NS, PTR, SMIMEA, SPF, SRV, SSHFP, TLSA, TXT, URI
- logging capability
- you can be notified of DNS changes via pushover.net

## Supported Platforms

| Binary | Device/Operating System | Tested On |
| :--- | :--- | :--- |
| binaries/er-cavium | Ubiquiti Cavium based EdgeRouter / EdgeOS.  Models: ERLite-3, ERPoE-5, ER-8, ERPro-8, EP-R8, ER-4, ER-6P, ER-12, ER-12P, ER-8-XG | Tested: ER-4 |
| binaries/er-mediatek | Ubiquiti MediaTek base EdgeRouters / EdgeOS.  Models: ER-X, ER-10X, ER-X-SFP, ER-R6 | Untested |
| binaries/osx | Yosemite, El Capitan, Sierra, High Sierra, Mojave, Catalina | Tested: all |
| binaries/qnap/x86_64 | QNAP x86_64 based systems | Tested: TS-453A |
| binaries/qnap/marvell | QNAP Marvell based systems | Untested |
| binaries/rpi | Raspberry Pi Model B Rev 1 running Rasbian 9.8 | Tested |
| binaries/x86_64 | Ubuntu 18.04.3 LTS | Tested |

## Requirements

- DNS hosted on cloudflare
- Cloudflare API token

## Installation

### Generic Script

### Docker

### Edgerouter

### Git
```
git clone git@github.com:smford/narcotk-cloudflare-dynamic-dns.git
cd narcotk-cloudflare-dynamic-dns
GO111MODULE=on go build narcotk-cf-ddns.go
```


### GO
```
go get -v github.com/smford/narcotk-cloudflare-dynamic-dns
```

### Linux

### OSX

### Other

If you can install go, you can compile and install narcotk-cf-ddns, first try using the Go method mentioned above then the Git method.

### QNAP

### Windows


## Usage
