package main

import "fmt"

// Wa-kNN features
func featName(i int) string {
	switch {
	case i == 0:
		return "number of packets, feat 1"
	case i == 1:
		return "number of outgoing packets, feat 2"
	case i == 2:
		return "number of incoming packets, feat 3"
	case i == 3:
		return "total transmission time, feat 4"
	case i >= 4 && i < 504:
		return fmt.Sprintf("position of outgoing packet %d, feat %d", i-3, i+1)
	case i >= 504 && i < 1004:
		return fmt.Sprintf("difference in position between outgoing packet %d and next outgoing packet, feat %d", i-503, i+1)
	case i >= 1004 && i < 1104:
		return fmt.Sprintf("number of outgoing packets in window %d, feat %d",
			i-1003, i+1)
	case i == 1104:
		return "longest burst, feat 1105"
	case i == 1105:
		return "mean size of bursts, feat 1106"
	case i == 1106:
		return "number of bursts, feat 1107"
	case i == 1107:
		return "number of bursts longer than 2, feat 1108"
	case i == 1108:
		return "number of bursts longer than 5, feat 1109"
	case i == 1109:
		return "number of bursts longer than 10, feat 1110"
	case i == 1110:
		return "number of bursts longer than 15, feat 1111"
	case i == 1111:
		return "number of bursts longer than 20, feat 1112"
	case i == 1112:
		return "number of bursts longer than 50, feat 1113"
	case i >= 1113 && i < 1213:
		return fmt.Sprintf("length of burst %d, feat %d", i-1112, i+1)
	case i >= 1213 && i < 1223:
		return fmt.Sprintf("direction of packet %d, feat %d", i-1211, i+1)
	case i == 1223:
		return "mean interpacket time, feat 1224"
	case i == 1224:
		return "std of interpacket time, feat 1225"
	}

	return "unknown"
}
