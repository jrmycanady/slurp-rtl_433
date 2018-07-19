# slurp-rtl_433
slurp-rtl_433 reads the output of rtl_433 and stores it into an InfluxDB database. The output of rtl_433 should be sent to a file that slurp will monitor. If connectivity to the database is lost the local file will continue to collect data until slurp is able to send it. 


## slurp-rtl_433 service
/usr/bin/slurp-rtl_433
/etc/default/slurp-rtl_433
/etc/slurp-rtl_433/conifg.toml
/var/logs/slurp-rtl_433/*log


## rtl_433 service
/usr/bin/start_rtl_433.sh
/etc/default/rtl_433
/var/log/rtl_433/data/rtl_433_data.log
/var/log/rtl_433/rtl_433.log