package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

func driver() {
	// randomize order once to use same order for all modes
	unmonitoredPerm := rand.Perm(len(sites))

	m := strings.Split(*methods, ",")
	for i := 0; i < len(m); i++ {
		activeMethod = m[i] // set new method and create torrc config
		torrc = strings.Replace(string(torrcTemplate),
			methodString, activeMethod, -1)
		err := os.MkdirAll(path.Join(*datadir, activeMethod), 0700)
		if err != nil {
			log.Fatalf("failed to create datadir for method (%s)", err)
		}

		log.Printf("collecting for method %s (%d/%d)", activeMethod, i+1, len(m))
		log.Println(" -monitored:")
		collectMonitored()

		log.Println(" -unmonitored:")
		collectUnmonitored(unmonitoredPerm)
	}

	log.Printf("finished")
	os.Exit(0)
}

func collectMonitored() {
	lock.Lock()
	done = make(map[string]bool)
	batchID = time.Now().String()
	work = make(chan item, *monitored**samples)
	flip := false // don't flip URLs on first try

	log.Println("\tchecking disk") // already on disk -> done
	for i := 0; i < *monitored; i++ {
		for sample := 0; sample < *samples; sample++ {
			id := sites[i][0] + "-" + strconv.Itoa(sample)
			if _, err := os.Stat(outputFileName(id + ".pcap")); !os.IsNotExist(err) {
				done[id] = true
			}
		}
	}
	lock.Unlock()
	log.Printf("\tfound %d samples on disk", len(done))
	log.Println("\tstarting to collect:")

	for { // loop until we got all samples
		lock.Lock()
		var w []item
		for i := 0; i < *monitored; i++ {
			for sample := 0; sample < *samples; sample++ {
				id := sites[i][0] + "-" + strconv.Itoa(sample)
				if _, exists := done[id]; !exists { // add if not done
					page, _ := url.Parse(sites[i][1]) // make proper URL
					if page.Scheme == "" {
						page.Scheme = *scheme
					}
					w = append(w, item{
						ID:  id,
						URL: flipurl(page.String(), flip),
					})
				}
			}
		}
		flip = true
		for _, i := range rand.Perm(len(w)) { // random order
			work <- w[i]
		}
		lock.Unlock()

		for { // OK, setup done: time to wait for collection
			lock.Lock()
			fmt.Printf("\r\t\t\t%d/%d done (%3.1f%%), %d workers",
				len(done), *monitored**samples,
				float64(len(done))/float64(*monitored**samples)*100, len(workers))
			lock.Unlock()

			if len(done) >= *monitored**samples { // are we done?
				fmt.Println("")
				return
			}
			if len(work) == 0 {
				break
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func collectUnmonitored(perm []int) {
	lock.Lock()
	done = make(map[string]bool)
	batchID = time.Now().String()
	work = make(chan item, len(sites))

	// check how many done on disk
	log.Println("\tchecking disk")
	files := listFiles()
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), "pcap") &&
			strings.Contains(f.Name(), "-") {
			// attempt to read site
			s, err := strconv.Atoi(f.Name()[:strings.Index(f.Name(), "-")])
			if err != nil {
				continue // not expected format
			}
			if s > *monitored { // mark unmonitored as done
				done[fmt.Sprintf("%d-0", s)] = true
			}
		}
	}
	log.Printf("\tfound %d samples on disk", len(done))

	// fill channel with all potential unmonitored sites
	for i := 0; i < len(perm); i++ {
		if perm[i] <= *monitored {
			continue // skip monitored sites
		}

		id := fmt.Sprintf("%d-0", perm[i])
		_, exists := done[id]
		if !exists {
			page, _ := url.Parse(sites[perm[i]][1]) // make proper URL
			if page.Scheme == "" {
				page.Scheme = *scheme
			}
			work <- item{
				ID:  id,
				URL: page.String(),
			}
		}
	}
	lock.Unlock()

	log.Println("\tstarting to collect:")
	for { // OK, setup done: time to wait for collection
		lock.Lock()
		fmt.Printf("\r\t\t\t%d/%d done (%3.1f%%), %d workers",
			len(done), *unmonitored,
			float64(len(done))/float64(*unmonitored)*100, len(workers))
		lock.Unlock()

		if len(work) == 0 {
			log.Fatalf("out of work: this should never happen")
		}
		if len(done) >= *unmonitored { // are we done?
			fmt.Println("")
			return
		}
		time.Sleep(1 * time.Second)
	}
}
