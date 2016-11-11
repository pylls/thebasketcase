/*
Package main implements a tool that extracts cells from pcap files in the
Wa-kNN format.  The result is written to ".cells" files used by the
extractfeatures tool.
*/
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var (
	workerFactor = flag.Int("f", 2,
		"the factor to multiply NumCPU with for creating workers")
	output = flag.String("o", "", "folder to store results in")
	bridge = flag.String("bridge", "192.168.60.184",
		"the bridge that our Tor client connected to")
)

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 {
		log.Fatal("need to specify pcap dir")
	}
	if *output == "" {
		*output = flag.Arg(0)
	}

	files, err := ioutil.ReadDir(flag.Arg(0))
	if err != nil {
		log.Fatalf("failed to read pcap dir (%s)", err)
	}

	for i := 0; i < len(files); i++ {
		if files[i].IsDir() {

		}
	}

	work := make(chan string)
	wg := new(sync.WaitGroup)
	wg.Add(runtime.NumCPU() * *workerFactor)
	for i := 0; i < runtime.NumCPU()**workerFactor; i++ {
		go doWork(work, wg)
	}

	log.Printf("starting to extract (%d workers)...",
		runtime.NumCPU()**workerFactor)
	extracted := 0
	for i := 0; i < len(files); i++ {
		if files[i].IsDir() { // support one level of sub-folders
			subfiles, err := ioutil.ReadDir(path.Join(flag.Arg(0), files[i].Name()))
			if err != nil {
				log.Fatalf("failed to read pcap subdir (%s)", err)
			}
			err = os.MkdirAll(path.Join(*output, files[i].Name()), 0700)
			if err != nil {
				log.Fatalf("failed to create subdir to store output in (%s)", err)
			}
			for _, f := range subfiles {
				if !f.IsDir() && strings.HasSuffix(f.Name(), ".pcap") {
					fmt.Printf("\rextracted %d", extracted)
					work <- path.Join(files[i].Name(), f.Name())
					extracted++
				}
			}
		}

		if !files[i].IsDir() && strings.HasSuffix(files[i].Name(), ".pcap") {
			fmt.Printf("\rextracted %d", extracted)
			work <- files[i].Name()
			extracted++
		}
	}
	close(work)
	wg.Wait()
	fmt.Printf("\rextracted %d\n", extracted)
	log.Println("done")
}

func doWork(input chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for file := range input {
		extract(file)
	}
}

func extract(file string) {
	cells, err := extractCells(path.Join(flag.Arg(0), file))
	if err != nil {
		log.Fatalf("failed to extract cells (%s)", err)
	}
	f, err := os.Create(path.Join(*output, file[:len(file)-5]+".cells"))
	if err != nil {
		log.Fatalf("failed to create file to store result in (%s)", err)
	}
	_, err = f.WriteString(cells)
	if err != nil {
		log.Fatalf("failed to write result to file (%s)", err)
	}
	err = f.Close()
	if err != nil {
		log.Fatalf("failed to close file (%s)", err)
	}
}

func extractCells(pcapfile string) (cells string, err error) {
	handle, err := pcap.OpenOffline(pcapfile)
	if err != nil {
		return "", fmt.Errorf("failed to open pcap file %s (%s)", pcapfile, err)
	}
	source := gopacket.NewPacketSource(handle, layers.LinkTypeEthernet)
	seenIPs := make(map[string]bool)
	var first *time.Time
	for packet := range source.Packets() {
		if packet.TransportLayer().LayerType() == layers.LayerTypeTCP {
			srcIP := packet.NetworkLayer().NetworkFlow().Src().String()
			seenIPs[srcIP] = true // register all IPs

			if first == nil { // time is relative to the first packet
				first = new(time.Time)
				*first = packet.Metadata().Timestamp
			}

			if packet.Metadata().Length >= 100 {
				// ACKs are 66 bytes, assume most under 100 bytes are noise
				if srcIP == *bridge { // did the bridge send this to the client?
					cells += fmt.Sprintf("%.3f\t-1\n", // incoming
						packet.Metadata().Timestamp.Sub(*first).Seconds())
				} else {
					cells += fmt.Sprintf("%.3f\t1\n", // outgoing
						packet.Metadata().Timestamp.Sub(*first).Seconds())
				}
			}
		}
	}
	handle.Close()

	if len(seenIPs) != 2 {
		return "", fmt.Errorf("pcap has %d IPs, expected 2 (%v)",
			len(seenIPs), seenIPs)
	}

	return
}
