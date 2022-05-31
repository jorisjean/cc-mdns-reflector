# Status
[![Build/Test](https://github.com/jorisjean/cc-mdns-reflector/actions/workflows/go.yml/badge.svg)](https://github.com/jorisjean/cc-mdns-reflector/actions/workflows/go.yml)

# cc-mdns-reflector
This project is used to reflect mdns-packet from one client in one vlan to a configured chromecast in another vlan.
The mapping between client and chromecast is doen by MAC address inside a json file.

![cc-mdns-reflector-hl](https://user-images.githubusercontent.com/20523713/171103169-f62ea813-0e16-4da6-b2bb-09d13c24ec82.png)

This program is inspired from https://github.com/Gandem/bonjour-reflector

## Installation
On Linux
```
sudo apt-get update
sudo apt-get install go-dep golang-go libpcap0.8-dev -y
```
This project depends on libpcap to be able to capture and write packets

Checkout the project in your go src directory
```
git clone https://github.com/jorisjean/cc-mdns-reflector.git
cd cc-mdns-reflector

dep init
dep ensure
go build
go test
```

## Configuration
There are two input files for this program
config.ini which describes the static informations
```
[mdns-reflector]
chromecast_vlan=10                  # VLAN id of the Chromecasts
client_vlan=20                      # VLAN id of the Clients
mdns_reflector_int=eth0             # Interface to capture packets on
chromecast_spoof_ip=192.168.1.100   # IP address inside the Chromecast VLAN
```
chromecast_spoof_ip is needed because Chromecasts are only responding to mDNS packets originating from within their own subnet. 

states.json which describes the mappings between clients and chromecasts
```
[
    {
        "client_mac": "11:22:33:44:55:66",
        "cc_mac": "aa:bb:cc:dd:ee:ff",
    },
    {
        "client_mac": "99:88:77:66:55:44",
        "cc_mac": "ff:ee:dd:cc:bb:aa",
    }
]
```

## Usage
Run with:
`~/go/src/cc-mdns-reflector/cc-mdns-reflector -config=~/config.ini -states=~/states.json`

You can changes the mappings inside the states file and reload the cc-mdns-reflector with a `SIGHUP` to reload the states file without stoping the program.
