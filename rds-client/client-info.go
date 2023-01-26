package rds_client

import (
	"changeme/rds-client/hashtag"
	result_model "changeme/rds-client/result-model"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"net"
	"strconv"
	"strings"
	"time"
)

type RedisClientContext struct {
	isCluster        bool //클러스터 여부
	client           *Client
	db               int                 //db 번호
	clusterNodeInfos []*RedisClusterSlot //Ip별 Client Map
	redisCommand     map[string]*RedisCommand
	password         string
	ipAndPort        string
}

func (rdsCtx *RedisClientContext) GetNodeNameList() []string {
	result := make([]string, 0)
	if rdsCtx.isCluster {
		for _, clusterInfo := range rdsCtx.clusterNodeInfos {
			for _, nodeInfo := range clusterInfo.nodeInfos {
				ipAndPort := nodeInfo.ip + ":" + strconv.FormatInt(nodeInfo.port, 10)
				result = append(result, strings.TrimSpace(ipAndPort))
			}
		}
	} else {
		result = append(result, rdsCtx.ipAndPort)
		return result
	}
	return result
}

func (rdsCtx *RedisClientContext) ScanRedisKeyInTargetNode(ipPortInNode string, cursor int64) *result_model.RedisScanResult {

	var targetClient *Client = nil

	if rdsCtx.isCluster {
	outer:
		for _, clusterNodeInfo := range rdsCtx.clusterNodeInfos {
			for _, nodeInfo := range clusterNodeInfo.nodeInfos {
				targetIpAndPort := nodeInfo.ip + ":" + strconv.FormatInt(nodeInfo.port, 10)
				if ipPortInNode == targetIpAndPort {
					targetClient = nodeInfo.client
					break outer
				}
			}
		}

	} else {
		targetClient = rdsCtx.client
	}

	scanResult, err := targetClient.Send([]interface{}{
		"scan",
		strconv.FormatInt(cursor, 10),
		"count",
		"800",
		"match",
		"*",
	})

	result := &result_model.RedisScanResult{}
	if err != nil {
		result.ErrorMessage = err.Error()
		return result
	}

	scanSlice, ok := scanResult.Content().([]interface{})
	if !ok {
		result.ErrorMessage = err.Error()
		return result
	}

	for index, scanData := range scanSlice {
		if index == 0 {
			fmt.Println(scanData)
			parseCursor, err := strconv.ParseInt(scanData.(string), 10, 64)
			if err != nil {
				result.ErrorMessage = err.Error()
				return result
			}
			result.Cursor = parseCursor
			if result.Cursor == 0 {
				result.Finish = true
			}
			continue
		}
		keys, ok := scanData.([]interface{})
		if !ok {
			break
		}
		resultKeys := make([]string, 0)
		for _, key := range keys {
			resultKeys = append(resultKeys, key.(string))
		}

		result.Keys = resultKeys
	}

	result.Success = true

	return result
}

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
	client *Client
}

func Connection(address string, passWord string) (*RedisClientContext, error) {
	rContext, err := makeRedisConnectionContext(address, passWord)
	if err != nil {
		return nil, err
	}

	err = authConnection(rContext.client, rContext.password)
	if err != nil {
		return nil, err
	}

	SetClusterIfExist(rContext)

	return rContext, nil
}

func makeConnection(address string) (*net.Conn, error) {
	dialer := net.Dialer{
		Timeout:   time.Second * 5,
		KeepAlive: time.Second * 10,
	}

	conn, err := dialer.DialContext(context.Background(), NetWorkType, address)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to %v", address)
	}
	return &conn, nil
}

func makeRedisConnectionContext(address string, password string) (*RedisClientContext, error) {
	conn, err := makeConnection(address)
	if err != nil {
		return nil, err
	}

	rContext := &RedisClientContext{
		isCluster: false,
		db:        0,
		client: &Client{
			connection: *conn,
			reader:     NewReader(*conn),
			writer:     NewWriter(*conn),
		},
		password:  password,
		ipAndPort: address,
	}
	return rContext, nil
}

func authConnection(client *Client, passWord string) error {
	if len(passWord) > 0 {
		_, err := client.Send([]interface{}{
			"AUTH",
			passWord,
		})

		if err != nil {
			return err
		}
	}
	return nil
}

func SetClusterIfExist(rContext *RedisClientContext) {
	connectClient := rContext.client
	result, err := connectClient.Send([]interface{}{
		"cluster",
		"slots",
	})

	if err != nil {
		return
	}

	_, errOk := result.Content().(error)
	if errOk {
		return
	}

	slots, ok := result.Content().([]interface{})
	if !ok {
		return
	}

	redisClusterSlots := make([]*RedisClusterSlot, len(slots))
	for index, targetSlotInfos := range slots {
		targetSlotInfoSlice, ok := targetSlotInfos.([]interface{})
		if !ok {
			continue
		}

		slotInfo := getSlotInfo(targetSlotInfoSlice)
		for _, nodeInfo := range slotInfo.nodeInfos {
			conn, err := makeConnection(nodeInfo.ip + ":" + strconv.FormatInt(nodeInfo.port, 10))
			if err != nil {
				continue
			}

			client := &Client{
				connection: *conn,
				reader:     NewReader(*conn),
				writer:     NewWriter(*conn),
			}

			err = authConnection(client, rContext.password)
			if err != nil {
				continue
			}

			nodeInfo.client = client
		}
		redisClusterSlots[index] = slotInfo
	}
	if len(redisClusterSlots) > 0 {
		rContext.isCluster = true
	}

	rContext.clusterNodeInfos = redisClusterSlots

	if rContext.isCluster {
		commandInfos, _ := getCommandInfoMap(rContext.client)
		rContext.redisCommand = commandInfos
	}
}

