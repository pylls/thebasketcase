package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
)

func addResult(base, result *metrics) {
	base.fn += result.fn
	base.fnp += result.fnp
	base.fpp += result.fpp
	base.tn += result.tn
	base.tp += result.tp
}

func getResult(output, trueclass int) (m metrics) {
	if output == trueclass {
		if trueclass < *sites {
			// found the right monitored site
			m.tp++
		} else {
			// correctly identified an unmonitored site
			m.tn++
		}
	} else {
		if output == *sites {
			// false negative: said unmonitored for a monitored
			m.fn++
		} else {
			if trueclass == *sites {
				// classifier said an unmonitored site was monitored
				m.fnp++
			} else {
				// classifier said the wrong monitored site
				m.fpp++
			}
		}
	}
	return
}

func instanceForTesting(i, fold int) bool {
	foldSize := *instances / *folds
	// the instances at [fold*foldSize,(fold+1)*foldSize) are for testing
	return i%*instances >= fold*foldSize && i%*instances < (fold+1)*foldSize
}

func getMaxInt(f []int) (val int, index int) {
	index = 0
	val = f[0]
	for i := 0; i < len(f); i++ {
		if f[i] > val {
			val = f[i]
			index = i
		}
	}
	return
}

// recall = TPR = TP / (TP + FN + FPP)
func recall(data []metrics) float64 {
	var p float64
	for i := 0; i < len(data); i++ {
		d := float64(data[i].tp) / float64(data[i].tp+data[i].fn+data[i].fpp)
		if !math.IsNaN(d) {
			p += d
		}
	}
	return p / float64(len(data))
}

// precision = TP / (TP + FPP + FNP)
func precision(data []metrics) float64 {
	var p float64
	for i := 0; i < len(data); i++ {
		d := float64(data[i].tp) / float64(data[i].tp+data[i].fpp+data[i].fnp)
		if !math.IsNaN(d) {
			p += d
		}
	}
	return p / float64(len(data))
}

// FPR = FP / non-monitored elements = (FPP + FNP) / (TN + FNP)
func fpr(data []metrics) float64 {
	var p float64
	for i := 0; i < len(data); i++ {
		d := float64(data[i].fpp+data[i].fnp) / float64(data[i].tn+data[i].fnp)
		if !math.IsNaN(d) {
			p += d
		}
	}
	return p / float64(len(data))
}

// f1 = 2 * [(precision*recall) / (precision + recall)]
func f1score(data []metrics) float64 {
	var p float64
	for i := 0; i < len(data); i++ {
		precision := float64(data[i].tp) / float64(data[i].tp+data[i].fpp+data[i].fnp)
		recall := float64(data[i].tp) / float64(data[i].tp+data[i].fn+data[i].fpp)
		if !math.IsNaN(precision) && !math.IsNaN(recall) {
			p += 2 * ((precision * recall) / (precision + recall))
		}
	}
	return p / float64(len(data))
}

// accuracy = (TP + TN) / (everything)
func accuracy(data []metrics) float64 {
	var p float64
	for i := 0; i < len(data); i++ {
		d := float64(data[i].tp+data[i].tn) /
			float64(data[i].fn+data[i].fnp+data[i].fpp+data[i].tn+data[i].tp)
		if !math.IsNaN(d) {
			p += d
		}
	}
	return p / float64(len(data))
}

func writeResults(results, name string) {
	err := ioutil.WriteFile(name, []byte(results), 0666)
	if err != nil {
		log.Fatalf("failed to write %s (%s)", name, err)
	}
}

func getMaxOccurance(values []int) (value, count int) {
	seen := make(map[int]int)
	for _, v := range values {
		seen[v]++
	}

	count = int(math.MinInt64)
	for v, c := range seen {
		if c > count {
			count = c
			value = v
		}
	}

	return
}

func writeTorpctCSV(metric func(data []metrics) float64,
	location string,
	results map[string][]metrics, // map["attack"] -> [folds]metrics
	attacks []string) {

	// headers
	output := "pct"
	for i := 0; i < len(attacks); i++ {
		output += "," + attacks[i]
	}
	output += "\n"

	// content
	for i := 0; i < len(attacks); i++ {
		output += fmt.Sprintf(",%.3f", metric(results[attacks[i]]))
	}
	output += "\n"

	writeResults(output, location)
}
