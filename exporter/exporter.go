package exporter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
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

type alicloudAccessConfig struct {
	REGIONID  string
	ACCESSKEY string
	SECRETKEY string
	TOKEN     string
}

func EcsClient() (client *ecs.Client) {
	var i alicloudAccessConfig
	var err error
	if os.Getenv("ALICLOUD_DEFAULT_REGION") == "" &&
		os.Getenv("ALICLOUD_ACCESS_KEY") == "" &&
		os.Getenv("ALICLOUD_SECRET_KEY") == "" {
		//get rolename
		resp, _ := http.Get("http://100.100.100.200/latest/meta-data/ram/security-credentials/")
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		ROLENAME := string(body)

		//get region-id
		resp, _ = http.Get("http://100.100.100.200/latest/meta-data/region-id")
		body, _ = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		i.REGIONID = string(body)
		//according to the rolename, get a json file.
		resp, _ = http.Get("http://100.100.100.200/latest/meta-data/ram/security-credentials/" + ROLENAME)
		body, _ = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		jsonRaw := body

		//convert json file to map
		var roleMap map[string]*json.RawMessage
		json.Unmarshal(jsonRaw, &roleMap)

		//extract related content from map
		json.Unmarshal(*roleMap["AccessKeyId"], &i.ACCESSKEY)
		json.Unmarshal(*roleMap["AccessKeySecret"], &i.SECRETKEY)
		json.Unmarshal(*roleMap["SecurityToken"], &i.TOKEN)

		//get instance name/environment/service/tier
		ecsClient, err := sdk.NewClientWithStsToken(
			i.REGIONID,
			i.ACCESSKEY,
			i.SECRETKEY,
			i.TOKEN,
		)
		client = &ecs.Client{
			Client: *ecsClient,
		}

		if err != nil {
			panic(err)
		}
	} else {
		REGIONID := os.Getenv("ALICLOUD_DEFAULT_REGION")
		ACCESSKEY := os.Getenv("ALICLOUD_ACCESS_KEY")
		SECRETKEY := os.Getenv("ALICLOUD_SECRET_KEY")
		client, err = ecs.NewClientWithAccessKey(
			REGIONID,
			ACCESSKEY,
			SECRETKEY,
		)
		if err != nil {
			panic(err)
		}
	}
	return client
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
