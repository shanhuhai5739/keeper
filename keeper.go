package keeper

import (
	"time"

	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"xlog"
)

func (k *KeeperImpl) connect() (err error) {
	cfg := client.Config{
		Username:                k.config.Username,
		Password:                k.config.Password,
		Endpoints:               k.config.Endpoints,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		return
	}

	kapi := client.NewKeysAPI(c)
	k.cli = kapi
	return
}

func (k *KeeperImpl) getNode(key string) (resp *client.Response, err error) {
	resp, err = k.cli.Get(context.Background(), key, &client.GetOptions{
		Quorum: true,
	})
	if err != nil {
		return
	}
	return
}

func (k *KeeperImpl) delNode(key, value string) (err error) {
	_, err = k.cli.Delete(context.Background(), key, &client.DeleteOptions{
		PrevValue: value,
	})
	if err != nil {
		return
	}
	return
}

// observe node
func (k *KeeperImpl) observe() {
	for {
		k.election()
		if k.leader {
			// master renew ttl
			time.Sleep(k.config.TTL / 2)
		} else {
			// wait a moment
			time.Sleep(2 * time.Second)
		}
	}
}

// CAS
func (k *KeeperImpl) election() {
	leader := false
	defer func() {
		k.leader = leader
	}()

	_, err := k.getNode(k.config.Leader_key)
	if err != nil {
		etcdErr, ok := err.(client.Error)
		if !ok {
			xlog.Warn("etcd: unexpceted err: %v", err)
			return
		}
		if etcdErr.Code == client.ErrorCodeKeyNotFound {
			// key doesn't exist
			opts := &client.SetOptions{
				Dir:       false,
				PrevExist: client.PrevNoExist,
				TTL:       k.config.TTL,
			}
			_, err = k.cli.Set(context.Background(), k.config.Leader_key, k.config.Node, opts)
			if err != nil {
				// set node failed, means other node success, we become slave
				xlog.Warn("Keeper: set node is err: %v\n", err)
				return
			}

			// set node success, we become master
			leader = true
			xlog.Notice("Keeper: I'm leader now [%v]", k.config.Node)
			return
		}
		xlog.Warn("etcd: unexpceted err: %v", err)
		return
	}
	// renew ttl for leader node
	opts := &client.SetOptions{
		Dir:       false,
		PrevValue: k.config.Node, // check value
		PrevExist: client.PrevExist,
		TTL:       k.config.TTL,
	}
	_, err = k.cli.Set(context.Background(), k.config.Leader_key, k.config.Node, opts)
	if err != nil {
		etcdErr, ok := err.(client.Error)
		if !ok {
			xlog.Warn("etcd: unexpceted err: %v", err)
			return
		}
		if etcdErr.Code == client.ErrorCodeTestFailed {
			xlog.Trace("Keeper: set node is err: %v\n", err)
			return
		}
		xlog.Warn("Keeper: set node is err: %v\n", err)
		return
	}

	leader = true
	xlog.Notice("Keeper: I'm leader [%v]", k.config.Node)
	return
}
