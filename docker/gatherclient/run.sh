#!/bin/sh
for ((n=0;n<$1;n++)) do
  docker run --privileged -d pulls/gatherclient ./gatherclient IP:55555
done
