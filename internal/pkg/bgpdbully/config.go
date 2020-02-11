package bgpdbully

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Global    Global     `mapstructure:"global"`
	Scenarios []Scenario `mapstructure:"scenarios"`
}

type Global struct {
	PeerIP   string `mapstructure:"peer_ip"`
	PeerPort int    `mapstructure:"peer_port"`
	Holdtime int    `mapstructure:"holdtime"`
	LocalAS  int32  `mapstructure:"local_as"`
	LocalID  string `mapstructure:"local_id"`
}

type Scenario struct {
	Operation  Operation                     `mapstructure:"ope"`
	Parameters []map[interface{}]interface{} `mapstructure:"param"`
}

type Operation string

func loadConfig(configFile string) *Config {
	viper.SetConfigName(configFile)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("error: %v", err)
		os.Exit(1)
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		log.Printf("error: configuration unmarshal error %v", err)
		os.Exit(1)
	}

	return config
}

func parseConfig(config *Config) {
	fmt.Println(config)
}
