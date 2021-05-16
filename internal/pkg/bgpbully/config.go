package bgpbully

import (
	"log"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Config struct {
	Global   GlobalConfig `mapstructure:"global"`
	Scenario []StepConfig `mapstructure:"scenario"`
}

type GlobalConfig struct {
	PeerIP   string `mapstructure:"peer_ip"`
	PeerPort int    `mapstructure:"peer_port"`
	Holdtime uint16 `mapstructure:"holdtime"`
	LocalAS  uint16 `mapstructure:"local_as"`
	LocalID  string `mapstructure:"local_id"`
}

type StepConfig struct {
	Operation Operation                   `mapstructure:"ope"`
	Parameter map[interface{}]interface{} `mapstructure:"param"`
}

type Operation string

func loadConfig(path string) (*Config, error) {
	config := &Config{}
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	var err error
	if err = v.ReadInConfig(); err != nil {
		return nil, err
	}
	if err = v.UnmarshalExact(config); err != nil {
		return nil, err
	}
	return config, nil
}

func parseGlobal(config *Config) (PeerInfo, LocalInfo) {
	peer := PeerInfo{
		IP:   config.Global.PeerIP,
		Port: config.Global.PeerPort,
	}
	local := LocalInfo{
		Holdtime: config.Global.Holdtime,
		AS:       config.Global.LocalAS,
		ID:       config.Global.LocalID,
	}
	return peer, local
}

func parseScenario(sc []StepConfig) []Step {
	var steps []Step
	for _, v := range sc {
		// TODO: extract common codes and remove them to other function
		switch v.Operation {
		case OPERATION_CONNECT:
			s := Step{
				Operation: OPERATION_CONNECT,
				Parameter: nil,
			}
			steps = append(steps, s)
		case OPERATION_CLOSE:
			s := Step{
				Operation: OPERATION_CLOSE,
				Parameter: nil,
			}
			steps = append(steps, s)
		case OPERATION_SEND_BGP_OPEN:
			var p OpenMessageParameter
			err := mapstructure.Decode(v.Parameter, &p)
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			s := Step{
				Operation: OPERATION_SEND_BGP_OPEN,
				Parameter: p,
			}
			steps = append(steps, s)
		case OPERATION_SEND_BGP_UPDATE:
			var p UpdateMessageParameter
			err := mapstructure.Decode(v.Parameter, &p)
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			s := Step{
				Operation: OPERATION_SEND_BGP_UPDATE,
				Parameter: p,
			}
			steps = append(steps, s)
		case OPERATION_SEND_BGP_NOTIFICATION:
			var p NotificationMessageParameter
			err := mapstructure.Decode(v.Parameter, &p)
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			s := Step{
				Operation: OPERATION_SEND_BGP_NOTIFICATION,
				Parameter: p,
			}
			steps = append(steps, s)
		case OPERATION_SEND_BGP_KEEPALIVE:
			s := Step{
				Operation: OPERATION_SEND_BGP_KEEPALIVE,
				Parameter: nil,
			}
			steps = append(steps, s)
		case OPERATION_SEND_BGP_ROUTEREFRESH:
			var p RouterefreshMessageParameter
			err := mapstructure.Decode(v.Parameter, &p)
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			s := Step{
				Operation: OPERATION_SEND_BGP_ROUTEREFRESH,
				Parameter: p,
			}
			steps = append(steps, s)
		case OPERATION_SEND_BGP_RAW:
			var p RawMessageParameter
			err := mapstructure.Decode(v.Parameter, &p)
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			s := Step{
				Operation: OPERATION_SEND_BGP_RAW,
				Parameter: p,
			}
			steps = append(steps, s)
		case OPERATION_RECEIVE_BGP_OPEN:
			s := Step{
				Operation: OPERATION_RECEIVE_BGP_OPEN,
				Parameter: nil,
			}
			steps = append(steps, s)
		case OPERATION_RECEIVE_BGP_UPDATE:
			s := Step{
				Operation: OPERATION_RECEIVE_BGP_UPDATE,
				Parameter: nil,
			}
			steps = append(steps, s)
		case OPERATION_RECEIVE_BGP_NOTIFICATION:
			s := Step{
				Operation: OPERATION_RECEIVE_BGP_NOTIFICATION,
				Parameter: nil,
			}
			steps = append(steps, s)
		case OPERATION_RECEIVE_BGP_KEEPALIVE:
			s := Step{
				Operation: OPERATION_RECEIVE_BGP_KEEPALIVE,
				Parameter: nil,
			}
			steps = append(steps, s)
		case OPERATION_RECEIVE_BGP_ROUTEREFRESH:
			s := Step{
				Operation: OPERATION_RECEIVE_BGP_ROUTEREFRESH,
				Parameter: nil,
			}
			steps = append(steps, s)
		case OPERATION_RECEIVE_NOTHING:
			var p NothingParameter
			err := mapstructure.Decode(v.Parameter, &p)
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			s := Step{
				Operation: OPERATION_RECEIVE_NOTHING,
				Parameter: p,
			}
			steps = append(steps, s)
		case OPERATION_RECEIVE_ONE_OF_THEM:
			var p BundledStepsParameter
			err := mapstructure.Decode(v.Parameter, &p)
			if err != nil {
				log.Fatalf("error: %v", err)
			}

			pp := parseScenario(p.Steps)
			s := Step{
				Operation: OPERATION_RECEIVE_ONE_OF_THEM,
				Parameter: ReceiveOneOfThemParameter(pp),
			}
			steps = append(steps, s)
		case OPERATION_SLEEP:
			var p SleepParameter
			err := mapstructure.Decode(v.Parameter, &p)
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			s := Step{
				Operation: OPERATION_SLEEP,
				Parameter: p,
			}
			steps = append(steps, s)
		default:
			log.Fatalf("error: no such operation %v", v.Operation)
			os.Exit(1)
		}
	}
	return steps
}
