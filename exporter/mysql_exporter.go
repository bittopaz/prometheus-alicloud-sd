package exporter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

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

func DiscoveryAlicloudMysql(filePath string) {
	var p []NodeInfo
	var q NodeInfo

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

	request := ecs.CreateDescribeInstancesRequest()
	request.PageSize = "100"
	request.InstanceName = "mysql*"
	response, err := ecsClient.DescribeInstances(request)
	if err != nil {
		panic(err)
	}
	for _, v := range response.Instances.Instance {
		q.Targets = append(q.Targets, v.InstanceName+":9100")
		q.Targets = append(q.Targets, v.InstanceName+":9104")
		for _, y := range v.Tags.Tag {
			if y.TagKey == "Env" {
				q.Labels.Env = y.TagValue
			} else if y.TagKey == "Job" {
				q.Labels.Job = y.TagValue
			} else if y.TagKey == "Loc" {
				q.Labels.Loc = y.TagValue
			} else if y.TagKey == "Service" {
				q.Labels.Service = y.TagValue
			} else if y.TagKey == "Tier" {
				q.Labels.Tier = y.TagValue
			}
		}
		p = append(p, q)
		q.Targets = nil
	}
	jsonScrapeConfig, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		fmt.Println("json err", err)
	}
	err = ioutil.WriteFile(filePath, jsonScrapeConfig, 0644)
	if err != nil {
		panic(err)
	}
}
