package etcd

import "time"

const DialTimeout = 5 * time.Second

func GetEtcdEndpoints() []string {
	return []string{"127.0.0.1:2379"}
}