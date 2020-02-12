package bgpdbully

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

const (
	OPERATION_CONNECT                  = "connect"
	OPERATION_SEND_BGP_OPEN            = "send_bgp_open"
	OPERATION_SEND_BGP_UPDATE          = "send_bgp_update"
	OPERATION_SEND_BGP_NOTIFICATION    = "send_bgp_notification"
	OPERATION_SEND_BGP_KEEPALIVE       = "send_bgp_keepalive"
	OPERATION_SEND_BGP_ROUTEREFRESH    = "send_bgp_routerefresh"
	OPERATION_RECEIVE_BGP_OPEN         = "receive_bgp_open"
	OPERATION_RECEIVE_BGP_UPDATE       = "receive_bgp_update"
	OPERATION_RECEIVE_BGP_NOTIFICATION = "receive_bgp_notification"
	OPERATION_RECEIVE_BGP_KEEPALIVE    = "receive_bgp_keepalive"
	OPERATION_RECEIVE_BGP_ROUTEREFRESH = "receive_bgp_routerefresh"
)

type Step struct {
	Operation Operation
	Parameter ParameterInterface
}

type ParameterInterface interface {
	Serialize()
}

type OpenMessageParameters struct {
	Parameters []OpenMessageParameter
}

func (o OpenMessageParameters) Serialize() {
}

type OpenMessageParameter struct {
	Type    string `mapstructure:"type"`
	Capcode int    `mapstructure:"capcode"`
	Data    string `mapstructure:"data"`
}

func connect(globalConfig *Global) net.Conn {
	conn, err := net.Dial("tcp", fmt.Sprintf("[%s]:%d", globalConfig.PeerIP, globalConfig.PeerPort))
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	log.Printf("connecting to %v", conn.RemoteAddr())
	return conn
}

func processSteps(config *Config, steps []Step) {
	for _, v := range steps {
		switch v.Operation {
		case OPERATION_CONNECT:
			connect(&((*config).Global))
		default:
			time.Sleep(10 * time.Second) // will be deleted
			log.Printf("no such operation, exit")
			os.Exit(1)
		}
	}
}

func Run(configFile *string) {
	log.Printf("start")

	config := loadConfig(*configFile)
	steps := parseScenariosConfig(config)
	//var conn net.Conn

	//bgpMsgCh := make(chan *bgp.BGPMessage)
	//closeCh := make(chan struct{})

	processSteps(config, steps)
}
