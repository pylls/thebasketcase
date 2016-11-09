/*
FIXME
*/
package main

import (
	"encoding/csv"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pylls/thebasketcase/model"

	"google.golang.org/grpc"
)

const (
	port       = ":55555"
	modeString = "{{.Mode}}"
)

type item struct {
	ID  string
	URL string
}

var (
	timeout = flag.Int("timeout", 60,
		"the timeout (s) for collecting a sample")
	monitored = flag.Int("monitored", 10, "the number of monitored sites")
	samples   = flag.Int("samples", 1,
		"the number of samples to get for each site")
	unmonitored = flag.Int("unmonitored", 10, "the number of unmonitored sites")
	datadir     = flag.String("f", "data", "the folder to store results in")
	scheme      = flag.String("scheme", "http",
		"the scheme for sites where not specified")
	minDataLen = flag.Int("min", 100,
		"the minimum number of bytes to accept as pcap data from a client")
	getTorLog = flag.Bool("getlog", false,
		"also get the output of the tor stdout log")
	modes = flag.String("modes",
		"TamarawBulk,Tamaraw,Obfs4PacketIAT,Obfs4BurstIAT,Obfs4Burst,Null",
		"the basket2 modes")
	torrcLocation = flag.String("torrc", "torrc",
		"the template for the Tor configuration file to be used by workers")

	lock    sync.Mutex
	work    chan item
	workers map[string]string
	sites   [][]string

	done                       map[string]bool
	batchID, activeMode, torrc string
	torrcTemplate              []byte
)

func main() {
	workers = make(map[string]string)

	flag.Parse()
	rand.Seed(time.Now().UnixNano()) // good enough
	if len(flag.Args()) == 0 {
		log.Fatal("need to specify file with sites as argument")
	}

	// make sure we can write to datadir
	err := os.MkdirAll(*datadir, 0700)
	if err != nil {
		log.Fatalf("failed to create datadir (%s)", err)
	}

	// read sites file, validating URLs
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatalf("failed to read file with sites (%s)", err)
	}
	r := csv.NewReader(f)
	sites, err = r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	// validate sites
	for i := 0; i < len(sites); i++ {
		_, err = url.Parse(sites[i][1])
		if err != nil {
			log.Fatalf("failed to parse site as URL (%s)", err)
		}
	}

	// read torrc file
	torrcTemplate, err = ioutil.ReadFile(*torrcLocation)
	if err != nil {
		log.Fatalf("failed to read torrc file (%s)", err)
	}
	if !strings.Contains(string(torrcTemplate), modeString) {
		log.Fatalf("torrc fails missing mode string %s", modeString)
	}

	// we can't validate modes further
	if len(*modes) == 0 {
		log.Fatalf("missing modes")
	}

	log.Printf("collecting %dx%d+%d dataset for modes %v",
		*monitored, *samples, *unmonitored, *modes)
	log.Printf("storing data in \"%s\", with %ds timeout, over %s",
		*datadir, *timeout, *scheme)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("listening on %s", lis.Addr())

	go driver() // do actual work

	s := grpc.NewServer()
	model.RegisterGatherServer(s, &server{})
	s.Serve(lis)
}
