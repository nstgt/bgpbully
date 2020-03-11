package bgpdbully

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/osrg/gobgp/pkg/packet/bgp"
)

const (
	OPERATION_CONNECT                  = "connect"
	OPERATION_SLEEP                    = "sleep"
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

type SleepParameter struct {
	Duration time.Duration `mapstructure:"sec"`
}

func (p SleepParameter) Serialize() {
}

func connect(globalConfig *Global) *net.Conn {
	conn, err := net.Dial("tcp", fmt.Sprintf("[%s]:%d", globalConfig.PeerIP, globalConfig.PeerPort))
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	log.Printf("connecting to %v", conn.RemoteAddr())
	return &conn
}

func splitBGP(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 || len(data) < bgp.BGP_HEADER_LENGTH {
		return 0, nil, nil
	}

	totalLen := binary.BigEndian.Uint16(data[16:18])
	if totalLen < bgp.BGP_HEADER_LENGTH {
		return 0, nil, fmt.Errorf("BGP Message length is too short: %d", totalLen)
	}
	if uint16(len(data)) < totalLen {
		return 0, nil, nil
	}
	return int(totalLen), data[0:totalLen], nil
}

func receiveBGPMessage(conn *net.Conn, bgpMsgCh chan *bgp.BGPMessage, closeCh chan struct{}) {
	scanner := bufio.NewScanner(bufio.NewReader((*conn)))
	scanner.Split(splitBGP)

	for scanner.Scan() {
		bgpMsg, err := bgp.ParseBGPMessage(scanner.Bytes())
		if err != nil {
			log.Printf("error: %v", err)
			continue
		}

		bgpMsgCh <- bgpMsg
	}
}

func sendBGPOpenMessage(conn *net.Conn, globalConfig *Global, params ParameterInterface) {
	log.Printf("send BGP Open Message")
	caps := make([]bgp.ParameterCapabilityInterface, 0)

	for _, v := range params.(OpenMessageParameters).Parameters {
		data, _ := hex.DecodeString(v.Data)
		cap := bgp.DefaultParameterCapability{
			CapCode:  bgp.BGPCapabilityCode(v.Capcode),
			CapLen:   uint8(len(data)),
			CapValue: data,
		}
		caps = append(caps, &cap)
	}
	opt := bgp.NewOptionParameterCapability(caps)
	msg := bgp.NewBGPOpenMessage((*globalConfig).LocalAS, (*globalConfig).Holdtime, (*globalConfig).LocalID, []bgp.OptionParameterInterface{opt})
	data, err := msg.Serialize()
	if err != nil {
		log.Fatalf("%v", err)
	}
	(*conn).Write(data)
}

func sendBGPKeepaliveMessage(conn *net.Conn) {
	log.Printf("send BGP Keepalive Message")
	msg := bgp.NewBGPKeepAliveMessage()
	data, _ := msg.Serialize()
	(*conn).Write(data)
}

func sleep(param ParameterInterface) {
	d := param.(SleepParameter).Duration
	log.Printf("sleep %v sec", int64(d))
	time.Sleep(d * time.Second)
}

func processSteps(config *Config, steps []Step, bgpMsgCh chan *bgp.BGPMessage, closeCh chan struct{}) {
	var conn *net.Conn

	for _, v := range steps {
		switch v.Operation {
		case OPERATION_CONNECT:
			conn = connect(&((*config).Global))
			go receiveBGPMessage(conn, bgpMsgCh, closeCh)
		case OPERATION_SEND_BGP_OPEN:
			sendBGPOpenMessage(conn, &((*config).Global), v.Parameter)
		case OPERATION_SEND_BGP_KEEPALIVE:
			sendBGPKeepaliveMessage(conn)
		case OPERATION_SLEEP:
			sleep(v.Parameter)
		default:
			time.Sleep(60 * time.Second) // will be deleted
			log.Printf("no such operation, exit")
			os.Exit(1)
		}
	}
}

func Run(configFile *string) {
	log.Printf("start")

	config := loadConfig(*configFile)
	steps := parseScenariosConfig(config)

	bgpMsgCh := make(chan *bgp.BGPMessage)
	closeCh := make(chan struct{})

	processSteps(config, steps, bgpMsgCh, closeCh)
}
