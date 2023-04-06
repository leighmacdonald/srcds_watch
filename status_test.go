package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseStatus(t *testing.T) {
	s, e := parseStatus(`hostname: Kittyland Server
version : 7961495/24 7961495 secure
udp/ip  : 1.2.33.44:27015
steamid : [G:1:411111] (8556839292111111)
account : not logged in  (No account specified)
map     : pl_upward at: 0 x, 0 y, 0 z
tags    : nocrits,nodmgspread,payload,uncletopia
sourcetv:  10.20.30.40:27015, delay 0.0s  (local: 10.20.30.40:27016)
players : 7 humans, 1 bots (33 max)
edicts  : 781 used of 2048 max
# userid name                uniqueid            connected ping loss state  adr
#      2 "Uncletopia | Seattle | 1 | All " BOT                       active
#    774 "Dred"              [U:1:102426391]     05:03       55    0 active 10.0.0.1:27005
#    775 "smiley"            [U:1:279850548]     04:53      120    0 active 10.0.0.2:27005
#    776 "Eve From Summertime Saga" [U:1:1121894230] 04:34   93    0 active 10.0.0.3:36973
#    753 "APPLEHACK FATMAGIC RELATIVE" [U:1:859279805] 37:10   87    0 active 10.0.0.4:27005
#    765 "Detrim"            [U:1:155803057]     19:22       80    0 active 10.0.0.5:27005
#    720 "viciousbeatmaker"  [U:1:126610924]      1:36:10    72    0 active 10.0.0.6:27005
#    684 "smeasly"           [U:1:68453084]       2:51:15    33    0 active 10.0.0.7:27005
`)
	require.NoError(t, e)
	require.Equal(t, 7, s.PlayersHumans)
	require.Equal(t, 1, s.PlayersBots)
	require.Equal(t, 781, s.Edicts)
	require.Equal(t, "pl_upward", s.Map)
	require.Equal(t, 33, s.PlayerLimit)
	require.Equal(t, []statusPlayer{
		{online: 303, ping: 55, loss: 0, address: "10.0.0.1:27005", port: 27005, ip: "10.0.0.1", steamID: 0x1100001061ae717},
		{online: 293, ping: 120, loss: 0, address: "10.0.0.2:27005", port: 27005, ip: "10.0.0.2", steamID: 0x110000110ae2e34},
		{online: 274, ping: 93, loss: 0, address: "10.0.0.3:36973", port: 36973, ip: "10.0.0.3", steamID: 0x110000142debf56},
		{online: 2230, ping: 87, loss: 0, address: "10.0.0.4:27005", port: 27005, ip: "10.0.0.4", steamID: 0x1100001333791bd},
		{online: 1162, ping: 80, loss: 0, address: "10.0.0.5:27005", port: 27005, ip: "10.0.0.5", steamID: 0x110000109495db1},
		{online: 5770, ping: 72, loss: 0, address: "10.0.0.6:27005", port: 27005, ip: "10.0.0.6", steamID: 0x1100001078bedec},
		{online: 10275, ping: 33, loss: 0, address: "10.0.0.7:27005", port: 27005, ip: "10.0.0.7", steamID: 0x1100001041482dc},
	}, s.Players)

}
