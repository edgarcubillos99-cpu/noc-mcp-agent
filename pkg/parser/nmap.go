package parser

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	nmapHostRe   = regexp.MustCompile(`Nmap scan report for\s+(.+)`)
	nmapStateRe  = regexp.MustCompile(`Host is (\w+)`)
	nmapPortRe   = regexp.MustCompile(`(?m)^(\d+/\w+)\s+(open|closed|filtered)\s+(.*)$`)
	nmapTimeRe   = regexp.MustCompile(`scanned in\s+([\d.]+)\s+seconds`)
	nmapNotShown = regexp.MustCompile(`Not shown:\s+(\d+)\s+(\w+)\s+ports?`)
)

func SummarizeNmap(rawOutput string) string {
	var b strings.Builder

	if m := nmapHostRe.FindStringSubmatch(rawOutput); len(m) >= 2 {
		fmt.Fprintf(&b, "Host: %s", m[1])
	}

	if m := nmapStateRe.FindStringSubmatch(rawOutput); len(m) >= 2 {
		fmt.Fprintf(&b, " | Estado: %s", m[1])
	}
	b.WriteString("\n")

	if m := nmapNotShown.FindStringSubmatch(rawOutput); len(m) >= 3 {
		fmt.Fprintf(&b, "No mostrados: %s puertos %s\n", m[1], m[2])
	}

	ports := nmapPortRe.FindAllStringSubmatch(rawOutput, -1)
	if len(ports) > 0 {
		b.WriteString("Puertos:\n")
		for _, p := range ports {
			svc := strings.TrimSpace(p[3])
			if svc == "" {
				svc = "desconocido"
			}
			fmt.Fprintf(&b, "  %s  %s  (%s)\n", p[1], p[2], svc)
		}
	} else {
		b.WriteString("No se detectaron puertos abiertos/filtrados.\n")
	}

	if m := nmapTimeRe.FindStringSubmatch(rawOutput); len(m) >= 2 {
		fmt.Fprintf(&b, "Duración del escaneo: %s s", m[1])
	}

	return strings.TrimSpace(b.String())
}
