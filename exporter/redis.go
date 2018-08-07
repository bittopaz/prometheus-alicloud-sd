package exporter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/r-kvstore"
)

func DiscoveryAlicloudRedis(filePath, exporterType string) {
	var nodeinfolist []NodeInfo
	redisClient := NewRedisClient()

	request := r_kvstore.CreateDescribeInstancesRequest()

	request.PageSize = requests.NewInteger(50)

	response, _ := redisClient.DescribeInstances(request)

	for _, instance := range response.Instances.KVStoreInstance {
		n := NodeInfo{
			Targets: []string{instance.InstanceId},
			Labels: Label{
				Service:   instance.InstanceName,
				Component: "redis",
				Job:       "alicloud_redis",
				Env:       splitEnvFromName(instance.InstanceName),
			},
		}
		nodeinfolist = append(nodeinfolist, n)
	}

	jsonScrapeConfig, err := json.MarshalIndent(nodeinfolist, "", "\t")
	if err != nil {
		fmt.Println("json err", err)
	}
	err = ioutil.WriteFile(filePath, jsonScrapeConfig, 0644)
	if err != nil {
		panic(err)
	}
}

func splitEnvFromName(input string) (output string) {
	output = strings.Split(input, "_")[0]
	return
}
