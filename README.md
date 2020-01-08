# drive_test
A collection of tools for drive-testing networks

## What's it for?
Drive-testing networks

## How do I build it?

    ./build.sh
    
Observe as the executables are built to

    cmd/gps_dumper/gps_dumper
    cmd/packet_dumper/packet_dumper
    cmd/ssh_dumper/ssh_dumper
    
Optionally, if you need to cross-compile (e.g. for an ARM device):

    GOOS=linux GOARCH=arm ./build.sh

## How do I run it?

Each tool loops around and dumps to a file in [JSON lines](http://jsonlines.org/) (for later processing).

Some of the commands take a JSON config (default `config.json`).

The following sections some maximal examples of usage...

### `gps_dumper`

    # command line
    ./gps_dumper -host 127.0.0.1 -port 2947 -output-path gps_output.jsonl

### `packet_dumper`

    # contents of config.json; note "filter" is in tcpdump / pcap format
    {
      "filter": "udp and port 3784"
    }
    
    # command line
    sudo ./packet_dumper -interface eth0 -config-path config.json -output-path packet_output.jsonl

### `ssh_dumper`

    # contents of config.son
    {
      "setup_commands": [
        "date"
      ],
      "cycle_commands": [
        "uname -a",
        "uptime",
        "ifconfig -a"
      ]
    }

    # command line
    ./packet_dumper \
        -host localhost \
        -port 22 \
        -timeout 5 \
        -username user \
        -password pass \
        -period 8 \
        -config-path config.json \ 
        -output-path ssh_output.jsonl \
        -remove-command-echo true
        -remove-prompt-echo true
        -trim-output true
        -dumb-authentication false  