func getSlotInfo(targetSlotInfoSlice []interface{}) *RedisClusterSlot {
	result := &RedisClusterSlot{}
	result.nodeInfos = make([]*ClusterNodeInfo, 0)

	for index, targetSlotInfo := range targetSlotInfoSlice {
		if index == 0 { //start
			result.start = targetSlotInfo.(int64)
		} else if index == 1 { //end
			result.end = targetSlotInfo.(int64)
		} else if index > 1 { //
			if nodeInfoSlice, ok := targetSlotInfo.([]interface{}); ok {
				resultNodeInfo := &ClusterNodeInfo{}
				for nodeIndex, nodeInfo := range nodeInfoSlice {
					if nodeIndex == 0 {
						resultNodeInfo.ip = nodeInfo.(string)
					} else if nodeIndex == 1 {
						resultNodeInfo.port = nodeInfo.(int64)
					} else if nodeIndex == 2 {
						resultNodeInfo.nodeID = nodeInfo.(string)

					}
				}
				result.nodeInfos = append(result.nodeInfos, resultNodeInfo)
			}

		}
	}
	return result
}

func getCommandInfoMap(client *Client) (map[string]*RedisCommand, error) {

	res, err := client.Send([]interface{}{
		"command",
	})

	if err != nil {
		return nil, err
	}

	result := make(map[string]*RedisCommand)

	resCmds, ok := res.Content().([]interface{})
	if !ok {
		return nil, err
	}

	for _, resTargetCmds := range resCmds {
		resTargetCmdInfos, ok := resTargetCmds.([]interface{})
		if !ok {
			continue
		}
		resultCommand := &RedisCommand{}
		for index, resTargetCmdInfo := range resTargetCmdInfos {
			if index == 0 {
				resultCommand.name = resTargetCmdInfo.(string)
				continue
			}

			if index == 1 {
				resultCommand.arity = resTargetCmdInfo.(int64)
				continue
			}

			if index == 2 {
				resFlags, ok := resTargetCmdInfo.([]interface{})
				if ok {
					resultCommand.flags = make([]string, 0)
					for _, resFlag := range resFlags {
						resultCommand.flags = append(resultCommand.flags, resFlag.(string))
					}
				}
				continue
			}

			if index == 3 {
				resultCommand.firstKey = resTargetCmdInfo.(int64)
				continue
			}

			if index == 4 {
				resultCommand.lastKey = resTargetCmdInfo.(int64)
				continue
			}

			if index == 5 {
				resultCommand.step = resTargetCmdInfo.(int64)
				continue
			}

		}
		if len(resultCommand.name) > 0 {
			result[resultCommand.name] = resultCommand
		}

	}
	return result, nil
}

func Close(rContext *RedisClientContext) {
	rContext.client.Close()
	for _, clusterInfo := range rContext.clusterNodeInfos {
		for _, nodeInfo := range clusterInfo.nodeInfos {
			if nodeInfo.client != nil {
				nodeInfo.client.Close()
			}
		}
	}
}

func SendCommand(rContext *RedisClientContext, command []interface{}) (*Result, error) {
	if rContext.isCluster == true {

		if rContext.redisCommand == nil {
			return nil, errors.New("Not Redis Command")
		}

		commandStr := strings.ToLower(command[0].(string))
		commandInfo := rContext.redisCommand[commandStr]

		if commandInfo == nil {
			return nil, errors.New("Not Redis Command")
		}

		key := command[commandInfo.firstKey].(string)
		key = hashtag.Key(key)

		targetSlot := hashtag.Slot(key)
		targetClusterSlot := slotNodes(targetSlot, rContext.clusterNodeInfos)

		//clusterNodeInfo := targetClusterSlot.nodeInfos[0]
		//cmdInfo := clusterNodeInfo.redisCommand[command[0].(string)]
		result, _ := targetClusterSlot.nodeInfos[0].client.Send(command)
		return result, nil

	} else {

	}
	return nil, nil
}

func slotNodes(slotNum int64, slots []*RedisClusterSlot) *RedisClusterSlot {
	for _, slot := range slots {
		if slotNum >= slot.start && slotNum <= slot.end {
			return slot
		}
	}
	//
	//
	//i := sort.Search(len(slots), func(i int) bool {
	//	return slots[i].end >= slotNum
	//})
	//if i >= len(slots) {
	//	return nil
	//}
	//
	//x := slots[i]
	//if slotNum >= x.start && slotNum <= x.end {
	//	return x
	//}
	return nil
}
