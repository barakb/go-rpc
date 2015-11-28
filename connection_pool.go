package rpc

import (
	"errors"
	"sync"
)

type ConnectionPool struct {
	size           int
	mutex          sync.RWMutex
	connectionsMap map[string]chan *Connection
	logger         Logger
	closed         bool
}

func NewConnectionPool(size int, logger Logger) *ConnectionPool {
	return &ConnectionPool{size: size, connectionsMap: make(map[string]chan *Connection), logger: logger}
}

func (p *ConnectionPool) Get(address string) (*Connection, error) {
	p.mutex.Lock()
	if p.closed {
		p.mutex.Unlock()
		return nil, errors.New("Pool is closed")
	}
	var connections chan *Connection
	if connections = p.connectionsMap[address]; connections == nil {
		connections = make(chan *Connection, p.size)
		p.connectionsMap[address] = connections
	}
	p.mutex.Unlock()

	select {
	case res := <-connections:
		return res, nil
	default:
		p.logger.Debug("Creating new connection [%s]", address)
		return OpenConnection(address)
	}
}

func (p *ConnectionPool) Put(connection *Connection) {
	p.mutex.RLock()
	if p.closed {
		p.mutex.RUnlock()
		p.logger.Debug("Disaposing live connection [%s], pool is closed", connection.localAddress)
		connection.Close()
		return
	}
	connections := p.connectionsMap[connection.localAddress]
	p.mutex.RUnlock()

	select {
	case connections <- connection:
		return
	default:
		p.logger.Debug("Disaposing live connection [%s]", connection.localAddress)
		connection.Close()
	}
}

func (p *ConnectionPool) Close() {
	p.mutex.Lock()
	p.closed = true
	p.mutex.Unlock()
	for key, connections := range p.connectionsMap {
		p.closeAll(key, connections)
	}
	p.connectionsMap = nil
}

func (p *ConnectionPool) closeAll(address string, connections chan *Connection) {
	close(connections)
	for {
		select {
		case connection := <-connections:
			if connection == nil {
				return
			}
			connection.Close()
		default:
			return
		}
	}
}
