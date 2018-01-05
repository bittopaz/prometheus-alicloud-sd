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
	"strings"

	"github.com/owitho/prometheus-alicloud-sd/exporter"

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
	host   string
	labels map[string]string
}

func discoveryAlicloud(filePath string) {
	accessKeyID := os.Getenv("ALICLOUD_ACCESS_KEY")
	accessKeySecret := os.Getenv("ALICLOUD_SECRET_KEY")
	defaultRegion := os.Getenv("ALICLOUD_DEFAULT_REGION")
	instanceTags := strings.Split(os.Getenv("ALICLOUD_SD_TAGS"), ",")
	client := ecs.NewClient(accessKeyID, accessKeySecret)

	region := common.Region(defaultRegion)
	instanceStatus := ecs.InstanceStatus("Running")

	var instances []ecs.InstanceAttributesType
	var pageInstances []ecs.InstanceAttributesType
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

	var targets []target

	validHostname := regexp.MustCompile(`^(?P<service>.+)-[0-9]+\.(?P<tier>[a-z]+)\.(?P<env>[a-z0-9]+)\.(?P<loc>[a-z0-9]+)\..+\..+`)
	for _, inst := range instances {
		if match := validHostname.FindStringSubmatch(inst.InstanceName); match != nil {

			ecsTags := inst.Tags.Tag

			tmpTags := make(map[string]string)
			tmpTags["Monitoring"] = "false"
			for _, eachTag := range ecsTags {
				tmpTags[eachTag.TagKey] = eachTag.TagValue
			}
			labels := make(map[string]string)
			if tagMonitoring, err := strconv.ParseBool(tmpTags["Monitoring"]); err == nil {
				if tagMonitoring == true {
					for _, eachLabel := range instanceTags {
						labels[eachLabel] = tmpTags[eachLabel]
					}
					promTarget := target{
						host:   fmt.Sprintf("%s:9100", inst.InstanceName),
						labels: labels,
					}
					targets = append(targets, promTarget)
				}
			}
		}
	}
	scrapeTasks := makeScapeTasks(targets, instanceTags)
	writeSDConfig(scrapeTasks, filePath)
}

func makeScapeTasks(targets []target, instanceTags []string) []scrapeTask {
	var scrapeTasks []scrapeTask

	tasks := make(map[string]*scrapeTask)

	for _, eachTarget := range targets {
		var values []string
		for _, eachTag := range instanceTags {
			values = append(values, eachTarget.labels[eachTag])
		}
		key := fmt.Sprintf(strings.Join(values, "."))
		if _, found := tasks[key]; found {
			tasks[key].Targets = append(tasks[key].Targets, eachTarget.host)
		} else {
			tasks[key] = &scrapeTask{
				Targets: []string{eachTarget.host},
				Labels:  eachTarget.labels,
			}
		}
	}

	for _, value := range tasks {
		value.Targets.Sort()
		scrapeTasks = append(scrapeTasks, *value)
	}

	return scrapeTasks
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
	var filePath string
	var nodetype string
	flag.StringVar(&filePath, "f", "", "Output filename")
	flag.StringVar(&nodetype, "t", "", "exporter type(node/mysql)")
	flag.Parse()

	if filePath == "" {
		fmt.Fprintf(os.Stderr, "required arguments -f must pass in.")
		os.Exit(1)
	}

	if nodetype == "node" {
		discoveryAlicloud(filePath)
	} else if nodetype == "mysql" {
		exporter.DiscoveryAlicloudMysql(filePath)
	} else if nodetype == "" {
		fmt.Fprintf(os.Stderr, "required arguments -t must pass in.")
		os.Exit(1)
	}
}
