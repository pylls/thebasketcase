# build basket2
- go get git.schwanenlied.me/yawning/basket2.git/basket2proxy
- if you plan to do large-scale collection, consider modifying
handshake/replayfilter.go to increase the size of the replay filter. 

# build a custom tor:
- use a debian:jessie system (TODO: make easy docker for build)
- git clone https://git.torproject.org/tor.git
- cd and ./configure

- in src/or/addressmap.c line ~665 in function client_dns_set_addressmap_impl:
  log_notice(LD_GENERAL, "DNSRESOLVED %s ip %s ttl %d", address, name, ttl);
- in src/or/relay.c line ~567 in function relay_send_command_from_edge_:
  log_notice(LD_GENERAL, "OUTGOING CIRC %u STREAM %d COMMAND %s(%d) length %zu",
  circ->n_circ_id, stream_id,
  relay_command_to_string(relay_command), relay_command, payload_len);
- in src/or/relay.c line ~1460 in function connection_edge_process_relay_cell:
  log_notice(LD_GENERAL, "INCOMING CIRC %u STREAM %d COMMAND %s(%d) length %d",
  circ->n_circ_id, rh.stream_id,
  relay_command_to_string(rh.command), rh.command, rh.length);

- make
- new tor is at src/or/tor

# download a fresh Tor Browser:
- https://www.torproject.org/download/download-easy.html.en
- edit Browser/TorBrowser/Data/Browser/profile.default/preferences/extension-overrides.js
user_pref("app.update.enabled", false);
user_pref("extensions.torlauncher.prompt_at_startup", false);
user_pref("datareporting.healthreport.nextDataSubmissionTime", "1859373924100");
user_pref("datareporting.policy.firstRunTime", "1859287524100");
user_pref("extensions.torbutton.lastUpdateCheck", "1859287542.7");
user_pref("extensions.torbutton.show_slider_notification", false);
user_pref("extensions.torbutton.updateNeeded", false);
user_pref("extensions.torbutton.versioncheck_url", "");
user_pref("extensions.torbutton.versioncheck_enabled", false);

- in Browser/TorBrowser, put the modified tor from before
- copy basket2proxy to Browser/TorBrowser/Tor/PluggableTransports
- edit Browser/TorBrowser/Data/Tor/torrc, the basket2 settings below depend on
where you run your basket2 server (see docker/basket2 to run your own).
LogTimeGranularity 1
UseBridges 1
Bridge basket2 192.168.60.184:11111 5DD80B4AC2F718F1D8CACDAD1FD88644950A52B6 basket2params=0:0001:QiNZ5eqnrzPOXv4NyQ3Og5UntIpClPX6GC4c4Cq/I0Y
ClientTransportPlugin basket2 exec ./TorBrowser/Tor/PluggableTransports/basket2proxy -enableLogging=true -logLevel DEBUG -paddingMethods Null
- start TB once to verify the bridge settings
- start TB just before you start collection with docker containers to cache
a fresh copy of the consensus
