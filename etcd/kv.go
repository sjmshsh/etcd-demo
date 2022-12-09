package etcd

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
)

// KvDemo put/get/del
func KvDemo() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   GetEtcdEndpoints(),
		DialTimeout: DialTimeout,
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer cli.Close()

	putRes, err := cli.Put(context.Background(), "key1", "value1")
	if err != nil {
		log.Fatalln(err)
	}

	putRes, err = cli.Put(context.Background(), "key1", "value11", clientv3.WithPrevKV())
	if err != nil {
		log.Fatalln(err)
	}
	if putRes.PrevKv != nil {
		fmt.Println(putRes.PrevKv.Value)
	}

	getRes, err := cli.Get(context.Background(), "key1", clientv3.WithPrefix())
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(getRes)

	cli.Delete(context.Background(), "key", clientv3.WithPrefix())
}
