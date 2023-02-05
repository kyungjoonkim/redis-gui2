package rctx

import (
	"changeme/rds-client/client"
	"changeme/rds-client/hashtag"
	"changeme/rds-client/model/bmodel"
	result_model "changeme/rds-client/model/rmodel"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type RedisClusterSlot struct {
	start, end int64
	nodeInfos  []*ClusterNodeInfo
}

type RedisCommand struct {
	name     string
	arity    int64
	flags    []string
	firstKey int64
	lastKey  int64
	step     int64
}

type ClusterNodeInfo struct {
	ip     string
	port   int64
	nodeID string
	client *client.Client
}

type RedisCommandContext struct {
	redisCommandMap map[string]*RedisCommand
}

type RedisClientContext interface {
	GetNodeNameList() []string
	ScanRedisKeyInTargetNode(ipPortInNode string, cursor int64) *result_model.RedisScanResult
	Close()
	SendCommand(command []interface{}) (*client.Result, error)
	GetRedisKeyData(nodeIpAndPort string, param *bmodel.RedisGetParamModel) *bmodel.RedisGetResModel
}

type SingleRedisClientContext struct {
	client    *client.Client
	db        int
	ipAndPort string
	*RedisCommandContext
}

func (src *SingleRedisClientContext) GetNodeNameList() []string {
	return []string{src.ipAndPort}
}

func (src *SingleRedisClientContext) ScanRedisKeyInTargetNode(ipPortInNode string, cursor int64) *result_model.RedisScanResult {
	return ScanRedisKeyInTargetNode(src.client, cursor)
}

func (src *SingleRedisClientContext) Close() {
	err := src.client.Close()
	if err != nil {
		return
	}
}

func (src *SingleRedisClientContext) SendCommand(command []interface{}) (*client.Result, error) {
	if src.redisCommandMap == nil {
		return nil, errors.New("Not Redis Command")
	}

	commandStr := strings.ToLower(command[0].(string))
	commandInfo := src.redisCommandMap[commandStr]

	if commandInfo == nil {
		return nil, errors.New("Not Redis Command")
	}

	result, _ := src.client.Send(command)
	return result, nil
}

func (src *SingleRedisClientContext) GetRedisKeyData(nodeIpAndPort string, param *bmodel.RedisGetParamModel) *bmodel.RedisGetResModel {
	return getRedisKeyData(src.client, param)
}

type ClusterRedisClientContext struct {
	ClusterNodeInfos []*RedisClusterSlot
	nodeInfoMap      map[string]*ClusterNodeInfo
	*RedisCommandContext
}

func (crc *ClusterRedisClientContext) GetNodeNameList() []string {
	result := make([]string, 0)
	for _, clusterInfo := range crc.ClusterNodeInfos {
		for _, nodeInfo := range clusterInfo.nodeInfos {
			ipAndPort := nodeInfo.ip + ":" + strconv.FormatInt(nodeInfo.port, 10)
			result = append(result, strings.TrimSpace(ipAndPort))
		}
	}
	return result
}

func (crc *ClusterRedisClientContext) ScanRedisKeyInTargetNode(ipPortInNode string, cursor int64) *result_model.RedisScanResult {
	for _, clusterNodeInfo := range crc.ClusterNodeInfos {
		for _, nodeInfo := range clusterNodeInfo.nodeInfos {
			targetIpAndPort := nodeInfo.ip + ":" + strconv.FormatInt(nodeInfo.port, 10)
			if ipPortInNode == targetIpAndPort {
				return ScanRedisKeyInTargetNode(nodeInfo.client, cursor)
			}
		}
	}
	return nil
}

func (crc *ClusterRedisClientContext) Close() {
	for _, clusterInfo := range crc.ClusterNodeInfos {
		for _, nodeInfo := range clusterInfo.nodeInfos {
			if nodeInfo.client != nil {
				nodeInfo.client.Close()
			}
		}
	}
}

func (crc *ClusterRedisClientContext) SendCommand(command []interface{}) (*client.Result, error) {
	if crc.redisCommandMap == nil {
		return nil, errors.New("Not Redis Command")
	}

	commandStr := strings.ToLower(command[0].(string))
	commandInfo := crc.redisCommandMap[commandStr]

	if commandInfo == nil {
		return nil, errors.New("Not Redis Command")
	}

	key := command[commandInfo.firstKey].(string)
	key = hashtag.Key(key)

	targetSlot := hashtag.Slot(key)
	targetClusterSlot := slotNodes(targetSlot, crc.ClusterNodeInfos)

	result, _ := targetClusterSlot.nodeInfos[0].client.Send(command)
	return result, nil

}

func (crc *ClusterRedisClientContext) GetRedisKeyData(nodeIpAndPort string, param *bmodel.RedisGetParamModel) *bmodel.RedisGetResModel {
	targetNodeInfo := crc.nodeInfoMap[nodeIpAndPort]
	return getRedisKeyData(targetNodeInfo.client, param)
}
