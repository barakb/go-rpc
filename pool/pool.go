package pool
import "sync"


type CreateResource func(name string) (interface{}, string, error)
type FreeResource func(resource interface{})

type resourceCollection struct {
	name       string
	resources map[string]interface{}
}

func newResourceCollection(name string) *resourceCollection{
	return &resourceCollection{name, make(map[string]interface{})}
}

func (r *resourceCollection) add(value interface{}, key string){
	r.resources[key] = value
}

func (r *resourceCollection) get() (value interface{}, key string, found bool){
	for key, value = range r.resources{
		found = true;
		break;
	}
	if found{
		delete(r.resources, key)
	}
	return value, key, found
}

type pool struct {
	createResource CreateResource
	freeResource   FreeResource
	mutex          sync.Mutex
	resources      map[string]*resourceCollection
	maxSize        int
	len            int
}

func CreatePool(maxSize int, createResource CreateResource, freeResource FreeResource) *pool {
	res := &pool{}
	res.maxSize = maxSize
	res.createResource = createResource
	res.freeResource = freeResource
	res.resources = make(map[string]*resourceCollection)
	return res
}

func (p *pool) Get(name string) (interface{}, string, error) {
	p.mutex.Lock()
	rc, ok := p.resources[name]
	if !ok {
		p.mutex.Unlock()
		return p.createResource(name)
	}
	val, key, ok := rc.get()
	if !ok{
		p.mutex.Unlock()
		return p.createResource(name)
	}
	p.len -= 1
	p.mutex.Unlock()
	return val, key, nil
}

func (p *pool) Len() int{
	return p.len
}

func (p *pool) Return(value interface{}, name string, key string) int {
	p.mutex.Lock()
	if(p.len < p.maxSize) {
		rc, ok := p.resources[name]
		if !ok {
			rc = newResourceCollection(name)
			p.resources[name] = rc
		}
		rc.add(value, key)
		p.len += 1
		p.mutex.Unlock()
		return p.len
	}else{
		p.mutex.Unlock()
		p.freeResource(value)
		return p.len
	}
}

