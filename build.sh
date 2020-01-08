#!/usr/bin/env bash

set -e

echo "cleaning..."
rm -fr dist/gps_dumper/gps_dumper 2>&1 || true
rm -fr dist/packet_dumper/packet_dumper 2>&1 || true
rm -fr dist/ssh_dumper/ssh_dumper 2>&1 || true
echo ""

echo "building..."
go build -v -o dist/gps_dumper/gps_dumper cmd/gps_dumper/main.go
go build -v -o dist/packet_dumper/packet_dumper cmd/packet_dumper/main.go
go build -v -o dist/ssh_dumper/ssh_dumper cmd/ssh_dumper/main.go
echo ""
