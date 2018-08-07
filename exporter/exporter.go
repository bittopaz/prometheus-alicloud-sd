package exporter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/r-kvstore"
)

const PAGESIZE = 100

type NodeInfo struct {
	Targets []string `json:"targets"`
	Labels  Label    `json:"labels"`
}

type Label struct {
	Env       string `json:"env"`
	Job       string `json:"job"`
	Component string `json:"component"`
	Service   string `json:"service"`
}

type alicloudAccessConfig struct {
	AlicloudRegionID  string
	AlicloudAccessKey string
	AlicloudSecretKey string
	SecurityToken     string
	Env               string
}

func (i *alicloudAccessConfig) init() {
	//get rolename
	if os.Getenv("ALICLOUD_DEFAULT_REGION") == "" ||
		os.Getenv("ALICLOUD_ACCESS_KEY") == "" ||
		os.Getenv("ALICLOUD_SECRET_KEY") == "" {
		resp, _ := http.Get("http://100.100.100.200/latest/meta-data/ram/security-credentials/")
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		ROLENAME := string(body)

		//get region-id
		resp, _ = http.Get("http://100.100.100.200/latest/meta-data/region-id")
		body, _ = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		i.AlicloudRegionID = string(body)

		//according to the rolename, get a json file.
		resp, _ = http.Get("http://100.100.100.200/latest/meta-data/ram/security-credentials/" + ROLENAME)
		body, _ = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		jsonRaw := body

		//convert json file to map
		var roleMap map[string]*json.RawMessage
		json.Unmarshal(jsonRaw, &roleMap)

		//extract related content from map
		json.Unmarshal(*roleMap["AccessKeyId"], &i.AlicloudAccessKey)
		json.Unmarshal(*roleMap["AccessKeySecret"], &i.AlicloudSecretKey)
		json.Unmarshal(*roleMap["SecurityToken"], &i.SecurityToken)
		i.Env = "remote"
	} else {
		i.AlicloudRegionID = os.Getenv("ALICLOUD_DEFAULT_REGION")
		i.AlicloudAccessKey = os.Getenv("ALICLOUD_ACCESS_KEY")
		i.AlicloudSecretKey = os.Getenv("ALICLOUD_SECRET_KEY")
		i.Env = "local"
	}
}

func NewEcsClient() (client *ecs.Client) {
	var err error
	var config alicloudAccessConfig
	config.init()
	if config.Env == "remote" {
		sdkClient, err := sdk.NewClientWithStsToken(
			config.AlicloudRegionID,
			config.AlicloudAccessKey,
			config.AlicloudSecretKey,
			config.SecurityToken,
		)
		client = &ecs.Client{
			Client: *sdkClient,
		}
		if err != nil {
			panic(err)
		}
	} else {
		client, err = ecs.NewClientWithAccessKey(
			config.AlicloudRegionID,
			config.AlicloudAccessKey,
			config.AlicloudSecretKey,
		)
		if err != nil {
			panic(err)
		}
	}
	return client
}

func NewRedisClient() (client *r_kvstore.Client) {
	var err error
	var config alicloudAccessConfig
	config.init()
	if config.Env == "remote" {
		sdkClient, err := sdk.NewClientWithStsToken(
			config.AlicloudRegionID,
			config.AlicloudAccessKey,
			config.AlicloudSecretKey,
			config.SecurityToken,
		)
		client = &r_kvstore.Client{
			Client: *sdkClient,
		}
		if err != nil {
			panic(err)
		}
	} else {
		client, err = r_kvstore.NewClientWithAccessKey(
			config.AlicloudRegionID,
			config.AlicloudAccessKey,
			config.AlicloudSecretKey,
		)
		if err != nil {
			panic(err)
		}
	}
	return client
}

func GetInstancesTotalCount(exportertype string) (totalcount int) {
	ecsClient := NewEcsClient()
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
