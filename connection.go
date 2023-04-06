package main

import (
	"context"
	"github.com/leighmacdonald/rcon/rcon"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type connection struct {
	address  string
	password string
	rcon     *rcon.RemoteConsole
	rconMu   *sync.RWMutex
}

func newConnection(target Target) *connection {
	return &connection{
		address:  target.Addr(),
		password: target.Password,
		rconMu:   &sync.RWMutex{},
	}
}

func (c *connection) Connect(ctx context.Context) error {
	if c.rcon == nil {
		conn, errConn := rcon.Dial(ctx, c.address, c.password, time.Second*10)
		if errConn != nil {
			return errors.Wrap(errConn, "Failed to connect to Host")
		}
		c.rconMu.Lock()
		c.rcon = conn
		c.rconMu.Unlock()
	}
	return nil
}

func (c *connection) Close() error {
	return nil
}

func (c *connection) Stats() (*stats, error) {
	body, errExec := c.rcon.Exec("stats")
	if errExec != nil {
		return nil, errors.Wrap(errExec, "Failed to execute rcon stats command")
	}
	return parseStats(body)
}

type connManager struct {
	sync.RWMutex
	connections map[string]*connection
}

func newConnManager() *connManager {
	return &connManager{
		connections: map[string]*connection{},
	}
}

func (cm *connManager) get(target Target) (*connection, error) {
	cm.RLock()
	conn, found := cm.connections[target.Name]
	if found {
		cm.RUnlock()
		return conn, nil
	}
	cm.RUnlock()
	newConn := newConnection(target)
	cm.Lock()
	cm.connections[target.Name] = newConn
	cm.Unlock()
	return newConn, nil
}
