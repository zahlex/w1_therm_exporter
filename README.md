# w1_therm_exporter
Exposes temperatures read via w1_therm linux kernel module from any 1-Wire bus for Prometheus consumption

You need to have [go installed](https://golang.org/doc/install) on your linux system.

```
git clone https://github.com/zahlex/w1_therm_exporter
cd w1_therm_exporter
go build
./w1_therm_exporter -httpAddr=:8080
```

Default HTTP Port is 8080, accessible from all interfaces and subnets.

Licence: MIT