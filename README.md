# drive_test
A collection of tools for drive-testing networks

## What's it for?
Drive-testing networks

## How do I use it?
As an overview, each tool loops around and dumps to a file in [JSON lines](http://jsonlines.org/) format (to be post-processed later), some of the commands take JSON configurations.

* use `ssh_dumper` to run SSH commands on a device
    * be sure to edit `config.json` first
    * `ssh_dumper -username some_user -password some_password -host some_host.org`
* use `gps_dumper` to pull locations from a local `gpsd` instance
    * `gps_dumper`
* use `packet_dumper` to capture traffic on a local network interface
    * be sure to edit `config.json` first
    * `sudo packet_dumper -interface eth0`
