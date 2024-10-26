package etcdRegister

import (
	"context"
	"encoding/json"
	"errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

const (
	MicroServicePrefix = "/microService/"
	LeaseTime          = 15 //租约时间15second
)

type EtcdRegister struct {
	client *clientv3.Client
	logger *zap.Logger

	//下面的字段是内部需要的
	lease     clientv3.Lease
	keepAlive <-chan *clientv3.LeaseKeepAliveResponse
	leaseId   clientv3.LeaseID
}

func NewEtcdRegister(logger *zap.Logger, client *clientv3.Client) *EtcdRegister {
	return &EtcdRegister{logger: logger, client: client}
}

// 创建租约
func (receiver *EtcdRegister) createLease() error {
	//创建lease对象
	receiver.lease = clientv3.NewLease(receiver.client)

	//分配租约
	ctx, _ := context.WithCancel(context.Background())
	GrantRes, err := receiver.lease.Grant(ctx, LeaseTime)
	if err != nil {
		return err
	}

	receiver.leaseId = GrantRes.ID

	//保持续租,获得一个chan
	//它这个和linux的客户端不一样,那个客户端会一直续租
	//而在go里,不会一直续租,如果一直没有人取出来keepAliveChain,则不会再续租
	//所以我们的程序如果关闭了,这个keepAliveChain就没人取了,etcd就不会再续租了
	receiver.keepAlive, err = receiver.lease.KeepAlive(ctx, receiver.leaseId)
	if err != nil {
		return err
	}

	return nil
}

// 持续续租
func (receiver *EtcdRegister) listenLease(leaseChan <-chan *clientv3.LeaseKeepAliveResponse) {
	for {
		select {
		//leaseChan会阻塞，并且15s的ttl，那么就是5s续租一次
		case <-leaseChan:
		}
	}
}

// 判断对应的服务名称里是否存在对应的ip+port,如果存在返回对应的key
func (receiver *EtcdRegister) isIpPortExist(info *ServerInfo) (bool, error) {
	ctx := context.Background()

	res, err := receiver.client.Get(ctx, MicroServicePrefix+info.ServerName, clientv3.WithPrefix())
	if res.Count == 0 {
		return false, err
	}

	isExist := false

	for _, kv := range res.Kvs {
		var resInfo ServerInfo
		err := json.Unmarshal(kv.Value, &resInfo)
		if err != nil {
			receiver.logger.Error("从etcd解析到ServerInfo失败", zap.Error(err))
			return false, err
		}

		//判定是不是空，如果是空，那就代表肯定不存在
		//只有不等于空的情况下才去判断，只有有一个存在，那就代表存在，我们肯定不能注册这个端口
		if resInfo.GrpcServer != "" {
			if resInfo.GrpcServer == info.GrpcServer {
				isExist = true
			}
		}
		if resInfo.HttpServer != "" {
			if resInfo.HttpServer == info.HttpServer {
				isExist = true
			}
		}

	}
	return isExist, nil
}

func (receiver *EtcdRegister) DeleteKey(key string) error {
	_, err := receiver.client.Delete(context.Background(), key)
	if err != nil {
		return err
	}
	return nil
}

// 设置key,以及key的租约时间
func (receiver *EtcdRegister) RegisterServer(ctx context.Context, info *ServerInfo) error {
	//创建一个租约
	err := receiver.createLease()
	if err != nil {
		return err
	}

	//给租约续期
	go receiver.listenLease(receiver.keepAlive)

	//查看服务里是不是有相同的port+key了，如果有，那就提示错误
	exist, err := receiver.isIpPortExist(info)
	if err != nil {
		return err
	}
	if exist {
		return errors.New(info.GrpcPort + "已存在")
	}

	//如果注册中心里没有我们的地址,则注册进去
	//每次存储的时候,然后生产一个唯一的UUID,然后把值存进去
	allKey := MicroServicePrefix + info.ServerName + "/" + info.Id
	marshalStr, _ := json.Marshal(info)
	_, err = receiver.client.Put(ctx, allKey, string(marshalStr), clientv3.WithLease(receiver.leaseId))
	if err != nil {
		receiver.logger.Error("注册服务失败", zap.Error(err))
		return err
	}
	return nil
}
