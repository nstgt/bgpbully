package bgpdbully

func Run(configFile *string) {
	config := loadConfig(*configFile)
	parseConfig(config)
}
