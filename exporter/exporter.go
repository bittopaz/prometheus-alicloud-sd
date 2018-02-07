package exporter

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

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

type Instance struct {
	REGIONID  string
	ACCESSKEY string
	SECRETKEY string
	TOKEN     string
}

func EcsClient() (client *ecs.Client) {
	var i Instance
	var err error
	if os.Getenv("ALICLOUD_DEFAULT_REGION") == "" {
		path, err := exec.LookPath("curl")
		if err != nil {
			panic(err)
		}
		//curl path

		// //get instance-id
		// cmdGetInstanceID := exec.Command(path,
		// 	"http://100.100.100.200/latest/meta-data/instance-id")
		// instanceIDRaw, err := cmdGetInstanceID.Output()
		// if err != nil {
		// 	panic(err)
		// }
		// ID := string(instanceIDRaw)
		// cmdGetInstanceID.Run()

		//get rolename
		cmdGetRoleName := exec.Command(path,
			"http://100.100.100.200/latest/meta-data/ram/security-credentials/")
		roleNameRaw, err := cmdGetRoleName.Output()
		if err != nil {
			panic(err)
		}
		ROLENAME := string(roleNameRaw)
		cmdGetRoleName.Run()

		//according to the rolename, get a json file.
		cmdGetJSON := exec.Command(path,
			"http://100.100.100.200/latest/meta-data/ram/security-credentials/"+ROLENAME)
		jsonRaw, err := cmdGetJSON.Output()
		if err != nil {
			panic(err)
		}
		cmdGetJSON.Run()

		//convert json file to map
		var roleMap map[string]*json.RawMessage
		json.Unmarshal(jsonRaw, &roleMap)

		//extract related content from map
		json.Unmarshal(*roleMap["AccessKeyId"], &i.ACCESSKEY)
		json.Unmarshal(*roleMap["AccessKeySecret"], &i.SECRETKEY)
		json.Unmarshal(*roleMap["SecurityToken"], &i.TOKEN)

		//get instance name/environment/service/tier
		i.REGIONID = "cn-shanghai"
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
