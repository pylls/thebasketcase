# build basket2
- go get github.com/pylls/basket2proxy

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

# update urls (point all to localhost)
user_pref("extensions.update.background.url", "localhost");
user_pref("extensions.update.url", "localhost");
user_pref("xpinstall.signatures.devInfoURL", "localhost");
user_pref("browser.aboutHomeSnippets.updateUrl", "localhost");
user_pref("noscript.subscription.trustedURL", "localhost");
user_pref("noscript.subscription.untrustedURL", "localhost");
user_pref("extensions.torbutton.test_url", "localhost");
user_pref("extensions.torbutton.test_url_interactive", "localhost");
user_pref("extensions.torbutton.versioncheck_url", "localhost");

# update timers (move far into the future)
user_pref("app.update.lastUpdateTime.xpi-signature-verification", "2476106725");
user_pref("app.update.lastUpdateTime.search-engine-update-timer", "2476106725");
user_pref("app.update.lastUpdateTime.experiments-update-timer", "2476106725");
user_pref("app.update.lastUpdateTime.browser-cleanup-thumbnails", "2476106725");
user_pref("app.update.lastUpdateTime.background-update-timer", "2476106725");
user_pref("app.update.lastUpdateTime.addon-background-update-timer", "2476106725");
user_pref("idle.lastDailyNotification", "2476106725");
user_pref("app.update.timerMinimumDelay", 31536000);
user_pref("datareporting.healthreport.nextDataSubmissionTime", "2476194415800");
user_pref("datareporting.policy.firstRunTime", "2476194415800");
user_pref("noscript.subscription.lastCheck", "2476106725");
user_pref("extensions.torbutton.lastUpdateCheck", "2476106725");
user_pref("app.update.lastUpdateTime.blocklist-background-update-timer", "2476106725");

# disable updates
user_pref("extensions.blocklist.enabled", false);
user_pref("browser.safebrowsing.downloads.remote.enabled", false);
user_pref("app.update.auto", false);
user_pref("app.update.enabled", false);
user_pref("extensions.torlauncher.prompt_at_startup", false);
user_pref("extensions.torbutton.no_updates", true);
user_pref("extensions.torbutton.show_slider_notification", false);
user_pref("extensions.torbutton.updateNeeded", false);
user_pref("extensions.torbutton.versioncheck_enabled", false);

- in Browser/TorBrowser, put the modified tor from before
- copy basket2proxy to Browser/TorBrowser/Tor/PluggableTransports
- edit Browser/TorBrowser/Data/Tor/torrc, the basket2 settings below depend on
where you run your basket2 server (see docker/basket2 to run your own).

LogTimeGranularity 1
UseBridges 1
UseMicrodescriptors 1
Bridge basket2 192.168.60.184:11111 3A134BEE92330CBEE3DAED5AC289426E159A13B8 basket2params=0:0001:2Hki+jhzsNwuGnVl28bynFpkgHHdDzT6VkA78tTXdUs
ClientTransportPlugin basket2 exec ./TorBrowser/Tor/PluggableTransports/basket2proxy -enableLogging=true -logLevel DEBUG -paddingMethods Null

- modify  Browser/TorBrowser/Data/Browser/profile.default/extensions/https-everywhere-eff@eff.org/components/ssl-observatory.js and comment out line 405:
if (topic == "browser-delayed-startup-finished") {
  //this.testProxySettings();
}

- start TB once to verify the bridge settings
