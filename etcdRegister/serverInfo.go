package etcdRegister

import (
	"github.com/google/uuid"
	"net"
	"strings"
)

type ServerInfo struct {
	Id         string         `json:"id"`
	ServerName string         `json:"serverName"`
	HttpServer string         `json:"httpServer"`
	GrpcServer string         `json:"grpcServer"`
	MetaData   map[string]any `json:"metaData"`
	HttpPort   string         `json:"-"`
	GrpcPort   string         `json:"-"`
}

func getIpAddress() string {
	//获取本机ip地址，不管是不是送达了，都能解析到本机的ip地址
	conn, err := net.Dial("udp", "1.1.1.1:53")
	if err != nil {
		panic(err)
	}
	addr := conn.LocalAddr().(*net.UDPAddr)
	ip := strings.Split(addr.String(), ":")[0]
	return ip
}

func NewServerInfo(ServerName string, grpcPort string, httpPort string) *ServerInfo {
	info := &ServerInfo{ServerName: ServerName, HttpPort: httpPort, GrpcPort: grpcPort}
	newUUID, _ := uuid.NewUUID()
	info.Id = newUUID.String()

	if grpcPort != "" {
		if startIndex := strings.Index(grpcPort, ":"); startIndex != -1 {
			grpcPort := grpcPort[startIndex+1:]
			info.GrpcServer = getIpAddress() + ":" + grpcPort
		}
	}

	if httpPort != "" {
		if startIndex := strings.Index(httpPort, ":"); startIndex != -1 {
			grpcPort := httpPort[startIndex+1:]
			info.HttpServer = getIpAddress() + ":" + grpcPort
		}
	}

	return info
}
