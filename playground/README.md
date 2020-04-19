# Playground
Running bgpdbully with GoBGP in docker-compose environment.

## Play

```
## in this directory
$ docker-compose up -d

$ docker exec gobgp gobgp neighbor
Peer         AS  Up/Down State       |#Received  Accepted
10.0.1.20 65001 00:00:02 Establ      |        2         2

$ docker logs bgpdbully
2020/04/19 05:33:11 start
2020/04/19 05:33:11 sleep 3 sec
2020/04/19 05:33:14 connecting to 10.0.1.10:179
2020/04/19 05:33:14 send BGP Open Message
2020/04/19 05:33:14 receive BGP Message, type 1
2020/04/19 05:33:14 send BGP Keepalive Message
2020/04/19 05:33:14 receive BGP Message, type 4
2020/04/19 05:33:14 send BGP Update Message
2020/04/19 05:33:14 sleep 10 sec
2020/04/19 05:33:24 send BGP Update Message
2020/04/19 05:33:24 sleep 10 sec
2020/04/19 05:33:34 send BGP RouteRefresh Message
2020/04/19 05:33:34 sleep 10 sec
2020/04/19 05:33:44 send BGP Notification Message
2020/04/19 05:33:44 closing connection to 10.0.1.10:179
```

## Clean up

```
$ docker-compose down
```