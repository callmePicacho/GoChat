package etcd

import (
	"GoChat/config"
	"context"
	"fmt"
	clientV3 "go.etcd.io/etcd/client/v3"
	"time"
)

// Register 服务注册
type Register struct {
	client        *clientV3.Client                        // etcd client
	leaseID       clientV3.LeaseID                        //租约ID
	keepAliveChan <-chan *clientV3.LeaseKeepAliveResponse // 租约 KeepAlive 相应chan
	key           string                                  // key
	val           string                                  // value
}

// RegisterServer 新建注册服务
func RegisterServer(key string, value string, lease int64) error {
	client, err := clientV3.New(clientV3.Config{
		Endpoints:   config.GlobalConfig.ETCD.Endpoints,
		DialTimeout: time.Duration(config.GlobalConfig.ETCD.Timeout) * time.Second,
	})
	if err != nil {
		fmt.Println("etcd err:", err)
		return err
	}

	ser := &Register{
		client: client,
		key:    key,
		val:    value,
	}

	//申请租约设置时间keepalive
	if err = ser.putKeyWithLease(lease); err != nil {
		return err
	}

	//监听续租相应chan
	go ser.ListenLeaseRespChan()

	return nil
}

// putKeyWithLease 设置 key 和租约
func (r *Register) putKeyWithLease(timeNum int64) error {
	//设置租约时间
	resp, err := r.client.Grant(context.TODO(), timeNum)
	if err != nil {
		return err
	}
	//注册服务并绑定租约
	_, err = r.client.Put(context.TODO(), r.key, r.val, clientV3.WithLease(resp.ID))
	if err != nil {
		return err
	}

	//设置续租 定期发送需求请求
	leaseRespChan, err := r.client.KeepAlive(context.TODO(), resp.ID)
	if err != nil {
		return err
	}

	r.leaseID = resp.ID
	r.keepAliveChan = leaseRespChan
	return nil
}

// ListenLeaseRespChan 监听 续租情况
func (r *Register) ListenLeaseRespChan() {
	defer r.close()

	for leaseKeepResp := range r.keepAliveChan {
		fmt.Printf("续租成功，leaseID:%d, Put key:%s,val:%s reps:+%v\n", r.leaseID, r.key, r.val, leaseKeepResp)
	}
	fmt.Printf("续约失败，leaseID:%d, Put key:%s,val:%s\n", r.leaseID, r.key, r.val)

}

// Close 撤销租约
func (r *Register) close() error {
	//撤销租约
	if _, err := r.client.Revoke(context.Background(), r.leaseID); err != nil {
		return err
	}
	fmt.Printf("撤销租约成功, leaseID:%d, Put key:%s,val:%s\n", r.leaseID, r.key, r.val)
	return r.client.Close()
}
