package parser

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	trHeaderRe = regexp.MustCompile(`traceroute to\s+(\S+).*?(\d+)\s+hops max`)
	trHopRe    = regexp.MustCompile(`(?m)^\s*(\d+)\s+(.+)$`)
	trLatRe    = regexp.MustCompile(`([\d.]+)\s+ms`)
)

func SummarizeTraceroute(rawOutput string) string {
	var b strings.Builder

	if m := trHeaderRe.FindStringSubmatch(rawOutput); len(m) >= 3 {
		fmt.Fprintf(&b, "Destino: %s (max hops: %s)\n", m[1], m[2])
	}

	hops := trHopRe.FindAllStringSubmatch(rawOutput, -1)
	timeouts := 0
	lastHop := 0

	b.WriteString("Saltos:\n")
	for _, h := range hops {
		hopNum := h[1]
		rest := strings.TrimSpace(h[2])

		if strings.HasPrefix(rest, "traceroute") {
			continue
		}

		fmt.Sscanf(hopNum, "%d", &lastHop)

		if strings.Count(rest, "*") == 3 && !strings.ContainsAny(rest, "0123456789.") {
			timeouts++
			fmt.Fprintf(&b, "  %s  * * * (sin respuesta)\n", hopNum)
			continue
		}

		latencies := trLatRe.FindAllStringSubmatch(rest, -1)
		ipParts := regexp.MustCompile(`[\d]+\.[\d]+\.[\d]+\.[\d]+`).FindString(rest)

		if ipParts != "" && len(latencies) > 0 {
			var lats []string
			for _, l := range latencies {
				lats = append(lats, l[1])
			}
			fmt.Fprintf(&b, "  %s  %s  [%s ms]\n", hopNum, ipParts, strings.Join(lats, "/"))
		} else if ipParts != "" {
			fmt.Fprintf(&b, "  %s  %s\n", hopNum, ipParts)
		} else {
			fmt.Fprintf(&b, "  %s  %s\n", hopNum, rest)
		}
	}

	fmt.Fprintf(&b, "Total saltos: %d | Timeouts: %d", lastHop, timeouts)

	return strings.TrimSpace(b.String())
}
