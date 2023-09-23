package main

import (
	"context"
	"sync"
	"time"

	"github.com/leighmacdonald/rcon/rcon"
	"github.com/pkg/errors"
)

type connection struct {
	address  string
	password string
	rcon     *rcon.RemoteConsole
	rconMu   *sync.RWMutex
	parser   statusParser
}

func newConnection(target Target) *connection {
	return &connection{
		address:  target.addr(),
		password: target.Password,
		rconMu:   &sync.RWMutex{},
		parser:   newStatusParser(),
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
	if err := c.rcon.Close(); err != nil {
		return errors.Wrap(err, "Failed to close conn")
	}

	return nil
}

func (c *connection) Status() (*status, error) {
	c.rconMu.RLock()
	defer c.rconMu.RUnlock()

	body, errExec := c.rcon.Exec("status;stats;sv_maxupdaterate;sm version;meta version;sv_visiblemaxplayers")
	if errExec != nil {
		return nil, errors.Wrap(errExec, "Failed to execute rcon status command")
	}

	return c.parser.parse(body)
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
