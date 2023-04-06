package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseStats(t *testing.T) {
	s, e := parseStats(`CPU    In_(KB/s)  Out_(KB/s)  Uptime  Map_changes  FPS      Players  Connects
0.10   4.84       5.22        103     1            66.67    1        2`)
	require.NoError(t, e)
	require.Equal(t, s.CPU, 0.10)
	require.Equal(t, s.NetIn, 4.84)
	require.Equal(t, s.NetOut, 5.22)
	require.Equal(t, s.Uptime, 103)
	require.Equal(t, s.Maps, 1)
	require.Equal(t, s.FPS, 66.67)
	require.Equal(t, s.Player, 1)
	require.Equal(t, s.Connects, 2)

}
