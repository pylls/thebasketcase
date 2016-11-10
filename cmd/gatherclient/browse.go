package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

func browseTB(url string, seconds int,
	sampleChan chan bool) (pcap, torlog []byte, err error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	for i := 0; i < *attempts; i++ { // because our xvfb+timeout+TB fails at times
		err = nil
		time.Sleep(1 * time.Second)

		err = setup()
		if err != nil {
			log.Printf("%s", err)
			continue
		}

		// create redirect file
		redirectFile := path.Join(browser, "Browser", "redirect.html")
		err = ioutil.WriteFile(redirectFile,
			[]byte(fmt.Sprintf(redirectFormat, redirectDelay, url)), 0666)
		if err != nil {
			log.Printf("%s", err)
			continue
		}

		tb := exec.Command("xvfb-run", "-s", *display, // in a xvfb display,
			"timeout", "-s", "9", // kill (no need to play nice) after
			strconv.Itoa(seconds),                              // seconds
			path.Join(browser, "Browser", "start-tor-browser"), // Tor Browser
			"--debug", redirectFile) // that visits a specific URL through redirect
		stdout.Reset()
		stderr.Reset()
		tb.Stdout = &stdout
		tb.Stderr = &stderr

		go func() {
			// overwrite the pcap after the redirect duration, hoping that most of
			// the "noise" relating to circuit construction and whatnot is done by
			// then (note that running TB takes non-zero time, so we have margin)
			time.Sleep(time.Duration(redirectDelay) * time.Second)
			sampleChan <- true // overwrites pcap
		}()

		// run TB (blocking), filling stdout and stderr
		tb.Run()

		// wait for killing TB and any lagging data
		time.Sleep(2 * time.Second)

		// check stdout to see if we got any useful data or not
		if !gotData(stdout) {
			err = fmt.Errorf("didn't get data browsing, stdout (%s), stderr (%s)",
				stdout.String(), stderr.String())
			continue
		}

		// we got data
		return pcapData.Bytes(), stdout.Bytes(), nil
	}

	return nil, nil, fmt.Errorf("failed to browse after %d attempts (%s) (%s)",
		*attempts, stdout.Bytes(), stderr.Bytes())
}

func warmup(sampleChan chan bool) {
	_, _, err := browseTB(warmupSite, 60, sampleChan)
	if err != nil {
		log.Fatalf("failed to warmup browse (%s)", err)
	}
}

func setup() (err error) {
	// get a fresh copy of the temporary data browser dir
	err = os.RemoveAll(path.Join(browser, dataBrowserDir))
	if err != nil {
		return fmt.Errorf("failed to remove Browser directory at %s (%s)",
			path.Join(browser, dataBrowserDir), err)
	}
	cp := exec.Command("cp", "-rfT", path.Join(*origBrowser, dataBrowserDir),
		path.Join(browser, dataBrowserDir))
	err = cp.Run()
	if err != nil {
		return fmt.Errorf("failed to copy Browser directory to %s (%s)",
			path.Join(browser, dataBrowserDir), err)
	}

	// delete files for Tor in the data dir we do not want to keep
	files, err := ioutil.ReadDir(path.Join(browser, dataTorDir))
	if err != nil {
		return fmt.Errorf("failed to read data dir (%s)", err)
	}
	for _, f := range files {
		if !f.IsDir() {
			ok := false
			for _, name := range okTorData {
				if strings.Contains(f.Name(), name) {
					ok = true
					break
				}
			}
			if !ok {
				os.RemoveAll(path.Join(browser, dataTorDir, f.Name()))
				if err != nil {
					return fmt.Errorf("failed to remove Tor data file %s (%s)",
						path.Join(dataTorDir, f.Name()), err)
				}
			}
		}
	}

	// write torrc file
	return ioutil.WriteFile(path.Join(browser, dataTorDir, "torrc"),
		[]byte(torrc), 0600)
}

func gotData(in bytes.Buffer) bool {
	// we got data if:
	bootstrapped := false // we bootstrapped,
	begin := false        // began sending outgoing data, and
	domain := false       // successfully resolved at least one domain

	scanner := bufio.NewScanner(bytes.NewReader(in.Bytes()))
	for scanner.Scan() {
		tokens := strings.Split(scanner.Text(), " ")
		if len(tokens) > 5 {
			switch tokens[4] {
			case "OUTGOING":
				if len(tokens) > 10 && tokens[10] == "BEGIN(1)" {
					begin = true
				}
			case "DNSRESOLVED":
				domain = true
			case "Bootstrapped":
				if len(tokens) > 6 && tokens[5] == "100%:" {
					bootstrapped = true
				}
			}
		}
		if begin && domain && bootstrapped {
			return true
		}
	}
	return false
}
