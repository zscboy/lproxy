package servercfg

import (
	"bufio"
	log "github.com/sirupsen/logrus"
	"os"
)

var (
	domains []string = make([]string, 0)
)

func loadDomainsFromFile(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		domains = append(domains, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
