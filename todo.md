

# TODO
* Add some sort of cache for failures or wait until they start going through. All others should backup on the channel.
* Add ability to dynamically add tags based on field values of a datapoint
* Ambient Weather
    * Add ability to reference last value of a sensor to perform offset calcs. 
        * offset = (temp - device.2.temp)
        * Should only be done within x time range.
* Dump points on dumper stop.
* Support Humidity Offset
* Correct name validation for log file due to number after name not before exentsion

## DONE
* Add slurper stop within EOF loop too.
* Fix false report of stop timeout hitting.
* Add flush timer
* Add ability to dynamically add tags based on tag values datapoint.
    * Channel = x => room:name
    * Temp = y => hot:yes
