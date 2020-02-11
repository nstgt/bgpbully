package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/osrg/gobgp/pkg/packet/bgp"
	"github.com/spf13/viper"
	//"gopkg.in/yaml.v3"
)

//type Parameter interface {
//	Serialize() ([]byte, error)
//}

type OpenMessageParameter struct {
	Type    string `mapstructure:"type"`
	Capcode string `mapstructure:"capcode"`
	Length  int    `mapstructure:"length"`
	Data    string `mapstructure:"data"`
}

func (o OpenMessageParameter) Serialize() ([]byte, error) {
	return nil, nil
}

type UpdateMessageParameter struct {
	Code string `mapstructure:"code"`
}

func (o UpdateMessageParameter) Serialize() ([]byte, error) {
	return nil, nil
}

type CapabilityParameter struct {
	Type    string `mapstructure:"type"`
	Capcode int    `mapstructure:"capcode"`
	Length  int    `mapstructure:"len"`
	Data    string `mapstructure:"data"`
}

func readConfig(configfile string) {
	viper.SetConfigName("sample")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("config file read error")
		fmt.Println(err)
		os.Exit(1)
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		fmt.Println("config file Unmarshal error")
		fmt.Println(err)
		os.Exit(1)
	}
	//fmt.Println(config)
	for _, v := range config.Scenarios {
		fmt.Printf("%v\n", v.Operation)
		for _, vv := range v.Parameters {
			var cp CapabilityParameter
			err := mapstructure.Decode(vv, &cp)
			if err != nil {
				panic(err)
			}
			fmt.Println(cp)
		}
	}
	//fmt.Println(config.Scenarios)
	//p := config.Scenarios[2].Parameters[0]
	//fmt.Println(p)

	//m := make(map[string]interface{})
	//for k, v := range p {
	//	strKey := fmt.Sprintf("%v", k)
	//
	//	m[strKey] = v
	//}
	//fmt.Println(m["capcode"])
	//fmt.Println(p.(map[interface{}]interface{})["capcode"])
	//switch val := p.(type) {
	//case OpenMessageParameter:
	//	fmt.Println("this is OpenMessage: %v", val)
	//case UpdateMessageParameter:
	//	fmt.Println("this is UpdateMessage: %v", val)
	//default:
	//	fmt.Println("this is Unknown Message Type")
	//}
	//fmt.Println()
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

//func readConfig(configfile string) *map[interface{}]interface{} {
//
//	buf, err := ioutil.ReadFile(configfile)
//	if err != nil {
//		panic(err)
//	}
//
//	m := make(map[interface{}]interface{})
//	err = yaml.Unmarshal(buf, &m)
//	if err != nil {
//		panic(err)
//	}
//
//	return &m
//}

type BGPInfo struct {
	peer_ip   string
	peer_port int
	holdtime  uint16
	local_as  uint16
	local_id  string
}

//type Scenario struct {
//	operation string
//	parameter interface{}
//}

func init() {
	//log.SetFormatter(&log.TextFormatter{
	//	FullTimestamp:   true,
	//	TimestampFormat: "2006/01/02 15:04:05",
	//})
	//log.SetLevel(log.DebugLevel)
}

//func parse_scenario(config map[interface{}]interface{}) (BGPInfo, []Scenario) {
//
//	bgpinfo := BGPInfo{
//		config["general"].(map[interface{}]interface{})["peer_ip"].(string),
//		config["general"].(map[interface{}]interface{})["peer_port"].(int),
//		uint16(reflect.ValueOf(config["general"].(map[interface{}]interface{})["holdtime"]).Int()),
//		uint16(reflect.ValueOf(config["general"].(map[interface{}]interface{})["local_as"]).Int()),
//		//		config["general"].(map[interface{}]interface{})["holdtime"].(uint16),
//		//		config["general"].(map[interface{}]interface{})["local_as"].(uint16),
//		config["general"].(map[interface{}]interface{})["local_id"].(string),
//	}
//
//	array := config["scenario"].([]interface{})
//	scenario_queue := make([]Scenario, 0)
//
//	for _, val := range array {
//		scenario := Scenario{
//			val.(map[interface{}]interface{})["ope"].(string),
//			val.(map[interface{}]interface{})["param"],
//		}
//		scenario_queue = append(scenario_queue, scenario)
//	}
//
//	return bgpinfo, scenario_queue
//}

func connect(bgpinfo BGPInfo) net.Conn {

	conn, err := net.Dial("tcp", fmt.Sprintf("[%s]:%d", bgpinfo.peer_ip, bgpinfo.peer_port))
	if err != nil {
		log.Fatalf("%v", err)
	}

	fmt.Println(" >>> connected: %v", conn.RemoteAddr())
	return conn
}

func wait(param interface{}) {

	var waittime int64
	if param == nil {
		waittime = int64(60)
	} else {
		wait_param, err := param.(map[interface{}]interface{})["waittime"]
		if !err {
			fmt.Println("Error: While reading waittime.")
		}
		waittime = int64(reflect.ValueOf(wait_param).Int())
	}

	fmt.Println(" >>> waiting %d sec.", waittime)
	time.Sleep(time.Duration(waittime) * time.Second)
}

