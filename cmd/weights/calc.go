package main

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/montanaflynn/stats"
)

func averageWeights(w map[string][][]float64) (o map[string][]float64) {
	o = make(map[string][]float64)

	for work, folds := range w {
		avg := make([]float64, len(folds[0]))

		for i := 0; i < len(folds); i++ { // sum up
			for j := 0; j < len(folds[i]); j++ {
				avg[j] += folds[i][j]
			}
		}
		for i := 0; i < len(avg); i++ { // divide by folds
			avg[i] = avg[i] / float64(len(folds))
		}

		o[work] = avg
	}
	return
}

func normalizeWeights(w map[string][]float64) (o map[string][]float64) {
	o = make(map[string][]float64)

	for work, weights := range w {
		c := make([]float64, len(weights)) // find smallest
		copy(c, weights)
		sort.Float64s(c)
		factor := 1.0 / c[0] // to normalize smallest to 1.0

		norm := make([]float64, len(weights))
		for i := 0; i < len(weights); i++ {
			norm[i] = weights[i] * factor // normalize
		}
		o[work] = norm
	}

	return
}

func printTop(x int, weights []float64) {
	c := make([]float64, len(weights))
	copy(c, weights)
	sort.Sort(sort.Reverse(sort.Float64Slice(c)))

	for i := 0; i < x; i++ {
		for j := 0; j < len(weights); j++ {
			if c[i] == weights[j] {
				log.Printf("\t%d. %s, rw %.2f",
					i+1, featName(j), weights[j])
			}
		}
	}
}

func printVolatile(x int, w map[string][]float64, order []string) {
	volatility := make([]float64, len(w[order[0]]))
	data := make([][]float64, len(w[order[0]]))

	for _, weights := range w {
		for i := 0; i < len(weights); i++ {
			data[i] = append(data[i], weights[i])
		}
	}

	for i := 0; i < len(data); i++ {
		d := stats.LoadRawData(data[i])
		g, err := d.GeometricMean()
		if err != nil {
			log.Fatalf("failed to calculate geometric mean (%s)", err)
		}
		volatility[i] = g
	}

	c := make([]float64, len(volatility))
	copy(c, volatility)
	sort.Sort(sort.Reverse(sort.Float64Slice(c)))

	for i := 0; i < x; i++ {
		for j := 0; j < len(volatility); j++ {
			if c[i] == volatility[j] {
				log.Printf("\t%d. %s, volatility %.1f",
					i+1, featName(j), volatility[j])
			}
		}
	}
}

func writeSortedFeatures(w map[string][]float64, order []string, fname string) {
	out := bytes.NewBufferString("")

	// work with features sorted in decreasing order
	for i := 0; i < len(order); i++ {
		str2buf(fmt.Sprintf("%s", order[i]), out)
		c := make([]float64, len(w[order[i]]))
		copy(c, w[order[i]])
		sort.Sort(sort.Reverse(sort.Float64Slice(c)))

		for j := 0; j < len(c); j++ {
			str2buf(fmt.Sprintf(",%s",
				strconv.FormatFloat(c[j], 'f', -1, 64)), out)
		}
		str2buf("\n", out)
	}

	// write to disk
	writeFile(out.String(), fname)
}

func writeFeatures(w map[string][]float64, order []string, fname string) {
	out := bytes.NewBufferString("")

	// work with features sorted in decreasing order
	for i := 0; i < len(order); i++ {
		str2buf(fmt.Sprintf("%s", order[i]), out)

		for j := 0; j < len(w[order[i]]); j++ {
			str2buf(fmt.Sprintf(",%s",
				strconv.FormatFloat(w[order[i]][j], 'f', -1, 64)), out)
		}
		str2buf("\n", out)
	}

	// write to disk
	writeFile(out.String(), fname)
}
