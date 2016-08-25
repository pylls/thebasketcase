# thebasketcase
+ pcap dump all headers for traffic from/to basket2 box = most generic
- we want a docker container that starts TB with a given padding, dumps all
traffic in a pcap, and sends it away
- split gatherserver and basket2 server into two docker containers
- basket2 server should be isolated by IP from the gatherserver


# getting overhead
parse server-side log, make a standalone tool for that later

# server-side steps
docker build -t pulls/b2server .
docker run -p 127.0.0.1:11111:11111 -i -t pulls/b2server bash
service tor start
cat /var/lib/tor/fingerprint
cat /var/lib/tor/pt_state/basket2_bridgeline.txt

# server-side torrc
ORPort 9001
BridgeRelay 1
ExtORPort auto
PublishServerDescriptor 0
ServerTransportPlugin basket2 exec /opt/basket2/basket2proxy -enableLogging=true -logLevel DEBUG -paddingMethods TamarawBulk,Tamaraw,Obfs4PacketIAT,Obfs4BurstIAT,Obfs4Burst,Null
ServerTransportListenAddr basket2 0.0.0.0:11111
Log notice file /var/log/tor/log

# client-side torrc example
Bridge basket2 127.0.0.1:11111 7A27CC0853CDBBD03BEF0FA6EAD7E518D36A714B basket2params=0:0001:DUsJlAkMF+m7LMZ3EDL5/IgiSlwjjSZxvFn5AqEoqxQ
ClientTransportPlugin basket2 exec ./TorBrowser/Tor/PluggableTransports/basket2proxy -enableLogging=true -logLevel DEBUG -paddingMethods METHOD

DataDirectory /home/pulls/Downloads/tor-browser_en-US/Browser/TorBrowser/Data/Tor
GeoIPFile /home/pulls/Downloads/tor-browser_en-US/Browser/TorBrowser/Data/Tor/geoip
GeoIPv6File /home/pulls/Downloads/tor-browser_en-US/Browser/TorBrowser/Data/Tor/geoip6
HiddenServiceStatistics 0
UseBridges 1
