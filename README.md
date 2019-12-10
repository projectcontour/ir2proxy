# ir2proxy [![Build Status](https://travis-ci.com/projectcontour/ir2proxy.svg?branch=master)](https://travis-ci.com/projectcontour/ir2proxy) [![Go Report Card](https://goreportcard.com/badge/github.com/projectcontour/ir2proxy)](https://goreportcard.com/report/github.com/projectcontour/ir2proxy) ![GitHub release](https://img.shields.io/github/release/projectcontour/ir2proxy.svg) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

ir2proxy is a tool to convert ir2proxy's IngressRoute resources to HTTPProxy resources.

## Usage

`ir2proxy` is intended for taking a yaml file containing one or more valid IngressRoute objects, and then outputting translated HTTPProxy objects to stdout.

Logging is done to stderr.

To use the tool, just run it with a filename as input.

```sh
$ ir2proxy basic.ingressroute.yaml
---
apiVersion: projectcontour.io/v1
kind: HTTPProxy
metadata:
  name: basic
  namespace: default
spec:
  routes:
  - conditions:
    - prefix: /
    services:
    - name: s1
      port: 80
  virtualhost:
    fqdn: foo-basic.bar.com
status:
  currentStatus: ""
  description: ""
```

## Installation

Go to the [releases](https://github.com/projectcontour/ir2proxy/releases) page and download the latest version.
