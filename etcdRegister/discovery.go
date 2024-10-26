package etcdRegister

import (
	"context"
	"encoding/json"
	"github.com/Golong-me/microServiceKit/etcdRegister/loadBalance"
	"github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

// 服务发现，需要
type DiscoveryService struct {
	client      *clientv3.Client
	logger      *zap.Logger
	LoadBalance loadBalance.LoadBalance
	ServerName  string
	ServerInfos []*ServerInfo
}

func NewDiscoveryService(client *clientv3.Client, logger *zap.Logger, serverName string, loadBalance loadBalance.LoadBalance) *DiscoveryService {
	return &DiscoveryService{client: client, logger: logger, LoadBalance: loadBalance, ServerName: serverName}
}

// 获取服务地址信息
func (receiver *DiscoveryService) GetAllServiceInfo() error {
	res, err := receiver.client.Get(context.Background(), MicroServicePrefix+receiver.ServerName, clientv3.WithPrefix())
	if err != nil {
		return errors.WithMessage(err, "获取远程服务出错")
	}
	for _, kv := range res.Kvs {
		var info ServerInfo
		err := json.Unmarshal(kv.Value, &info)
		if err != nil {
			return errors.WithMessage(err, "从etcd解析到ServerInfo失败")
		}
		receiver.ServerInfos = append(receiver.ServerInfos, &info)
	}
	return nil
}

func (receiver *DiscoveryService) GetGrpcIpPort() (string, error) {
	err := receiver.GetAllServiceInfo()
	if err != nil {
		return "", err
	}

	var grpcIpPorts []string
	for _, info := range receiver.ServerInfos {
		grpcIpPorts = append(grpcIpPorts, info.GrpcServer)
	}

	return receiver.LoadBalance.SelectOne(grpcIpPorts), nil
}

func (receiver *DiscoveryService) GetHttpIpPort() (string, error) {
	err := receiver.GetAllServiceInfo()
	if err != nil {
		return "", err
	}

	var httpIpPorts []string
	for _, info := range receiver.ServerInfos {
		httpIpPorts = append(httpIpPorts, info.HttpServer)
	}

	return receiver.LoadBalance.SelectOne(httpIpPorts), nil
}
