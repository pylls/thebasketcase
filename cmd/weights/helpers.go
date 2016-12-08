package main

import (
	"bytes"
	"encoding/csv"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

// w[work] -> [fold][feature]weight, where fold and feature are ints
func readWeightsFile(filename string) (w map[string][][]float64) {
	w = make(map[string][][]float64)
	f, err := os.Open(filename)
	if err != nil {
		log.Fatalf("failed to open weights file (%s)", err)
	}
	r := csv.NewReader(f)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("failed to read record from weights file (%s)", err)
		}
		if len(record) <= 2 {
			log.Fatalf("record length too small (%d)", len(record))
		}
		if record[0] == "work" && record[1] == "fold" {
			continue // headers
		}

		work, exists := w[record[0]] // work already there?
		if !exists {
			work = make([][]float64, 0)
		}
		var fold []float64
		for i := 2; i < len(record); i++ {
			v, err := strconv.ParseFloat(record[i], 64)
			if err != nil {
				log.Fatalf("failed to parse weight (%s)", err)
			}
			fold = append(fold, v)
		}
		w[record[0]] = append(work, fold) // save (allocation can create new work)
	}

	return
}

func str2buf(s string, buf *bytes.Buffer) {
	_, err := buf.WriteString(s)
	if err != nil {
		log.Fatalf("failed to write string (%s)", err)
	}
}

func writeFile(results, name string) {
	err := ioutil.WriteFile(name, []byte(results), 0666)
	if err != nil {
		log.Fatalf("failed to write %s (%s)", name, err)
	}
}