func close(conn net.Conn) {
	conn.Close()
	fmt.Println(" >>> closed: %v", conn.RemoteAddr())
}

func print_screen(param interface{}) {
	content, err := param.(map[interface{}]interface{})["content"]
	if !err {
		fmt.Println("Error: While reading content.")
	}

	fmt.Println(" >>> print: %v", content)
}

func receive_bgp_msg(conn net.Conn, bgpMsg_ch chan *bgp.BGPMessage, exit_ch chan bool) {
	scanner := bufio.NewScanner(bufio.NewReader(conn))
	scanner.Split(splitBGP)

	fmt.Println(" >>> receiver goroutine started.")
	for scanner.Scan() {

		bgpMsg, err := bgp.ParseBGPMessage(scanner.Bytes())
		if err != nil {
			fmt.Println("error")
			continue
		}
		//fmt.Println("DEBUG: goroutine received >>> %#v", bgpMsg)
		bgpMsg_ch <- bgpMsg

		//		select {
		//		case <-exit_ch:
		//			break
		//		}
		//		if <-exit_ch {
		//			fmt.Println("EXIT!!")
		//			break
		//		}
	}
}

func receive_bgp_open(bgpMsg_ch chan *bgp.BGPMessage) {

	var expected_type uint8 = bgp.BGP_MSG_OPEN
	bgpMsg := <-bgpMsg_ch

	if bgpMsg.Header.Type == expected_type {
		bgpSpecifiedMsg := bgpMsg.Body.(*bgp.BGPOpen)
		fmt.Println(" >>> got OPEN: %#v", bgpSpecifiedMsg)
	} else {
		fmt.Println(" <<< Missmatching with Scenario. Exit Program. >>>")
		fmt.Println("   >>> Expected: %#v, But Received: %#v", expected_type, bgpMsg.Header.Type)
		os.Exit(1)
	}
}

func receive_bgp_update(bgpMsg_ch chan *bgp.BGPMessage) {

	var expected_type uint8 = bgp.BGP_MSG_UPDATE
	bgpMsg := <-bgpMsg_ch

	if bgpMsg.Header.Type == bgp.BGP_MSG_UPDATE {
		bgpSpecifiedMsg := bgpMsg.Body.(*bgp.BGPUpdate)
		fmt.Println(" >>> got UPDATE: %#v", bgpSpecifiedMsg)
	} else {
		fmt.Println(" <<< Missmatching with Scenario. Exit Program. >>>")
		fmt.Println("   >>> Expected: %#v, But Received: %#v", expected_type, bgpMsg.Header.Type)
		os.Exit(1)
	}
}

func receive_bgp_keepalive(bgpMsg_ch chan *bgp.BGPMessage) {

	var expected_type uint8 = bgp.BGP_MSG_KEEPALIVE
	bgpMsg := <-bgpMsg_ch

	if bgpMsg.Header.Type == bgp.BGP_MSG_KEEPALIVE {
		bgpSpecifiedMsg := bgpMsg.Body.(*bgp.BGPKeepAlive)
		fmt.Println(" >>> got KEEPALIVE: %#v", bgpSpecifiedMsg)
	} else {
		fmt.Println(" <<< Missmatching with Scenario. Exit Program. >>>")
		fmt.Println("   >>> Expected: %#v, But Received: %#v", expected_type, bgpMsg.Header.Type)
		os.Exit(1)
	}
}

func receive_bgp_notification(bgpMsg_ch chan *bgp.BGPMessage) {

	var expected_type uint8 = bgp.BGP_MSG_NOTIFICATION
	bgpMsg := <-bgpMsg_ch

	if bgpMsg.Header.Type == bgp.BGP_MSG_NOTIFICATION {
		bgpSpecifiedMsg := bgpMsg.Body.(*bgp.BGPNotification)
		fmt.Println(" >>> got NOTIFICATION: %#v", bgpSpecifiedMsg)
	} else {
		fmt.Println(" <<< Missmatching with Scenario. Exit Program. >>>")
		fmt.Println("   >>> Expected: %#v, But Received: %#v", expected_type, bgpMsg.Header.Type)
		os.Exit(1)
	}
}

