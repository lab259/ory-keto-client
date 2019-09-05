# ory-keto-client [![Actions Status](https://github.com/lab259/ory-keto-client/workflows/Go/badge.svg)](https://github.com/lab259/ory-keto-client/actions)

## Getting Started

`ory-keto-client` is the [Lab259](https://github.com/lab259) implementation of
the [ORY Keto](https://github.com/ory/keto) client library.

This client library uses [gojek/heimdall](https://github.com/gojek/heimdall)
to make the requests. `Heimdall` uses [afex/hystrix-go](https://github.com/afex/hystrix-go),
a [netflix/Hystrix](https://github.com/netflix/Hystrix) implementation in Go, to
provide retriers and circuit breaker. Check [here](https://github.com/netflix/Hystrix/wiki)
to see why is this important.

### Usage

TODO

### Prerequisites

What things you need to setup the project:

- [go](https://golang.org/doc/install)
- [ginkgo](http://onsi.github.io/ginkgo/)

### Running tests

First, we have to bring up the dependencies:

```bash
docker-compose up -d
```

Then we are able to run the tests:

```bash
make test
```

To enable coverage, execute:

```bash
make coverage
```

To generate the HTML coverage report, execute:

```bash
make coverage coverage-html
```
