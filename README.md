# slurp-rtl_433
slurp-rtl_433 reads the output of rtl_433 and stores it into an InfluxDB database. The output of rtl_433 should be sent to a file that slurp will monitor. If connectivity to the database is lost the local file will continue to collect data until slurp is able to send it. 

## Usage
To run slrup-rtl_433 you need to have an InfluxDB online and accepting connections. At the minimum the database must already be created.

`slurp-rtl_433 -c /path/to/config/file`

### Parameters

**NOTE:** All individual command line parameters supersede the configuration file parameters.

|Name|Flag|Description|Default|
|----|----|-----------|-------|
|--config|-c|The path to the config file.||
|--data-location|-d|The location and search string for data files.|/rtl_433.out*|
|--fqdn|-f|The FQDN to the InfluxDB server.|localhost|
|--port|-P|The port to the InfluxDB server.|8086|
|--username|-u|The username to connect to InfluxDB with.||
|--password|-p|The password to connect to InfluxDB with.||
|--database|-d|The name of the database to send data to.|slurp-rtl_433|
|--verbose|-v|Enable verbose logging.|false|
|--debug|-D|Enable debug logging.|false|



