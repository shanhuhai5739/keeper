package main

import (
	"flag"
	"log"
	"time"

	"net/http"

	"keeper"
	"xlog"
)

var (
	endpoint   = flag.String("endpoint", "http://127.0.0.1:2379", "etcd endpoint")
	username   = flag.String("username", "root", "username")
	password   = flag.String("password", "root", "password")
	leader_key = flag.String("leader_key", "/guard/leader", "leader_key")
	id         = flag.String("id", "host1", "id")
	httpServer = flag.String("http", ":9801", "httpServer")

	logger   = xlog.NewXLog("scanner")
	isleader = false
)

func main() {
	config := &keeper.Config{
		Endpoints:  []string{*endpoint},
		Username:   *username,
		Password:   *password,
		Leader_key: *leader_key,
		Node:       *id,
		TTL:        10 * time.Second,
	}

	go httpHandle()
	keeper, err := keeper.New(config)
	if err != nil {
		log.Fatal(err)
	}

	for {
		leader := keeper.IsLeader()
		if leader {
			//log.Println("===", k)
			isleader = true
		} else {
			isleader = false
		}
		time.Sleep(1 * time.Second)
	}
}

func httpHandle() {
	http.HandleFunc("/leader", func(w http.ResponseWriter, r *http.Request) {
		if isleader {
			w.Write([]byte("true"))
		} else {
			w.Write([]byte("false"))
		}
	})

	err := http.ListenAndServe(*httpServer, nil)
	if err != nil {
		panic(err)
	}
	return
}

func init() {
	flag.Parse()

	logger.Open("./logs", "app.log", "debug")
}
