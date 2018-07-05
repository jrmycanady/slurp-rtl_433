# slurp-rtl_433
slurp-rtl_433 reads the output of rtl_433 and stores it into an InfluxDB database. The output of rtl_433 should be sent to a file that slurp will monitor. If connectivity to the database is lost the local file will continue to collect data until slurp is able to send it. 

