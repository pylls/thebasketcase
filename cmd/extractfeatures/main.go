/*
Package main implements feature extraction from packet traces (".cells" files).
The output format is a fixed and optimized (to Tor) version of the one used by
Wa-kNN (https://crysp.uwaterloo.ca/software/webfingerprint/).
*/
package main

import (
	"bufio"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var (
	suffix = flag.String("suffix", ".feat",
		"the suffix for the resulting files with parsed features")
	intype = flag.String("intype", ".cells",
		"the suffix of the type of file to parse")
	output = flag.String("o", "", "folder to store results in")
)

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 {
		log.Fatal("need to specify data dir")
	}
	if *output == "" {
		*output = flag.Arg(0)
	}

	// workers
	wg := new(sync.WaitGroup)
	work := make(chan string)
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filename := range work {
				parse(filename)
			}
		}()
	}

	log.Printf("getting list of files in %s", flag.Arg(0))
	files, er := ioutil.ReadDir(flag.Arg(0))
	if er != nil {
		log.Fatalf("failed to read data dir (%s)", er)
	}
	log.Printf("OK, starting to parse...")

	samples := 0
	for i := 0; i < len(files); i++ {
		if files[i].IsDir() { // support one level of sub-folders
			subfiles, err := ioutil.ReadDir(path.Join(flag.Arg(0), files[i].Name()))
			if err != nil {
				log.Fatalf("failed to read subdir (%s)", err)
			}
			err = os.MkdirAll(path.Join(*output, files[i].Name()), 0700)
			if err != nil {
				log.Fatalf("failed to create subdir to store output in (%s)", err)
			}
			for _, f := range subfiles {
				if !f.IsDir() && strings.HasSuffix(f.Name(), *intype) {
					samples++
					work <- path.Join(files[i].Name(), f.Name())
				}
			}
		}
		if !files[i].IsDir() && strings.HasSuffix(files[i].Name(), *intype) {
			samples++
			work <- files[i].Name()
		}
	}

	close(work)
	wg.Wait()

	log.Printf("done parsing %d samples in folder \"%s\" of type \"%s\"",
		samples, flag.Arg(0), *intype)
	log.Printf("results stored in folder \"%s\" with suffix \"%s\"",
		*output, *suffix)
}

func parse(filename string) {
	file, err := os.Open(path.Join(flag.Arg(0), filename))
	if err != nil {
		log.Fatalf("failed to read file %s, got error %s", filename, err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	var times []float64
	var sizes []int
	for scanner.Scan() {
		items := strings.Split(scanner.Text(), "\t")
		if len(items) != 2 {
			log.Fatalf("expected 2 items in line for filename %s, got %d",
				filename, len(items))
		}

		t, er := strconv.ParseFloat(items[0], 64)
		if er != nil {
			log.Fatalf("failed to parse time for filename %s, %s", filename, err)
		}
		times = append(times, t)

		s, er := strconv.ParseInt(items[1], 10, 64)
		if er != nil {
			log.Fatalf("failed to parse size for filename %s, %s", filename, err)
		}
		sizes = append(sizes, int(s))
	}

	features, err := extract(times, sizes)
	if err != nil {
		log.Fatalf("failed to extract features for filename %s, %s", filename, err)
	}
	err = ioutil.WriteFile(path.Join(*output, strings.Replace(filename, *intype, *suffix, 1)),
		[]byte(features+FeatureDelimiter), 0666)
	if err != nil {
		log.Fatalf("failed to write features file for filename %s, %s",
			filename, err)
	}
}
