package main

import (
	"math"
	"strconv"
)

// FeatureDelimiter is the delimiter in the output between features
const FeatureDelimiter = " "

func extract(times []float64, sizes []int) (features string, err error) {
	// transmission size features
	count := 0
	for _, s := range sizes {
		if s > 0 {
			count++
		}
	}
	features = strconv.Itoa(len(times))
	features += FeatureDelimiter + strconv.Itoa(count)
	features += FeatureDelimiter + strconv.Itoa(len(times)-count)
	features += FeatureDelimiter +
		strconv.FormatFloat((times[len(times)-1]-times[0]), 'f', -1, 64)

	// position of the first 500 outgoing packets
	count = 0
	for i := 0; i < len(sizes); i++ {
		if sizes[i] > 0 {
			count++
			features += FeatureDelimiter + strconv.Itoa(i)
		}

		if count == 500 {
			break
		}
	}
	for i := count; i < 500; i++ {
		features += FeatureDelimiter + "'X'"
	}

	// difference in position between the first 500 outgoing packets
	// and the next outgoing packet
	count = 0
	prevloc := 0
	for i := 0; i < len(sizes); i++ {
		if sizes[i] > 0 {
			count++
			features += FeatureDelimiter + strconv.Itoa(i-prevloc)
			prevloc = i
		}
		if count == 500 {
			break
		}
	}
	for i := count; i < 500; i++ {
		features += FeatureDelimiter + "'X'"
	}

	// packet distributions (where are the outgoing packets concentrated)
	count = 0
	for i := 0; i < len(sizes) && i < 3000; i++ {
		if i%30 != 29 {
			if sizes[i] > 0 {
				count++
			}
		} else {
			features += FeatureDelimiter + strconv.Itoa(count)
			count = 0
		}
	}
	for i := len(sizes) / 30; i < 100; i++ {
		features += FeatureDelimiter + strconv.Itoa(0)
	}

	// Bursts (calc)
	var bursts []int
	outgoing := true // outgoing (positive) or incoming (negative)
	count = 0        // number of packets in the direction
	for i := 0; i < len(sizes); i++ {
		if sizes[i] > 0 == outgoing {
			// the packet goes in the same direction
			count++
		} else {
			// changing direction
			if count > 1 {
				// a burt is only defined for a sequence of packets
				bursts = append(bursts, count)
			}
			count = 1
			outgoing = sizes[i] > 0 // set direction
		}
	}
	max := -1
	sum := 0
	for i := 0; i < len(bursts); i++ {
		sum += bursts[i]
		if bursts[i] > max {
			max = bursts[i]
		}
	}
	// longest burst, mean size of burst, and number of bursts
	features += FeatureDelimiter + strconv.Itoa(max)
	if len(bursts) > 0 {
		features += FeatureDelimiter + strconv.Itoa(sum/len(bursts))
	} else {
		features += FeatureDelimiter + strconv.Itoa(0)
	}
	features += FeatureDelimiter + strconv.Itoa(len(bursts))

	// the number of bursts with lengths longer than 2,5,10,15,20,50
	counts := make([]int, 6)
	for i := 0; i < len(bursts); i++ {
		if bursts[i] > 2 {
			counts[0]++
		}
		if bursts[i] > 5 {
			counts[1]++
		}
		if bursts[i] > 10 {
			counts[2]++
		}
		if bursts[i] > 15 {
			counts[3]++
		}
		if bursts[i] > 20 {
			counts[4]++
		}
		if bursts[i] > 50 {
			counts[5]++
		}
	}
	for i := 0; i < len(counts); i++ {
		features += FeatureDelimiter + strconv.Itoa(counts[i])
	}

	// the length of the first 100 bursts
	for i := 0; i < 100; i++ {
		if len(bursts) > i {
			features += FeatureDelimiter + strconv.Itoa(bursts[i])
		} else {
			features += FeatureDelimiter + "'X'"
		}
	}

	// the direction of the first 10 packets
	// (we add MTU since -1 as feature is used internally)
	for i := 0; i < 10; i++ {
		if len(sizes) > i {
			features += FeatureDelimiter + strconv.Itoa(sizes[i]+1500)
		} else {
			features += FeatureDelimiter + "'X'"
		}
	}

	// interpacket timing: mean and standard deviation
	var total, variance float64
	current := times[0]
	for i := 1; i < len(times); i++ {
		total += times[i] - current
		current = times[i]
	}
	mean := total / float64((len(times) - 1))

	current = times[0]
	for i := 1; i < len(times); i++ {
		// -2 due to Bessel's correlation and interpacket timing def.
		variance += (times[i] - current) * (times[i] - current) /
			float64(len(times)-2)
		current = times[i]
	}

	features += FeatureDelimiter + strconv.FormatFloat((mean), 'f', -1, 64)
	features += FeatureDelimiter +
		strconv.FormatFloat((math.Sqrt(variance)), 'f', -1, 64)

	return
}
