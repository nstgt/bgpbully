package bgpbully

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/osrg/gobgp/pkg/packet/bgp"
)

const (
	OPERATION_CONNECT                  = "connect"
	OPERATION_SLEEP                    = "sleep"
	OPERATION_CLOSE                    = "close"
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

type PeerInfo struct {
	IP   string
	Port int
}

type LocalInfo struct {
	Holdtime uint16
	AS       uint16
	ID       string
}

type Step struct {
	Operation Operation
	Parameter ParameterInterface
}

type ParameterInterface interface {
	Serialize() ([]byte, error)
}

type OpenMessageParameter struct {
	AS           uint16
	Holdtime     uint16
	ID           string
	Capabilities []Capability `mapstructure:"capabilities"`
}

type Capability struct {
	Type  int    `mapstructure:"type"`
	Value string `mapstructure:"value"`
}

func (p OpenMessageParameter) Serialize() ([]byte, error) {
	caps := make([]bgp.ParameterCapabilityInterface, 0)
	for _, v := range p.Capabilities {
		data, _ := hex.DecodeString(v.Value)
		cap := bgp.DefaultParameterCapability{
			CapCode:  bgp.BGPCapabilityCode(v.Type),
			CapLen:   uint8(len(data)),
			CapValue: data,
		}
		caps = append(caps, &cap)
	}
	opt := bgp.NewOptionParameterCapability(caps)
	msg := bgp.NewBGPOpenMessage(p.AS, p.Holdtime, p.ID, []bgp.OptionParameterInterface{opt})
	data, err := msg.Serialize()
	return data, err
}

type UpdateMessageParameter struct {
	WithdrawnRoutes []string        `mapstructure:"withdrawn_routes"`
	PathAttributes  []PathAttribute `mapstructure:"path_attributes"`
	NLRI            []string        `mapstructure:"nlri"`
}

type PathAttribute struct {
	Flag   string `mapstructure:"flag"`
	Type   uint8  `mapstructure:"type"`
	Length uint16 `mapstructure:"len"`
	Value  string `mapstructure:"value"`
}

func (p UpdateMessageParameter) Serialize() ([]byte, error) {
	var withdrawnRoutes []*bgp.IPAddrPrefix
	for _, v := range p.WithdrawnRoutes {
		i := convertIPfromStringToIPAddrPrefix(v)
		withdrawnRoutes = append(withdrawnRoutes, i)
	}

	var nlri []*bgp.IPAddrPrefix
	for _, v := range p.NLRI {
		i := convertIPfromStringToIPAddrPrefix(v)
		nlri = append(nlri, i)
	}

	var pathattrs []bgp.PathAttributeInterface
	for _, v := range p.PathAttributes {
		flag, _ := strconv.ParseUint(v.Flag, 16, 8)
		value, _ := hex.DecodeString(v.Value)
		pa := bgp.NewPathAttributeUnknown(bgp.BGPAttrFlag(uint8(flag)), bgp.BGPAttrType(v.Type), value)
		pathattrs = append(pathattrs, pa)
	}

	msg := &bgp.BGPMessage{
		Header: bgp.BGPHeader{Type: bgp.BGP_MSG_UPDATE},
		Body: &bgp.BGPUpdate{
			WithdrawnRoutesLen:    0,
			WithdrawnRoutes:       withdrawnRoutes,
			TotalPathAttributeLen: 0,
			PathAttributes:        pathattrs,
			NLRI:                  nlri,
		},
	}
	data, err := msg.Serialize()
	return data, err
}

type NotificationMessageParameter struct {
	Code    uint8 `mapstructure:"code"`
	SubCode uint8 `mapstructure:"subcode"`
}

func (p NotificationMessageParameter) Serialize() ([]byte, error) {
	msg := bgp.NewBGPNotificationMessage(p.Code, p.SubCode, nil)
	data, err := msg.Serialize()
	return data, err
}

type RouterefreshMessageParameter struct {
	AFI  uint16 `mapstructure:"afi"`
	SAFI uint8  `mapstructure:"safi"`
}

func (p RouterefreshMessageParameter) Serialize() ([]byte, error) {
	msg := bgp.NewBGPRouteRefreshMessage(p.AFI, 0, p.SAFI)
	data, err := msg.Serialize()
	if err != nil {
		log.Fatalf("%v", err)
	}
	return data, err
}

type SleepParameter struct {
	Duration time.Duration `mapstructure:"sec"`
}

func (p SleepParameter) Serialize() ([]byte, error) {
	return nil, nil
}

type IPAddrPrefix struct {
	bgp.IPAddrPrefixDefault
	addrlen uint8
}

func connect(peer PeerInfo) *net.Conn {
	conn, err := net.Dial("tcp", fmt.Sprintf("[%s]:%d", peer.IP, peer.Port))
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	log.Printf("connecting to %v", conn.RemoteAddr())
	return &conn
}

func close(conn *net.Conn) {
	log.Printf("closing connection to %v", (*conn).RemoteAddr())
	(*conn).Close()
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

func acceptArrivalBGPMessage(conn *net.Conn, bgpMsgCh chan *bgp.BGPMessage, closeCh chan struct{}) {
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

func receiveBGPMessage(bgpMsgCh chan *bgp.BGPMessage, expectedMsgType uint8) {
	bgpMsg := <-bgpMsgCh

	if bgpMsg.Header.Type == expectedMsgType {
		log.Printf("receive BGP Message, type %v", expectedMsgType)
	} else {
		log.Printf("error: expected type %v, got type %v", expectedMsgType, bgpMsg.Header.Type)
		os.Exit(1)
	}
}

func sendBGPOpenMessage(conn *net.Conn, param ParameterInterface) {
	log.Printf("send BGP Open Message")
	data, err := param.(OpenMessageParameter).Serialize()
	if err != nil {
		log.Fatalf("%v", err)
	}
	(*conn).Write(data)
}

func sendBGPUpdateMessage(conn *net.Conn, param ParameterInterface) {
	log.Printf("send BGP Update Message")
	data, err := param.(UpdateMessageParameter).Serialize()
	if err != nil {
		log.Fatalf("%v", err)
	}
	(*conn).Write(data)
}

func sendBGPNotificationMessage(conn *net.Conn, param ParameterInterface) {
	log.Printf("send BGP Notification Message")
	data, err := param.(NotificationMessageParameter).Serialize()
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

func sendBGPRouteRefreshMessage(conn *net.Conn, param ParameterInterface) {
	log.Printf("send BGP RouteRefresh Message")
	data, err := param.(RouterefreshMessageParameter).Serialize()
	if err != nil {
		log.Fatalf("%v", err)
	}
	(*conn).Write(data)
}

func convertIPfromStringToIPAddrPrefix(ip string) *bgp.IPAddrPrefix {
	addr := strings.Split(ip, "/")[0]
	len, _ := strconv.ParseUint(strings.Split(ip, "/")[1], 10, 8)
	ipap := bgp.NewIPAddrPrefix(uint8(len), addr)
	return ipap
}

func sleep(param ParameterInterface) {
	d := param.(SleepParameter).Duration
	log.Printf("sleep %v sec", int64(d))
	time.Sleep(d * time.Second)
}

func processSteps(peer PeerInfo, local LocalInfo, steps []Step, bgpMsgCh chan *bgp.BGPMessage, closeCh chan struct{}) {
	var conn *net.Conn

	for _, v := range steps {
		switch v.Operation {
		case OPERATION_CONNECT:
			conn = connect(peer)
			go acceptArrivalBGPMessage(conn, bgpMsgCh, closeCh)
		case OPERATION_CLOSE:
			close(conn)
		case OPERATION_SEND_BGP_OPEN:
			vv := v.Parameter.(OpenMessageParameter)
			vv.AS = local.AS
			vv.ID = local.ID
			vv.Holdtime = local.Holdtime
			sendBGPOpenMessage(conn, vv)
		case OPERATION_SEND_BGP_UPDATE:
			sendBGPUpdateMessage(conn, v.Parameter)
		case OPERATION_SEND_BGP_NOTIFICATION:
			sendBGPNotificationMessage(conn, v.Parameter)
		case OPERATION_SEND_BGP_KEEPALIVE:
			sendBGPKeepaliveMessage(conn)
		case OPERATION_SEND_BGP_ROUTEREFRESH:
			sendBGPRouteRefreshMessage(conn, v.Parameter)
		case OPERATION_RECEIVE_BGP_OPEN:
			receiveBGPMessage(bgpMsgCh, bgp.BGP_MSG_OPEN)
		case OPERATION_RECEIVE_BGP_UPDATE:
			receiveBGPMessage(bgpMsgCh, bgp.BGP_MSG_UPDATE)
		case OPERATION_RECEIVE_BGP_NOTIFICATION:
			receiveBGPMessage(bgpMsgCh, bgp.BGP_MSG_NOTIFICATION)
		case OPERATION_RECEIVE_BGP_KEEPALIVE:
			receiveBGPMessage(bgpMsgCh, bgp.BGP_MSG_KEEPALIVE)
		case OPERATION_RECEIVE_BGP_ROUTEREFRESH:
			receiveBGPMessage(bgpMsgCh, bgp.BGP_MSG_ROUTE_REFRESH)
		case OPERATION_SLEEP:
			sleep(v.Parameter)
		default:
			log.Printf("no such operation, exit")
			os.Exit(1)
		}
	}
}

func Run(configFile string) {
	log.Printf("start")

	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("error: %v", err)
		os.Exit(1)
	}
	peer, local := parseGlobal(config)
	steps := parseScenario(config)

	bgpMsgCh := make(chan *bgp.BGPMessage)
	closeCh := make(chan struct{})

	processSteps(peer, local, steps, bgpMsgCh, closeCh)
}
