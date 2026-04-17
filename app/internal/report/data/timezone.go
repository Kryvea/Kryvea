package reportdata

import (
	"bufio"
	"os"
	"strings"
	"sync"
)

// Lazy index of IANA timezone -> ISO country code using tzdata's zone1970.tab.
var (
	zoneCountryOnce sync.Once
	zoneToCountry   map[string]string
)

func buildZoneCountryIndex() {
	zoneToCountry = map[string]string{}
	f, err := os.Open("/usr/share/zoneinfo/zone1970.tab")
	if err != nil {
		return // best-effort; fallback layouts will apply
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Format: cc[","cc...] TAB coordinates TAB TZ TAB comments
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}
		ccField := parts[0]
		tzField := parts[2]
		tz := strings.TrimSpace(tzField)
		if tz == "" {
			continue
		}
		// Take the first country code if multiple are present.
		cc := strings.TrimSpace(strings.Split(ccField, ",")[0])
		if cc == "" {
			continue
		}
		if _, exists := zoneToCountry[tz]; !exists {
			zoneToCountry[tz] = strings.ToUpper(cc)
		}
	}
}

func countryForZone(zone string) (string, bool) {
	zoneCountryOnce.Do(buildZoneCountryIndex)
	cc, ok := zoneToCountry[zone]
	return cc, ok
}

func layoutForCountry(cc string) string {
	switch strings.ToUpper(cc) {
	case "US":
		return "01/02/2006" // MM/DD/YYYY
	case "JP", "CN", "KR":
		return "2006/01/02" // YYYY/MM/DD
	default:
		return "02/01/2006" // DD/MM/YYYY
	}
}
