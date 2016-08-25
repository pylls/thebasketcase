/*
FIXME
*/
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/pylls/thebasketcase/model"

	"google.golang.org/grpc"
)

const (
	port = ":55555"
)

type item struct {
	ID  string
	URL string
}

var (
	timeout = flag.Int("t", 15, "the timeout (seconds) for each sample load")
	samples = flag.Int("s", 1, "the number of samples to get for each page")
	datadir = flag.String("f", "data", "the folder to store results in")
	scheme  = flag.String("scheme", "http",
		"the scheme for pages where not specified")
	minDataLen = flag.Int("m", 25,
		"the minimum number of bytes to accept as data from a client")
	outputSuffix = flag.String("o", ".pcap", "the suffix for the output files")

	lock    sync.Mutex
	work    chan item
	workers map[string]string
	done    int
)

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano()) // good enough
	if len(flag.Args()) == 0 {
		log.Fatal("need to specify file with pages as argument")
	}

	// make sure we can write to datadir
	err := os.MkdirAll(*datadir, 0700)
	if err != nil {
		log.Fatalf("failed to create datadir (%s)", err)
	}

	// read pages and validate as URLs
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatalf("failed to read file with pages (%s)", err)
	}
	r := csv.NewReader(f)
	pages, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(pages); i++ {
		_, err = url.Parse(pages[i][1])
		if err != nil {
			log.Fatalf("failed to parse page as URL (%s)", err)
		}
	}
	workers = make(map[string]string)
	total := len(pages) * *samples

	var w []item // create work items
	for s := 0; s < *samples; s++ {
		for i := 0; i < len(pages); i++ {
			page, _ := url.Parse(pages[i][1])
			if page.Scheme == "" {
				page.Scheme = *scheme
			}
			id := pages[i][0] + "-" + strconv.Itoa(s)
			if _, err = os.Stat(outputFileName(id)); os.IsNotExist(err) {
				// only perform work if we have to
				w = append(w, item{
					ID:  id,
					URL: page.String(),
				})

			} else {
				done++
			}
		}
	}

	// the work channel holds all work items and has a bigger capacity to
	// prevent locks on restart of server and workers reporting in incomplete
	// work that's put back into the work channel
	work = make(chan item, total)
	for _, i := range rand.Perm(len(w)) { // random order
		work <- w[i]
	}

	log.Printf("collecting %d sample(s) of %d sites over %s",
		*samples, len(pages), *scheme)
	log.Printf("%d seconds timeout and storing results in \"%s\"",
		*timeout, *datadir)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("listening on %s", lis.Addr())

	// progress function
	go func() {
		for {
			lock.Lock()
			if done == total {
				fmt.Println("")
				log.Printf("finished")
				os.Exit(0)
			}
			fmt.Printf("\r %8d done (%3.1f%%), %8d left to distribute (%3d workers)",
				done, float64(done)/float64(total)*100, len(work), len(workers))
			lock.Unlock()
			time.Sleep(1 * time.Second)
		}
	}()

	s := grpc.NewServer()
	model.RegisterGatherServer(s, &server{})
	s.Serve(lis)
}
