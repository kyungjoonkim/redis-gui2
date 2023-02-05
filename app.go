package main

import (
	"changeme/rds-client/model/bmodel"
	result_model "changeme/rds-client/model/rmodel"
	"changeme/rds-client/rctx"
	"context"
	"fmt"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

type LoginResult struct {
	Success      bool   `json:"success"`
	ErrorMessage string `json:"errorMessage"`
}

var rdsContext rctx.RedisClientContext

func (a *App) Login(ipAndPort string, password string) *LoginResult {
	result := &LoginResult{}
	rContext, err := rctx.LoginRedisServer(ipAndPort, password)
	if err != nil {
		fmt.Println("fail :", err)
		result.ErrorMessage = err.Error()
		return result
	}
	rdsContext = rContext
	fmt.Println("success: ")
	result.Success = true
	return result
}

func (a *App) GetSlotList() []string {
	result := rdsContext.GetNodeNameList()
	fmt.Println("GetSlotList APP 2 : ", result)
	return result
}

func (a *App) GetScanRedisKey(nodeIpAndPort string, cursor int64) *result_model.RedisScanResult {
	return rdsContext.ScanRedisKeyInTargetNode(nodeIpAndPort, cursor)
}

func (a *App) GetRedisKeyData(nodeIpAndPort string, redisKey string, start int64, end int64) *bmodel.RedisGetResModel {
	fmt.Println("===GetRedisKeyData==")
	fmt.Println("nodeIpAndPort:{}, redisKey:{}, start:{}, end:{} ", nodeIpAndPort, redisKey, start, end)
	param := bmodel.NewRedisGetParamModel(redisKey, start, end)
	return rdsContext.GetRedisKeyData(nodeIpAndPort, param)
}
