package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
)

// PageSize is limited 50 from official
const PageSize = 50

type scrapeTask struct {
	Targets sort.StringSlice  `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

type target struct {
	host    string
	env     string
	tier    string
	loc     string
	service string
}

func discoveryAlicloud(filePath string) {
	accessKeyID := os.Getenv("ALICLOUD_ACCESS_KEY")
	accessKeySecret := os.Getenv("ALICLOUD_SECRET_KEY")
	defaultRegion := os.Getenv("ALICLOUD_DEFAULT_REGION")
	client := ecs.NewClient(accessKeyID, accessKeySecret)

	region := common.Region(defaultRegion)
	instanceStatus := ecs.InstanceStatus("Running")

	var instances []ecs.InstanceAttributesType
	var pageInstances []ecs.InstanceAttributesType
	var err error
	var paginationResult *common.PaginationResult

	//var scrapeTasks []scrapeTask

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

	for {
		pageInstances, paginationResult, err = client.DescribeInstances(describeInstancesArgs)

		if err != nil {
			fmt.Println(err)
		} else {
			instances = append(instances, pageInstances...)
		}

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
	//tasks := make(map[string]*scrapeTask)

	var targets []target

	validHostname := regexp.MustCompile(`^(?P<service>.+)-[0-9]+\.(?P<tier>[a-z]+)\.(?P<env>[a-z0-9]+)\.(?P<loc>[a-z0-9]+)\..+\..+`)
	for _, inst := range instances {
		if match := validHostname.FindStringSubmatch(inst.InstanceName); match != nil {

			tags := inst.Tags.Tag
			for _, eachTag := range tags {
				if eachTag.TagKey == "Monitoring" {
					if tagMonitoring, err := strconv.ParseBool(eachTag.TagValue); err == nil {
						if tagMonitoring == true {

							label := make(map[string]string)
							for i, name := range validHostname.SubexpNames() {
								if name != "" {
									label[name] = match[i]
								}
							}

							target := target{
								host:    fmt.Sprintf("%s:9100", inst.InstanceName),
								env:     label["env"],
								tier:    label["tier"],
								loc:     label["loc"],
								service: label["service"],
							}
							targets = append(targets, target)

							// label["job"] = "node"
							// key := fmt.Sprintf("%s.%s.%s.%s", label["service"], label["tier"], label["env"], label["loc"])
							// if _, found := tasks[key]; found {
							// tasks[key].Targets = append(tasks[key].Targets, fmt.Sprintf("%s:9100", inst.InstanceName))
							// } else {
							// tasks[key] = &scrapeTask{
							// Targets: []string{fmt.Sprintf("%s:9100", inst.InstanceName)},
							// Labels:  label,
							// }
							// }
						}
						break
					}
				}
			}
		}
	}
	// for _, value := range tasks {
	// 	value.Targets.Sort()
	// 	scrapeTasks = append(scrapeTasks, *value)
	// }
	// writeSDConfig(scrapeTasks, filePath)
}

func writeSDConfig(scrapeTasks []scrapeTask, output string) {
	jsonScrapeConfig, err := json.MarshalIndent(scrapeTasks, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Print("Writing Prometheus config file")

	err = ioutil.WriteFile(output, jsonScrapeConfig, 0644)
	if err != nil {
		panic(err)
	}
}

func main() {
	filePath := flag.String("f", "PATH", "Output filename")
	discoveryAlicloud(*filePath)
}
