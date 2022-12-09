package discovery

import (
	"context"
	"etcd/etcd"
	"fmt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"sync"
)

type Service struct {
	Name     string
	IP       string
	Port     string
	Protocol string // 协议
}

func ServiceRegister(s *Service) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcd.GetEtcdEndpoints(),
		DialTimeout: etcd.DialTimeout,
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer cli.Close()

	var grantLease bool
	var leaseID clientv3.LeaseID
	ctx := context.Background()
	// 判断key是否存在
	getRes, err := cli.Get(ctx, s.Name, clientv3.WithCountOnly())
	if err != nil {
		log.Fatalln(err)
	}
	if getRes.Count == 0 {
		grantLease = true
	}

	// 租约声明
	if grantLease {
		leaseRes, err := cli.Grant(ctx, 10)
		if err != nil {
			log.Fatalln(err)
		}
		leaseID = leaseRes.ID
	}

	kv := clientv3.NewKV(cli)
	txn := kv.Txn(ctx)
	_, err = txn.If(clientv3.Compare(clientv3.CreateRevision(s.Name), "=", 0)).
		Then(
			clientv3.OpPut(s.Name, s.Name, clientv3.WithLease(leaseID)),
			clientv3.OpPut(s.Name+".ip", s.Name, clientv3.WithLease(leaseID)),
			clientv3.OpPut(s.Name+".port", s.Name, clientv3.WithLease(leaseID)),
			clientv3.OpPut(s.Name+".protocol", s.Name, clientv3.WithLease(leaseID)),
		).
		Else(
			clientv3.OpPut(s.Name, s.Name, clientv3.WithIgnoreLease()),
			clientv3.OpPut(s.Name+".ip", s.Name, clientv3.WithIgnoreLease()),
			clientv3.OpPut(s.Name+".port", s.Name, clientv3.WithIgnoreLease()),
			clientv3.OpPut(s.Name+".protocol", s.Name, clientv3.WithIgnoreLease()),
		).
		Commit()
	if err != nil {
		log.Fatalln(err)
	}

	if grantLease {
		leaseKeepalive, err := cli.KeepAlive(ctx, leaseID)
		if err != nil {
			log.Fatalln(err)
		}
		for lease := range leaseKeepalive {
			fmt.Printf("leaseID:%x, ttl:%d\n", lease.ID, lease.TTL)
		}
	}
	// 租约
	// 事务
	// 数据操作

}

type Services struct {
	services map[string]*Service
	sync.RWMutex
}

var myService = &Services{
	services: map[string]*Service{
		"hello.Greeter": &Service{},
	},
}

func ServerDiscovery(svcName string) *Service {
	var s *Service = nil
	myService.RLock()
	s, _ = myService.services[svcName]
	myService.RUnlock()
	return s
}

func WatchServiceName(svcName string) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcd.GetEtcdEndpoints(),
		DialTimeout: etcd.DialTimeout,
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer cli.Close()

	getRes, err := cli.Get(context.Background(), svcName, clientv3.WithPrefix())
	if err != nil {
		log.Fatalln(err)
	}
	if getRes.Count > 0 {
		mp := sliceToMap(getRes.Kvs)
		s := &Service{}
		if kv, ok := mp[svcName]; ok {
			s.Name = string(kv.Value)
		}
		if kv, ok := mp[svcName+".ip"]; ok {
			s.Name = string(kv.Value)
		}
		if kv, ok := mp[svcName+".port"]; ok {
			s.Name = string(kv.Value)
		}
		if kv, ok := mp[svcName+".protocol"]; ok {
			s.Name = string(kv.Value)
		}
		myService.Lock()
		myService.services[svcName] = s
		myService.Unlock()
	}

	rch := cli.Watch(context.Background(), svcName, clientv3.WithPrefix())
	for wres := range rch {
		for _, ev := range wres.Events {
			if ev.Type == clientv3.EventTypeDelete {
				myService.Lock()
				delete(myService.services, svcName)
				myService.Unlock()
			}
			if ev.Type == clientv3.EventTypePut {
				myService.Lock()
				switch string(ev.Kv.Key) {
				case svcName:
					myService.services[svcName].Name = string(ev.Kv.Value)

				case svcName + ".ip":
					myService.services[svcName].Name = string(ev.Kv.Value)
				case svcName + ".port":
					myService.services[svcName].Name = string(ev.Kv.Value)

				case svcName + ".protocol":
					myService.services[svcName].Name = string(ev.Kv.Value)
				}
				myService.Unlock()
			}
		}
	}
}

func sliceToMap(list []*mvccpb.KeyValue) map[string]*mvccpb.KeyValue {
	mp := make(map[string]*mvccpb.KeyValue, 0)
	for _, item := range list {
		mp[string(item.Key)] = item
	}
	return mp
}