func send_bgp_open(conn net.Conn, bgpinfo BGPInfo, param interface{}) {

	caps := make([]bgp.ParameterCapabilityInterface, 0)
	param_array := param.([]interface{})

	// Capabilityを確認
	for _, val := range param_array {
		one_param_map := val.(map[interface{}]interface{})

		capcode := uint8(reflect.ValueOf(one_param_map["capcode"]).Int())
		len := uint8(reflect.ValueOf(one_param_map["len"]).Int())

		var data []byte

		//とてもイケてないコード
		list := strings.Split(one_param_map["data"].(string), "")
		str := ""
		for i, c := range list {
			if i > 0 && i%2 == 0 {
				byte_data, err := strconv.ParseUint(str, 16, 8)
				if err != nil {
					fmt.Printf("Error: %s", err)
					os.Exit(1)
				}
				data = append(data, uint8(byte_data))
				str = ""
			}
			str += c
		}
		byte_data, _ := strconv.ParseUint(str, 16, 8)
		data = append(data, uint8(byte_data))

		//このコードの方が完結だけど、len(runes)がnon-functionで怒られるため後で考える
		//	    splitlen := 2
		//  	var hoge string = "hoge"//one_param_map["data"].(string)
		//   	bytes := []byte(one_param_map["data"].(string))
		//		for i := 0; i < len(bytes); i += splitlen {
		//			if i+splitlen < len(bytes) {
		//				fmt.Println("hoge")
		//				fmt.Println(string(bytes[i:(i + splitlen)]))
		//				byte_data, err := strconv.ParseUint(string(bytes[i:(i + splitlen)]), 16, 8)
		//				if err != nil {
		//					fmt.Printf("Error: %s", err)
		//					os.Exit(1)
		//				}
		//				data = append(data, uint8(byte_data))
		//			} else {
		//				fmt.Println("fuga")
		//				fmt.Println(string(bytes[i:]))
		//			}
		//		}

		cap := bgp.DefaultParameterCapability{
			CapCode:  bgp.BGPCapabilityCode(capcode),
			CapLen:   len,
			CapValue: data,
		}

		caps = append(caps, &cap)
	}

	opt := bgp.NewOptionParameterCapability(caps)
	msg := bgp.NewBGPOpenMessage(bgpinfo.local_as, bgpinfo.holdtime, bgpinfo.local_id, []bgp.OptionParameterInterface{opt})
	data, _ := msg.Serialize()
	conn.Write(data)

	fmt.Println(" >>> sent OPEN: %#v", msg.Body)
}

func send_bgp_keepalive(conn net.Conn) {
	msg := bgp.NewBGPKeepAliveMessage()
	data, _ := msg.Serialize()
	conn.Write(data)

	fmt.Println(" >>> sent KEEPALIVE: %#v", msg.Body)
}

func send_bgp_notification(conn net.Conn) {
	errorcode := uint8(1)
	errorsubcode := uint8(1)
	var errordata []byte
	msg := bgp.NewBGPNotificationMessage(errorcode, errorsubcode, errordata)

	data, _ := msg.Serialize()
	conn.Write(data)
	fmt.Println(" >>> sent NOTIFICATION: %#v", msg.Body)
}

func main() {
	// コマンドライン引数処理

	// configとシナリオの読み込み
	readConfig(*f)
	//	bgpinfo, scenario_queue := parse_scenario(config)
	//
	//	var conn net.Conn
	//	bgpMsg_ch := make(chan *bgp.BGPMessage)
	//	exit_ch := make(chan bool)
	//
	//	for key, val := range scenario_queue {
	//
	//		switch val.operation {
	//		case "connect":
	//			fmt.Println("[Scenario%d: %s]", key, val.operation)
	//			conn = connect(bgpinfo)
	//
	//			go receive_bgp_msg(conn, bgpMsg_ch, exit_ch)
	//		case "receive_bgp_open":
	//			fmt.Println("[Scenario%d: %s]", key, val.operation)
	//			receive_bgp_open(bgpMsg_ch)
	//		case "receive_bgp_update":
	//			fmt.Println("[Scenario%d: %s]", key, val.operation)
	//			receive_bgp_update(bgpMsg_ch)
	//		case "receive_bgp_keepalive":
	//			fmt.Println("[Scenario%d: %s]", key, val.operation)
	//			receive_bgp_keepalive(bgpMsg_ch)
	//		case "receive_bgp_notification":
	//			fmt.Println("[Scenario%d: %s]", key, val.operation)
	//			receive_bgp_notification(bgpMsg_ch)
	//		case "send_bgp_open":
	//			fmt.Println("[Scenario%d: %s]", key, val.operation)
	//			send_bgp_open(conn, bgpinfo, val.parameter)
	//		case "send_bgp_keepalive":
	//			fmt.Println("[Scenario%d: %s]", key, val.operation)
	//			send_bgp_keepalive(conn)
	//		case "send_bgp_notification":
	//			fmt.Println("[Scenario%d: %s]", key, val.operation)
	//			send_bgp_notification(conn)
	//		case "wait":
	//			fmt.Println("[Scenario%d: %s]", key, val.operation)
	//			wait(val.parameter)
	//		case "print_screen":
	//			fmt.Println("[Scenario%d: %s]", key, val.operation)
	//			print_screen(val.parameter)
	//		case "close":
	//			fmt.Println("[Scenario%d: %s]", key, val.operation)
	//			close(conn)
	//			//exit_ch <- true
	//		default:
	//			fmt.Println("Error: No Such Operation (%s). Exit Program.", val.operation)
	//			os.Exit(1)
	//		}
	//
	//		//time.Sleep(1 * time.Second)
	//
	//	}
	//	fmt.Println("")
	//	fmt.Println("<<< Success!! All Scenarios have been Passed!! >>>")
	//	fmt.Println("")
}
