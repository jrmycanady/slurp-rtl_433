# slurp-rtl_433
slurp-rtl_433 is a simple executable that dumps data from [rtl_433](https://github.com/merbanan/rtl_433) to [InfluxDB](https://www.influxdata.com/time-series-platform/influxdb/) with plans to support [Elasticsearch](https://www.elastic.co/products). 

## Usage
To use slurp-rtl_433, rtl_433 must be redirecting the json format output to a file. The file can be anywhere and rotated with logrotate. The defaut location and name is ./rtl_433_data.log or 



## Overview


## slurp-rtl_433 service
/usr/bin/slurp-rtl_433
/etc/default/slurp-rtl_433
/etc/slurp-rtl_433/conifg.toml
/var/logs/slurp-rtl_433/slurp-rtl_433.log
/var/lib/slurp-rtl_433/meta/


## rtl_433 service
/usr/bin/start_rtl_433.sh
/etc/default/rtl_433
/var/logs/rtl_433/data/rtl_433_data.log
/var/logs/rtl_433/rtl_433.log