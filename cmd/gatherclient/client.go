package main

import (
	"bytes"
	"flag"
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
	okTorData      = []string{"torrc-defaults",
		"geoip",
		"cached-descript",
		"cached-microdesc",
		"cached-certs"}
	pcapData       bytes.Buffer
	serverIP       = ""
	warmupSite     = "https://www.kau.se"
	redirectFormat = "<meta http-equiv=\"refresh\" content=\"%d; url=%s\" />"
	redirectDelay  = 10
	torrc          = ""
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

	work(client, identity, sampleChan)
}

func work(client model.GatherClient, identity string,
	sampleChan chan bool) (err error) {
	// we start with no completed work, then get to work
	report := &model.Report{
		WorkerID: identity,
		Browse: &model.Browse{
			ID: "",
		}}

	var lastWarmup time.Time
	for {
		// report and get work
		do, err := client.Work(context.Background(), report)
		if err != nil {
			log.Printf("failed to work (%s)", err)
			time.Sleep(10 * time.Second) // prevent spamming to connect
			continue
		}
		report.Browse = do // ugly, but needed
		torrc = report.Browse.Torrc
		report.Log = nil
		report.Pcap = nil

		if report.Browse.ID == "" {
			time.Sleep(time.Duration(report.Browse.Timeout) * time.Second)
			log.Printf("no work, sleeping for %d", report.Browse.Timeout)
			continue
		}

		// do we need to warmup?
		if time.Now().Sub(lastWarmup).Hours() >= 1 {
			log.Printf("warmup visit to %s", warmupSite)
			warmup(sampleChan)
			lastWarmup = time.Now()
		}

		log.Printf("starting work: %s", report.Browse.URL)
		pcap, torlog, err := browseTB(report.Browse.URL,
			int(report.Browse.Timeout), sampleChan)
		if err != nil {
			log.Printf("failed to browse (%s)", err)
		}
		report.Pcap = pcap
		if report.Browse.Log {
			report.Log = torlog
		}
	}
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
