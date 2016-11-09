package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/pylls/thebasketcase/model"
)

func store(in *model.Report) (err error) {
	if in.Browse.Log { // write log if we should
		err = ioutil.WriteFile(outputFileName(path.Clean(in.Browse.ID)+".torlog"),
			in.Log, 0666)
		if err != nil {
			return
		}
	}

	// write PCAP
	return ioutil.WriteFile(outputFileName(path.Clean(in.Browse.ID)+".pcap"),
		in.Pcap, 0666)
}

func outputFileName(id string) string {
	// structure is datadir/<mode>/<files>
	return path.Join(*datadir, activeMode, id)
}

func listFiles() []os.FileInfo {
	files, err := ioutil.ReadDir(path.Join(*datadir, activeMode))
	if err != nil {
		log.Fatal(err)
	}
	return files
}

func flipurl(u string, flip bool) string {
	if !flip {
		return u
	}

	if strings.Contains(u, "www.") {
		return strings.Replace(u, "www.", "", 1)
	}
	return strings.Replace(u, "://", "://www.", 1)
}
