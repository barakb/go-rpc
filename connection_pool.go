package rpc
import (
	"sync"
	"errors"
)

type ConnectionPool struct {
	size           int
	mutex          sync.RWMutex
	connectionsMap map[string]chan *Connection
	logger         Logger
	closed         bool
}

func NewConnectionPool(size int, logger Logger) *ConnectionPool {
	return &ConnectionPool{size:size, connectionsMap:make(map[string]chan *Connection), logger:logger}
}

func (p *ConnectionPool) Get(address string) (*Connection, error) {
	p.mutex.Lock();
	if p.closed {
		p.mutex.Unlock()
		return nil, errors.New("Pool is closed")
	}
	connections := p.connectionsMap[address]
	if connections == nil {
		if connections = p.connectionsMap[address]; connections == nil {
			connections = make(chan *Connection, p.size)
			p.connectionsMap[address] = connections
		}
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
		p.mutex.Unlock()
		p.logger.Debug("Disaposing live connection [%s]", connection.address)
		connection.Close()
		return
	}
	connections := p.connectionsMap[connection.address]
	p.mutex.RUnlock()

	select {
	case connections <- connection:
		return
	default:
		p.logger.Debug("Disaposing live connection [%s]", connection.address)
		connection.Close()
	}
}

func (p *ConnectionPool) Close() {
	p.mutex.Lock();
	p.closed = true
	p.mutex.Unlock()
	for _, connections := range p.connectionsMap{
		close(connections)
		select {
		case connection := <-connections:
			connection.Close()
		default:
		}
	}
	p.connectionsMap = make(map[string]chan *Connection)

}



