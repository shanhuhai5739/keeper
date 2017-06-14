package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

var (
	endpoint = flag.String("endpoint", "http://127.0.0.1:2379", "etcd endpoint")
	leader   = flag.String("leader", "host1", "id")

	leader_key = "/guard/leader"
	cli        client.KeysAPI
)

func main() {
	err := Connect()
	if err != nil {
		panic(err)
	}

	for {
		resp, err := cli.Get(context.Background(), leader_key, &client.GetOptions{
			Quorum: true,
		})
		if err != nil {
			etcdErr, ok := err.(client.Error)
			if !ok {
				log.Println(etcdErr)
				continue
			}
			if etcdErr.Code == client.ErrorCodeKeyNotFound {
				continue
			}
			panic(err)
		}

		val := resp.Node.Value
		if val != *leader {
			err = fmt.Errorf("%v != leader", val)
			panic(err)
		}

		time.Sleep(time.Second)
	}
}

func Connect() (err error) {
	cfg := client.Config{
		Endpoints:               []string{*endpoint},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		return
	}

	kapi := client.NewKeysAPI(c)
	cli = kapi
	return
}

func init() {
	flag.Parse()
}
