# start from base
FROM debian:jessie
MAINTAINER Tobias Pulls <tobias.pulls@kau.se>

RUN gpg --keyserver keys.gnupg.net --recv 886DDD89
RUN gpg --export A3C4F0F979CAA22CDBA8F512EE8CBC9E886DDD89 | apt-key add -
RUN echo 'deb http://deb.torproject.org/torproject.org jessie main' >> /etc/apt/sources.list.d/tor.list

RUN apt-get update && apt-get install -y \
	tor \
	tor-geoipdb \
	--no-install-recommends

COPY dumb-init_1.0.1_amd64.deb /
RUN dpkg -i dumb-init_*.deb
RUN rm dumb-init_1.0.1_amd64.deb && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN usermod -u 9001 debian-tor
RUN groupmod -g 9001 debian-tor
RUN chown -R debian-tor:debian-tor /var/lib/tor /var/log/tor

ENV BASKET_DIR /opt/basket2/

RUN mkdir $BASKET_DIR
COPY basket2proxy $BASKET_DIR
COPY torrc /etc/tor/

EXPOSE 11111

ENTRYPOINT ["dumb-init", "--"]
