# ksei-exporter
Prometheus exporter for KSEI financial data as metrics. Using [GoKSEI](https://github.com/chickenzord/goksei) library for API client.

## Features

- Aggregating multiple KSEI accounts in a single exporter
- Supports Equity, Bond, and Mutual Funds

## Example metrics (redacted)

```
ksei_asset_value{asset_name="GOTO GOJEK TOKOPEDIA Tbk",asset_symbol="GOTO",asset_type="equity",currency="IDR",ksei_account="***@gmail.com",security_account="XL001******",security_name="PT. Stockbit Sekuritas Digital"} 99999
```

## TODO

- [ ] Support cash balance
- [ ] Docker image build
- [ ] Setup instruction
- [ ] Grafana dashboard example
