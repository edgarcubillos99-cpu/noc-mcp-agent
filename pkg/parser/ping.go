package parser

import (
	"fmt"
	"regexp"
	"strings"
)

type PingSummary struct {
	Target      string
	Transmitted int
	Received    int
	LossPercent string
	RTTMin      string
	RTTAvg      string
	RTTMax      string
	Reachable   bool
}

var (
	pingStatsRe = regexp.MustCompile(`(\d+)\s+packets?\s+transmitted,\s+(\d+)\s+received.*?(\d+(?:\.\d+)?%)\s+packet loss`)
	pingRTTRe   = regexp.MustCompile(`rtt\s+min/avg/max/\S+\s*=\s*([\d.]+)/([\d.]+)/([\d.]+)`)
)

func SummarizePing(target, rawOutput string) string {
	s := PingSummary{Target: target}

	if m := pingStatsRe.FindStringSubmatch(rawOutput); len(m) >= 4 {
		fmt.Sscanf(m[1], "%d", &s.Transmitted)
		fmt.Sscanf(m[2], "%d", &s.Received)
		s.LossPercent = m[3]
		s.Reachable = s.Received > 0
	}

	if m := pingRTTRe.FindStringSubmatch(rawOutput); len(m) >= 4 {
		s.RTTMin = m[1]
		s.RTTAvg = m[2]
		s.RTTMax = m[3]
	}

	var b strings.Builder
	if s.Reachable {
		fmt.Fprintf(&b, "[OK] %s — %d/%d recibidos, pérdida: %s",
			s.Target, s.Received, s.Transmitted, s.LossPercent)
		if s.RTTAvg != "" {
			fmt.Fprintf(&b, ", RTT avg: %s ms (min: %s, max: %s)", s.RTTAvg, s.RTTMin, s.RTTMax)
		}
	} else {
		fmt.Fprintf(&b, "[FALLO] %s — 0/%d recibidos, pérdida: %s (host inalcanzable)",
			s.Target, s.Transmitted, s.LossPercent)
	}

	return b.String()
}
