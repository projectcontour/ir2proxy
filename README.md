# ir2proxy [![Build Status](https://travis-ci.com/projectcontour/ir2proxy.svg?branch=main)](https://travis-ci.com/projectcontour/ir2proxy) [![Go Report Card](https://goreportcard.com/badge/github.com/projectcontour/ir2proxy)](https://goreportcard.com/report/github.com/projectcontour/ir2proxy) ![GitHub release](https://img.shields.io/github/release/projectcontour/ir2proxy.svg) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

ir2proxy is a tool to convert ir2proxy's IngressRoute resources to HTTPProxy resources.

## Features

ir2proxy can translate an IngressRoute object to a HTTPProxy object.
The full featureset of IngressRoute should be translated correctly.
If not, please [log an issue](https://github.com/projectcontour/ir2proxy/issues), specifying what didn't work and supplying the sanitized IngressRoute YAML.

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
status: {}
```

It's intended mode of operation is in a one-file-at-a-time manner, so it's easier to use it in a Unix pipe.

## Installation

### Homebrew

Add the `ir2proxy` tap (located at [homebrew-ir2proxy](https://github.com/projectcontour/homebrew-ir2proxy)):

```sh
brew tap projectcontour/ir2proxy
```

Then install ir2proxy:

```sh
brew install ir2proxy
```

To upgrade, use

```sh
brew upgrade ir2proxy
```

### Non-homebrew

Go to the [releases](https://github.com/projectcontour/ir2proxy/releases) page and download the latest version.

## Possible issues with conversion and what to do about them

### Prefix behavior in IngressRoute vs HTTPProxy

In IngressRoute, delegation was a route-level construct, that required that the delegated IngressRoutes have the full prefix, including the delegation prefix.
So a nonroot Ingressroute that wanted to accept traffic for `/foo/bar` would have a `match` entry of `/foo/bar`.

For HTTPProxy, inclusion is a top-level construct, and the included HTTPProxy does *not* need to have the full prefix, and can be included at multiple paths if required.
So a nonroot HTTPProxy that wanted to accept traffic for `/foo/bar` would have a `prefix` `condition` of `/bar`, and be included using a `prefix` `condition` of `/foo`.

`ir2proxy` tries to guess what the prefix should be, and puts its guess into generated nonroot HTTPProxy objects.
It will warn you on stderr and in the generated file what its guess means if it's not sure.
(For some specific cases, the tool can be sure what you mean.)

### Load Balancing Strategy

In IngressRoute, setting the load balancing strategy was originally designed as a route-level default that could be overwritten by a service-level setting.
However, only the service-level setting was implemented.

HTTPProxy currently only has the route-level setting implemented, so `ir2proxy` will take the first setting of `strategy` in IngressRoute to be the correct setting for HTTPProxy.

A warning will be output to stderr and as a comment in the file.

### Healthchecks

In IngressRoute, healthchecks were only configurable at a service level, not defaulted at a route level.

In HTTPProxy, healtchchecks are only configurable at a route level.
Accordingly, `ir2proxy` will take overwrite the healthcheck found and record it at the HTTPProxy Route level.
This means that for multiple healthchecks, the last will take precedence.

A warning will be output to stderr and as a comment in the file.
