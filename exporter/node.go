package exporter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

func DiscoveryAlicloudNode(filePath, exporterType string) {
	var nodeInfoList []NodeInfo
	var nodeinfo NodeInfo
	var flag bool = false
	ecsClient := EcsClient()
	totalcount := GetInstancesTotalCount(exporterType)

	for i := 0; i <= (totalcount / PAGESIZE); i++ {
		request := ecs.CreateDescribeInstancesRequest()
		request.PageSize = requests.NewInteger(PAGESIZE)
		request.PageNumber = requests.NewInteger(i + 1)
		request.Status = "Running"
		request.Tag1Key = "Monitoring"
		request.Tag1Value = "true"
		response, err := ecsClient.DescribeInstances(request)
		if err != nil {
			fmt.Println(err)
		}
		for _, v := range response.Instances.Instance {
			for _, y := range v.Tags.Tag {
				if y.TagKey == "Env" {
					nodeinfo.Labels.Env = y.TagValue
				} else if y.TagKey == "Job" {
					nodeinfo.Labels.Job = y.TagValue
				} else if y.TagKey == "Component" {
					nodeinfo.Labels.Component = y.TagValue
				} else if y.TagKey == "Service" {
					nodeinfo.Labels.Service = y.TagValue
				}

				if nodeinfo.Labels.Job == "" {
					nodeinfo.Labels.Job = "node"
				}
			}

			for m, n := range nodeInfoList {
				if n.Labels.Env == nodeinfo.Labels.Env && n.Labels.Service == nodeinfo.Labels.Service && n.Labels.Service != "" {
					nodeInfoList[m].Targets = append(nodeInfoList[m].Targets, v.InstanceName+":9100")
					flag = true
					break
				} else {
					flag = false
				}
			}
			if flag == false {
				nodeinfo.Targets = append(nodeinfo.Targets, v.InstanceName+":9100")
				nodeInfoList = append(nodeInfoList, nodeinfo)
			}
			nodeinfo.Targets = nil
			nodeinfo.Labels.Env = ""
			nodeinfo.Labels.Job = ""
			nodeinfo.Labels.Component = ""
			nodeinfo.Labels.Service = ""
		}
	}
	jsonScrapeConfig, err := json.MarshalIndent(nodeInfoList, "", "\t")

	if err != nil {
		fmt.Println("json err", err)
	}
	err = ioutil.WriteFile(filePath, jsonScrapeConfig, 0644)
	if err != nil {
		panic(err)
	}
}
