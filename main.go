package main

import (
	"fmt"
	"os"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
)

// PageSize is now upto 50
const PageSize = 50

func main() {
	accessKeyID := os.Getenv("ALICLOUD_ACCESS_KEY")
	accessKeySecret := os.Getenv("ALICLOUD_SECRET_KEY")
	defaultRegion := os.Getenv("ALICLOUD_DEFAULT_REGION")
	client := ecs.NewClient(accessKeyID, accessKeySecret)

	region := common.Region(defaultRegion)
	instanceStatus := ecs.InstanceStatus("Running")

	var instances []ecs.InstanceAttributesType
	var err error
	var paginationResult *common.PaginationResult

	recevied := 0
	pagination := &common.Pagination{
		PageNumber: 1,
		PageSize:   PageSize,
	}

	describeInstancesArgs := &ecs.DescribeInstancesArgs{
		RegionId:   region,
		Status:     instanceStatus,
		Pagination: *pagination,
	}
	fmt.Printf("\n\nPAGE: %v", pagination.PageNumber)
	for {
		instances, paginationResult, err = client.DescribeInstances(describeInstancesArgs)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("%+v\n", instances[0])
			fmt.Printf("%+v\n", paginationResult)
		}
		fmt.Printf("\n\nPAGE: %v\n", pagination.PageNumber)

		recevied += paginationResult.PageSize
		if recevied >= paginationResult.TotalCount {
			break
		}

		pagination.PageNumber++
		describeInstancesArgs = &ecs.DescribeInstancesArgs{
			RegionId:   region,
			Status:     instanceStatus,
			Pagination: *pagination,
		}

	}
}
