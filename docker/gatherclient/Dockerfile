# start from base
FROM debian:jessie
MAINTAINER Tobias Pulls <tobias.pulls@kau.se>

RUN apt-get update && apt-get install -y \
	xvfb \
	libpcap-dev \
	libasound2 \
	libdbus-glib-1-2 \
	libgtk2.0-0 \
	libxrender1 \
	libxt6 \
	xz-utils \
  	xauth \
	psmisc \
	file \
	--no-install-recommends

COPY dumb-init*.deb /
RUN dpkg -i dumb-init*.deb
RUN rm dumb*.deb && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ENV HOME /home/user
ENV LANG C.UTF-8

# create user (start-tor-browser.sh prevents us from running as root)
RUN useradd --create-home --home-dir $HOME user

COPY gatherclient $HOME/
COPY tor-browser_en-US $HOME/tor-browser_en-US

# make executable and give capability to dump network traffic
RUN chown -R user:user $HOME \
	&& chmod +x $HOME/gatherclient \
  && setcap 'CAP_NET_RAW+eip CAP_NET_ADMIN+eip' $HOME/gatherclient

WORKDIR $HOME
USER user
ENTRYPOINT ["dumb-init", "--"]
