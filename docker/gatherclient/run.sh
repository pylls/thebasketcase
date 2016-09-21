#!/bin/sh
for ((n=0;n<$1;n++)) do
  docker run --privileged -d pulls/gatherclient ./gatherclient 130.243.26.103:55555
done
