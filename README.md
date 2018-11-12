# drive_test
A collection of tools for drive-testing networks

## What's it for?
Drive-testing networks

## How do I use it?
As an overview, each tool loops around and dumps to a file in [JSON lines](http://jsonlines.org/) format (to be post-processed later).

* use `ssh_dumper` to run SSH commands on a device
* use `gps_dumper` to pull locations from a local `gpsd` instance
* use `packet_dumper` to capture traffic on a local network interface
