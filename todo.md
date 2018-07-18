

# TODO
* Add some sort of cache for failures or wait until they start going through. All others should backup on the channel.
* Add ability to dynamically add tags based on values of the DataPoint
    * Channel = x => room:name
    * Temp = y => hot:yes
* Ambient Weather
    * Add ability to reference last value of a sensor to perform offset calcs. 
        * offset = (temp - device.2.temp)
        * Should only be done within x time range.
* Add flush timer
* Fix up flush datapoint count and make sure example shows true default.
* Add slurper stop withink EOF loop too.

## DONE
* Fix false report of stop timeout hitting.

