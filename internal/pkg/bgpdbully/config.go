package bgpdbully

import (
	"log"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Config struct {
	Global    Global     `mapstructure:"global"`
	Scenarios []Scenario `mapstructure:"scenarios"`
}

type Global struct {
	PeerIP   string `mapstructure:"peer_ip"`
	PeerPort int    `mapstructure:"peer_port"`
	Holdtime uint16 `mapstructure:"holdtime"`
	LocalAS  uint16 `mapstructure:"local_as"`
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
		log.Printf("error: %v", err)
		os.Exit(1)
	}

	return config
}

func parseScenariosConfig(config *Config) []Step {
	var steps []Step
	for _, v := range config.Scenarios {
		switch v.Operation {
		case OPERATION_CONNECT:
			s := Step{
				Operation: OPERATION_CONNECT,
				Parameter: nil,
			}
			steps = append(steps, s)
		case OPERATION_SEND_BGP_OPEN:
			var params OpenMessageParameters
			for _, vv := range v.Parameters {
				var p OpenMessageParameter
				err := mapstructure.Decode(vv, &p)
				if err != nil {
					log.Fatalf("error: %v", err)
				}
				params.Parameters = append(params.Parameters, p)
			}
			s := Step{
				Operation: OPERATION_SEND_BGP_OPEN,
				Parameter: params,
			}
			steps = append(steps, s)
		case OPERATION_RECEIVE_BGP_OPEN:
		case OPERATION_SEND_BGP_UPDATE:
		case OPERATION_RECEIVE_BGP_UPDATE:
		case OPERATION_SEND_BGP_NOTIFICATION:
		case OPERATION_RECEIVE_BGP_NOTIFICATION:
		case OPERATION_SEND_BGP_KEEPALIVE:
			s := Step{
				Operation: OPERATION_SEND_BGP_KEEPALIVE,
				Parameter: nil,
			}
			steps = append(steps, s)
		case OPERATION_RECEIVE_BGP_KEEPALIVE:
		case OPERATION_SEND_BGP_ROUTEREFRESH:
		case OPERATION_RECEIVE_BGP_ROUTEREFRESH:
		case OPERATION_SLEEP:
			var p SleepParameter
			err := mapstructure.Decode(v.Parameters[0], &p)
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			s := Step{
				Operation: OPERATION_SLEEP,
				Parameter: p,
			}
			steps = append(steps, s)
		default:
		}
	}
	return steps
}
