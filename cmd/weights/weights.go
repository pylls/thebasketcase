package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
)

var (
	top = flag.Int("top", 5, "number of top features to show")
)

func main() {
	flag.Parse()

	// read weights from file and then average over folds
	weights := normalizeWeights(averageWeights(readWeightsFile(flag.Arg(0))))

	// sort work strings for deterministic output
	var order []string
	for w := range weights {
		order = append(order, w)
	}
	sort.Strings(order)

	// print top features
	for i := 0; i < len(order); i++ {
		log.Printf("%s: top %d features", order[i], *top)
		printTop(*top, weights[order[i]])
		fmt.Println("")
	}

	// print most volatile features
	if len(weights) > 1 {
		log.Printf("top %d volatile features", *top)
		printVolatile(*top, weights, order)
	}

	// write features as-is and sorted to disk
	writeFeatures(weights, order, flag.Arg(0)+".norm")
	writeSortedFeatures(weights, order, flag.Arg(0)+".sorted")
}
