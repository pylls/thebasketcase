#!/bin/sh
rm -rf Dockerfile README.md dumb-init_1.0.1_amd64.deb expNull.tar.xz* gatherclient run.sh tor-browser_en-US/
wget pulls/expNull.tar.xz
tar -xf exp*.tar.xz
chmod +x run.sh clean.sh
ls
docker rmi pulls/gatherclient
docker build -t pulls/gatherclient .
