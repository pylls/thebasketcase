# build basket2
- go get github.com/pylls/basket2proxy

# create and run docker container
cd into this dir
docker build -t pulls/basket2bridge .

docker run -p <IP>:11111:11111 -i -t pulls/basket2bridge bash
root@d07b96f4057c:/# service tor start
[ ok ] Starting tor daemon...done.
root@d07b96f4057c:/# cd /var/lib/tor/
root@85252b2e7a49:/var/lib/tor# cat fingerprint
Unnamed 5DD80B4AC2F718F1D8CACDAD1FD88644950A52B6
root@85252b2e7a49:/var/lib/tor# cat pt_state/basket2_bridgeline.txt
# basket2 torrc client bridge line
#
# This file is an automatically generated bridge line based on
# the current basket2proxy configuration.  EDITING IT WILL HAVE
# NO EFFECT.
#
# Before distributing this Bridge, edit the placeholder fields
# to contain the actual values:
#  <IP ADDRESS>  - The public IP address of your obfs4 bridge.
#  <PORT>        - The TCP/IP port of your obfs4 bridge.
#  <FINGERPRINT> - The bridge's fingerprint.

Bridge basket2 <IP ADDRESS>:<PORT> <FINGERPRINT> basket2params=0:0001:QiNZ5eqnrzPOXv4NyQ3Og5UntIpClPX6GC4c4Cq/I0Y

# example line for torrc by TB
Assuming the server with docker has IP 192.168.60.184, and the fingerprint
and basket2_bridgeline from above:

Bridge basket2 192.168.60.184:11111 5DD80B4AC2F718F1D8CACDAD1FD88644950A52B6 basket2params=0:0001:QiNZ5eqnrzPOXv4NyQ3Og5UntIpClPX6GC4c4Cq/I0Y
