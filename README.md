# ksei-exporter
Prometheus exporter for KSEI financial data as metrics. Using [GoKSEI](https://github.com/chickenzord/goksei) library for API client.

[![Go Report Card](https://goreportcard.com/badge/github.com/chickenzord/ksei-exporter)](https://goreportcard.com/report/github.com/chickenzord/ksei-exporter)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/chickenzord/ksei-exporter)
![Go Build](https://github.com/chickenzord/ksei-exporter/actions/workflows/go.yml/badge.svg?branch=main)
![Docker Build](https://github.com/chickenzord/ksei-exporter/actions/workflows/docker.yml/badge.svg?branch=main)
![Code License](https://img.shields.io/github/license/chickenzord/ksei-exporter)


## Features

- Aggregating multiple KSEI accounts in a single exporter
- Supports Equity, Bond, and Mutual Funds

## Example metrics (redacted)

```
ksei_asset_value{asset_name="GOTO GOJEK TOKOPEDIA Tbk",asset_symbol="GOTO",asset_type="equity",currency="IDR",ksei_account="***@gmail.com",security_account="XL001******",security_name="PT. Stockbit Sekuritas Digital"} 99999
```

## Configuration

ksei-exporter is configured using enviroment variables:

```sh
SERVER_BIND_HOST=0.0.0.0
SERVER_BIND_PORT=8080

KSEI_ACCOUNTS="
john.doe@example.com:johnsaltedpassword
jane.doe@example.com:janesaltedpassword
"
KSEI_AUTH_DIR=.goksei.auth
KSEI_REFRESH_INTERVAL=1h
KSEI_REFRESH_JITTER=0.2

```

## TODO

- [x] Support cash balance
- [x] Docker image build
- [ ] Setup instruction
- [ ] Grafana dashboard example
