package mycache

import (
	"fmt"
	"github.com/GoodOneGuy/mycache/util"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) (interface{}, error)
}

type GetterFunc func(key string) (interface{}, error)

func (f GetterFunc) Get(key string) (interface{}, error) {
	return f(key)
}

type Group struct {
	name      string
	getter    Getter
	mainCache util.Cache
	peers     util.PeerPicker
	loader    *util.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func (g *Group) RegisterPeers(peers util.PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func NewGroup(name string, maxNum int, getter Getter) *Group {
	if getter == nil {
		panic("nil getter")
	}

	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: util.NewMutexLRUCache(maxNum),
		loader:    &util.Group{},
	}

	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return groups[name]
}

func (g *Group) Get(key string) (interface{}, error) {
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	if obj := g.mainCache.Find(key); obj != nil {
		log.Println("[GeeCache] hit, key =", key, "value=", obj)
		return obj, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (interface{}, error) {

	obj, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err := g.getFromPeer(peer, key); err != nil {
					return value, nil
				}
			}
		}
		return g.getLocal(key)
	})

	if err != nil {
		return obj, nil
	}

	return nil, err

}

func (g *Group) getLocal(key string) (interface{}, error) {
	obj, err := g.getter.Get(key)
	if err != nil {
		return nil, err
	}

	g.addCache(key, obj)
	return obj, nil
}

func (g *Group) getFromPeer(peer util.PeerGetter, key string) (interface{}, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (g *Group) addCache(key string, value interface{}) {
	g.mainCache.Insert(key, value)
}
