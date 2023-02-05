package rds_client_test

import (
	client2 "changeme/rds-client/client"
	"changeme/rds-client/model/bmodel"
	"changeme/rds-client/rctx"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"testing"
)

func TestConnect222(t *testing.T) {
	tt := []struct {
		name     string
		commands []command
	}{
		{
			name: "execute a set and a get",
			commands: []command{

				{
					input: []interface{}{
						"GET",
						"test",
					},
					output: "Maurício",
				},
			},
		},
	}

	for _, ts := range tt {
		t.Run(ts.name, func(t *testing.T) {
			client, err := client2.Connect(context.Background(), "localhost:6379", "1111")
			require.NoError(t, err)

			defer client.Close()

			for _, c := range ts.commands {
				result, err := client.Send(c.input)
				require.NoError(t, err)
				fmt.Println(result.Content())

				assert.Equal(t, c.output, result.Content())
			}

		})
	}
}
func TestConnectClientNew(t *testing.T) {
	//rds_client.Connection(context.Background(), "127.0.0.1:6379", "1111")
	rContext, err := rctx.LoginRedisServer("localhost:6379", "1111")
	if err != nil {
		return
	}

	input := []interface{}{
		"cluster",
		"info",
	}
	result, _ := rContext.SendCommand(input)
	data := result.Content()
	fmt.Println(data)

}

func TestConnectClient2New(t *testing.T) {
	rContext, err := rctx.LoginRedisServer("en-tane-alpha-2.cocone:7000", "")
	if err != nil {
		return
	}
	fmt.Println(rContext)

	param := bmodel.NewRedisGetParamModel("HELPER_300000236", 0, 1)
	//param := bmodel.NewRedisGetParamModel("BA_300000300", 0, 1)
	result := rContext.GetRedisKeyData("10.120.100.242:7000", param)
	fmt.Println(result)
	//result, _ := rContext.GetRedisKeyData
	//data := result.Content()
	//fmt.Println(data)

}
