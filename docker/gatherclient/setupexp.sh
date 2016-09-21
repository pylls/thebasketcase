#!/bin/sh
wget http://www.cs.kau.se/pulls/basketexp.tar.gz
tar -zxf basketexp*.tar.gz
chmod +x run.sh clean.sh
ls
docker rmi pulls/gatherclient
docker build -t pulls/gatherclient .
