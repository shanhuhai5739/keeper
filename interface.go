package keeper

import (
	"time"
	"os"
	"github.com/coreos/etcd/client"
)

type (
	Config struct {
		Endpoints  []string
		Username   string
		Password   string
		Leader_key string
		Node       string
		TTL        time.Duration
	}
	KeeperImpl struct {
		config Config
		cli    client.KeysAPI
		leader bool
	}
	Keeper interface {
		Start() error
		IsLeader() bool
	}
)

const (
	DefautlEtcd = "http://127.0.0.1:2379"
	DefaultTTL  = time.Minute
)

func NewKeeper(config *Config) (k Keeper, err error) {
	if config == nil {
		config = new(Config)
	}
	
	if len(config.Endpoints) <= 0 {
		config.Endpoints = []string{DefautlEtcd}
	}

	if len(config.Node) <= 0 {
		hostname, err := os.Hostname()
		if err != nil {
			panic(err)
		}
		config.Node = hostname
	}

	if config.TTL <= 0 {
		config.TTL = DefaultTTL
	}

	k = &KeeperImpl{
		config: *config,
		leader: false,
	}
	return
}

func (k *KeeperImpl) Start() (err error) {
	err = k.connect()
	if err != nil {
		return
	}

	go k.observe()
	//	go keeper.WatchNode()

	return
}

func (k *KeeperImpl) IsLeader() bool {
	return k.leader
}
