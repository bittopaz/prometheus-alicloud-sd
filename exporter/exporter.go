package exporter

import (
	"fmt"
	"os"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

const PAGESIZE = 100

type NodeInfo struct {
	Targets []string `json:"targets"`
	Labels  Lable    `json:"lables"`
}

type Lable struct {
	Env     string `json:"env"`
	Job     string `json:"job"`
	Loc     string `json:"loc"`
	Service string `json:"service"`
	Tier    string `json:"tier"`
}

func EcsClient() (client *ecs.Client) {
	REGIONID := os.Getenv("ALICLOUD_DEFAULT_REGION")
	ACCESSKEY := os.Getenv("ALICLOUD_ACCESS_KEY")
	SECRETKEY := os.Getenv("ALICLOUD_SECRET_KEY")
	ecsClient, err := ecs.NewClientWithAccessKey(
		REGIONID,
		ACCESSKEY,
		SECRETKEY,
	)
	if err != nil {
		panic(err)
	}
	return ecsClient
}

func GetInstancesTotalCount(exportertype string) (totalcount int) {
	ecsClient := EcsClient()
	request := ecs.CreateDescribeInstancesRequest()
	if exportertype == "node" {
		request.Tag3Key = "Monitoring"
	} else if exportertype == "mysql" {
		request.InstanceName = "mysql*"
	}
	response, err := ecsClient.DescribeInstances(request)
	if err != nil {
		fmt.Println(err)
	}
	totalcount = response.TotalCount
	return totalcount
}
