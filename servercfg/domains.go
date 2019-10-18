package servercfg

import (
	"bufio"
	"os"

	"github.com/blang/semver"
	log "github.com/sirupsen/logrus"
)

var (
	domains []string = make([]string, 0)

	// DomainsCfgVer domains txt file version
	DomainsCfgVer = semver.MustParse("0.1.0")
	// DomainsCfgVerStr domains txt file version
	DomainsCfgVerStr = "0.1.0"
)

func loadDomainsFromFile(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	// version
	if scanner.Scan() {
		text := scanner.Text()
		v, e := semver.Make(text)
		if e == nil {
			log.Println("domains file version:", text)
			DomainsCfgVer = v
			DomainsCfgVerStr = text
		} else {
			log.Println("domains file get version failed:", e)
		}
	}

	for scanner.Scan() {
		domains = append(domains, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

// GetDomains get domains cfg
func GetDomains() []string {
	return domains
}
