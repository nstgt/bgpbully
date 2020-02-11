package bgpdbully

import (
	"log"
)

func Run(configFile *string) {
	log.Printf("start")

	config := loadConfig(*configFile)
	parseConfig(config)
}
