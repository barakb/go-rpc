package pool
import "com.githab.barakb/rpc"


type ConnectionPool struct {
	pool
}


func newConnection(address string) (interface{}, string, error){
	var ret interface{}
	ret, err := rpc.OpenConnection(address);
	if err != nil{
		return nil, "", err
	}
	remoteAddress := ret.(*rpc.Connection).RemoteAddress()
	return ret, remoteAddress, nil
}

func closeConnection(c interface{}){
	c.(*rpc.Connection).Close()
}


func NewConnectionPool(size int) *ConnectionPool{
	return &ConnectionPool{pool : *CreatePool(size, newConnection, closeConnection)}
}

func (p *ConnectionPool) Get(address string) (*rpc.Connection, error) {
	value, _, err := p.pool.Get(address)
	if err != nil{
		return nil, err
	}
	return value.(*rpc.Connection), nil;
}

func (p *ConnectionPool) Put(connection *rpc.Connection) {
	p.pool.Return(connection, connection.RemoteAddress(), connection.LocalAddress())
}

func (p *ConnectionPool) Close() {
}
