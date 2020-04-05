# bgpdbully

bgpdbully is a test tool for BGP.<br>
It receives and sends BGP packets according to YAML syntax scenario, even if they are malformed.

## Install

```
$ go get github.com/nstgt/bgpdbully/cmd/bgpdbully
```

## Execution

Write config file and run with it.

```
$ bgpdbully -f sample.yaml
```

## Configuration

Configuration file is devided into 2 parts: `global` and `scenario`. See [example](sample.yaml).

In the `global` part, you neet to specify BGP connecting information.

```
global:
  peer_ip:   "127.0.0.1"  # IP address of remote BGP speaker in string
  peer_port: 179          # TCP port number of remoe BGP speaker in decimal number
  holdtime:  90           # hold time value in decimal number
  local_as:  65001        # local AS number in decimal number
  local_id:  "10.0.0.1"   # local router ID in string
```

In the `scenario` part, you can decide the BGP behavior of bgpdbully by using instructions composed of operation (`ope`) and its parameter (`param`, if needed).

```
scenario:
- ope: 'Operation'
  param: 'Parameter'
```

Here is a list for available instructions:

| Operation (`ope`) | Parameter (`param`) | Description |
| :--- | :--- | :--- |
| connect | - | connect to remote BGP speaker according to `global` configuration |
| sleep | needed | sleep given duration |
| close | - | close connection |
| send_bgp_open | needed | send BGP OPEN Message to the remote BGP speaker |
| send_bgp_update | needed | send BGP UPDATE Message to the remote BGP speaker |
| send_bgp_notification | needed | send BGP NOTIFICAYION Message to the remote BGP speaker |
| send_bgp_keepalive | - | send BGP KEEPALIVE Message to the remote BGP speaker |
| send_bgp_routerefresh | needed | send BGP ROUTEREFRESH Message to the remote BGP speaker |
| receive_bgp_open | - | wait for BGP OPEN Message from the remote BGP speaker |
| receive_bgp_update | - | wait for BGP UPDATE Message from the remote BGP speaker |
| receive_bgp_notification | - | wait for BGP NOTIFICATION Message from the remote BGP speaker |
| receive_bgp_keepalive | - | wait for BGP KEEPALIVE Message from the remote BGP speaker |
| receive_bgp_routerefresh | - | wait for BGP ROUTEREFRESH Message from the remote BGP speaker |

### Parameter for

#### `sleep`

```
param:
  sec: 10 #secounds in decimal number
```

#### `send_bgp_open`

```
param:
  capabilities:
  - type: 1            # capability type code in decimal number
    value: "00010001"  # capability data in hexadecimal string
```

Refer [IANA: Capability Codes](https://www.iana.org/assignments/capability-codes/capability-codes.xhtml) for `capabilities[*].type`.<br>
No human writable interface for `capabilities[*].value` yet :(

#### `send_bgp_update`

```
param:
  nlri:             # array of advertised IP prefixes in string
  - "10.0.0.0/24"
  withdrawn_routes: # array of withdrawn IP prefixes in string
  - "10.0.0.0/24"
  path_attributes:  # array of path attribution like follows
  - flag: "40"      # path attribution flag in hexadecimal string
    type: 1         # path attribution type code in decimal number
    value: "00"     # path attribution value in hexadecimal string
```
Refer [IANA: BGP Path Attributes](https://www.iana.org/assignments/bgp-parameters/bgp-parameters.xhtml#bgp-parameters-2) for `.path_attributes[*].type`.<br>
No human writable interface for `path_attributes[*].value` yet :(

#### `send_bgp_notification`

```
param:
  code: 1     # error code in decimal number
  subcode: 1  # error subcode in decimal number
```

Refer [IANA: BGP Error (Notification) Codes](https://www.iana.org/assignments/bgp-parameters/bgp-parameters.xhtml#bgp-parameters-3) for `code` and [IANA: BGP Error Subcodes](https://www.iana.org/assignments/bgp-parameters/bgp-parameters.xhtml#bgp-parameters-4) for `subcode`.


#### `send_bgp_routerefresh`

```
param:
  afi: 1   # address family indicator in decimal number
  safi: 1  # subsequent address family indicator in decimal number
```
Refer [IANA: Address Family Numbers](https://www.iana.org/assignments/address-family-numbers/address-family-numbers.xhtml) for `afi` and [IANA: Subsequent Address Family Identifiers (SAFI) Parameters
](https://www.iana.org/assignments/safi-namespace/safi-namespace.xhtml) for `safi`.

### `receive_bgp_*`
Currently, `receive_bgp_*` operations just validate BGP Message type of receiving BGP packet from the remote BGP speaker.

## Play

Let's bully [GoBGP](https://github.com/osrg/gobgp).<br>
You need to install GoBGP. If you don't have, see [here](https://github.com/osrg/gobgp#install).

Write config and start GoBGP.
```
$ cat > gobgpd.conf <<EOF
[global.config]
as = 64512
router-id = "10.0.0.100"

[[neighbors]]
  [neighbors.config]
      peer-as = 65001
      auth-password = "password"
      neighbor-address = "127.0.0.1"
      local-as = 64512
  [neighbors.transport.config]
      passive-mode = true
EOF

$ gobgpd -f gobgpd.conf
{"level":"info","msg":"gobgpd started","time":"2020-04-05T22:29:05+09:00"}
...
```

Run bgpdbully using [sample config](sample.yaml).

```
$ bgpdbully -f sample.yaml
2020/04/05 22:24:00 start
2020/04/05 22:24:00 connecting to 127.0.0.1:179
2020/04/05 22:24:00 send BGP Open Message
2020/04/05 22:24:00 receive BGP Message, type 1
2020/04/05 22:24:00 send BGP Keepalive Message
2020/04/05 22:24:00 receive BGP Message, type 4
2020/04/05 22:24:00 send BGP Update Message
2020/04/05 22:24:00 sleep 10 sec
2020/04/05 22:24:10 send BGP Update Message
2020/04/05 22:24:10 sleep 10 sec
2020/04/05 22:24:20 send BGP RouteRefresh Message
2020/04/05 22:24:20 sleep 10 sec
2020/04/05 22:24:30 send BGP Notification Message
2020/04/05 22:24:30 closing connection to 127.0.0.1:179
```

You'll see GoBGP establish BGP peer with bgpdbully, and receive routes from it.

```
$ gobgp neighbor
Peer         AS  Up/Down State       |#Received  Accepted
127.0.0.1 65001 00:00:02 Establ      |        2         2

$ gobgp neighbor 127.0.0.1 adj-in
ID  Network              Next Hop             AS_PATH              Age        Attrs
0   10.0.0.0/24          10.0.0.1             65001                00:00:05   [{Origin: i}]
0   10.0.1.0/24          10.0.0.1             65001                00:00:05   [{Origin: i}]

## '10.0.0.0/24' was withdrawn
$  gobgp neighbor 127.0.0.1 adj-in
ID  Network              Next Hop             AS_PATH              Age        Attrs
0   10.0.1.0/24          10.0.0.1             65001                00:00:11   [{Origin: i}]

$ gobgp neighbor
Peer         AS  Up/Down State       |#Received  Accepted
127.0.0.1 65001 00:00:03 Idle        |        0         0
```