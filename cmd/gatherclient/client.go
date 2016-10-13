package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"

	"github.com/pylls/thebasketcase/model"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	attempts = flag.Int("a", 5,
		"the number of attempts per browse to launch TB")
	origBrowser = flag.String("b", "tor-browser_en-US",
		"the location of the TB folder")
	display = flag.String("display", "-screen 0 1024x768x24",
		"the xvfb display to use")
	nic     = flag.String("nic", "eth0", "the NIC to listen on for traffic")
	snaplen = flag.Int("snaplen", 58, "the snaplen to capture")

	tmpDir         = path.Join(os.TempDir(), "hotexp")
	browser        = path.Join(tmpDir, "browser")
	dataBrowserDir = "Browser/TorBrowser/Data/Browser"
	dataTorDir     = "Browser/TorBrowser/Data/Tor"
	okTorData      = []string{"torrc",
		"geoip",
		"cached-descript",
		"cached-microdesc",
		"cached-certs"}
	pcapData       bytes.Buffer
	serverIP       = ""
	warmupSite     = "https://www.kau.se"
	redirectFormat = "<meta http-equiv=\"refresh\" content=\"%d; url=%s\" />"
	redirectDelay  = 10
)

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 {
		log.Fatal("need to specify server address")
	}
	serverIP = strings.Split(flag.Arg(0), ":")[0]

	os.Remove(tmpDir)
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		return
	}
	defer os.Remove(tmpDir)

	// copy entire browser to a temporary location
	err = os.MkdirAll(browser, 0755)
	if err != nil {
		return
	}
	cp := exec.Command("cp", "-rfT", *origBrowser, browser)
	err = cp.Run()
	if err != nil {
		log.Fatalf("failed to copy to %s (%s)", browser, err)
	}

	conn, err := grpc.Dial(flag.Arg(0), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := model.NewGatherClient(conn)

	// base identity reported to server on IPs for easy remote access
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatalf("failed to get network interfaces (%s)", err)
	}
	identity := strconv.Itoa(int(time.Now().UnixNano())) + "\t"
	for i := 0; i < len(addrs); i++ {
		identity += addrs[i].String() + " "
	}

	// start traffic capture
	handler, err := pcap.OpenLive(*nic, int32(*snaplen), false, pcap.BlockForever)
	if err != nil {
		log.Fatalf("failed to open capture (%s)", err)
	}
	defer handler.Close()
	source := gopacket.NewPacketSource(handler, layers.LinkTypeEthernet)
	sampleChan := make(chan bool)
	defer close(sampleChan)
	log.Println("collecting network traffic")
	go collectNetwork(source.Packets(), sampleChan)

	// we start with no completed work, then get to work
	work := new(model.Req)
	work.WorkerID = identity
	work.Browse = &model.Browse{
		ID: "",
	}

	var lastWarmup time.Time
	for {
		if time.Now().Sub(lastWarmup).Hours() >= 1 {
			log.Printf("warmup visit to %s", warmupSite)
			warmup(sampleChan)
			lastWarmup = time.Now()
		}

		// report and get work
		browse, err := client.Work(context.Background(), work)
		if err != nil {
			log.Printf("failed to work (%s)", err)
			time.Sleep(10 * time.Second) // prevent spamming to connect
			continue
		}
		work.Browse = browse
		if browse.ID == "" {
			time.Sleep(time.Duration(browse.Timeout) * time.Second)
			log.Printf("no work, sleeping for %d", browse.Timeout)
			continue
		}

		log.Printf("starting work: %s", browse.URL)
		data, err := browseTB(browse.URL, int(browse.Timeout), sampleChan)
		if err != nil {
			log.Printf("failed to browse (%s)", err)
			data = []byte("none")
		}
		browse.Data = data
	}
}

func browseTB(url string, seconds int,
	sampleChan chan bool) (data []byte, err error) {
	for i := 0; i < *attempts; i++ { // because our xvfb+timeout+TB fails at times
		err = nil
		time.Sleep(1 * time.Second)

		err = clean()
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
		var stdout bytes.Buffer
		var stderr bytes.Buffer
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
			err = fmt.Errorf("didn't get enough data while attempting to browse, stdout (%s), stderr (%s)",
				stdout.String(), stderr.String())
			continue
		}

		// we got data, return the raw _pcap_ data, not torlog in stdout, since
		// we're interested in how it looks on the wire with PT defenses, not
		// without defenses in stdout
		return pcapData.Bytes(), nil
	}
	return nil, fmt.Errorf("failed to browse")
}

func warmup(sampleChan chan bool) {
	_, err := browseTB(warmupSite, 60, sampleChan)
	if err != nil {
		log.Fatalf("failed to warmup browse (%s)", err)
	}
}

func clean() (err error) {
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
	return
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

func collectNetwork(pChan chan gopacket.Packet, sampleChan chan bool) {
	var w *pcapgo.Writer
	var err error
	for {
		select {
		case _ = <-sampleChan:
			// truncate pcap-data
			pcapData.Reset()
			w = pcapgo.NewWriter(&pcapData)
			// new pcap, must write headers with snaplen
			err = w.WriteFileHeader(uint32(*snaplen), layers.LinkTypeEthernet)
			if err != nil {
				log.Fatalf("failed to write pcap header (%s)", err)
			}
		case packet := <-pChan:
			// parse packet
			if w != nil {
				var src, dst string
				if packet.NetworkLayer() != nil {
					src = packet.NetworkLayer().NetworkFlow().Src().String()
					dst = packet.NetworkLayer().NetworkFlow().Dst().String()
				}
				// only capture TCP packets, that if sent by IP, are not to or from
				// the gatherserver (serverIP)
				if packet.TransportLayer() != nil &&
					packet.TransportLayer().LayerType() == layers.LayerTypeTCP &&
					!strings.Contains(src, serverIP) && !strings.Contains(dst, serverIP) {
					err := w.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
					if err != nil {
						log.Fatalf("failed to write packet to pcap (%s)", err)
					}
				}
			}
		}
	}
}
