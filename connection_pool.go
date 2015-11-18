package rpc
import "sync"

type ConnectionPool struct {
	size           int
	mutex          sync.RWMutex
	connectionsMap map[string]chan *Connection
	logger Logger
}

func NewConnectionPool(size int, logger Logger) *ConnectionPool {
	return &ConnectionPool{size:size, connectionsMap:make(map[string]chan *Connection), logger:logger}
}

func (p *ConnectionPool) Get(address string) (*Connection, error) {
	connections := p.connectionsMap[address]
	if connections == nil {
		p.mutex.Lock();
		if connections = p.connectionsMap[address]; connections == nil {
			connections = make(chan *Connection, p.size)
			p.connectionsMap[address] = connections
		}
		p.mutex.Unlock()
	}
	select {
	case res := <-connections:
		return res, nil
	default:
		return OpenConnection(address)
	}
}

func (p *ConnectionPool) Put(connection *Connection) {
	p.mutex.RLock()
	connections := p.connectionsMap[connection.address]
	p.mutex.RUnlock()

	select {
	case connections <- connection:
		return
	default:
		connection.Close()
	}
}



