package main

import (
	"flag"
	"fmt"
	"github.com/manamanmana/agt4ri/aggregation"
	"os"
	"strings"
)

var (
	profile    string
	assumeRole string
)

func init() {
	flag.StringVar(&profile, "profile", "", "Speficy your profile name in ~/.aws/credentials (optional)")
	flag.StringVar(&assumeRole, "assume-role", "", "Specify your assume role arn (optional)")
}

func main() {
	flag.Parse()
	//fmt.Println(profile)
	//fmt.Println(assumeRole)

	// Get all the regions
	var regions *[]string
	var err error
	regions, err = aggregation.Regions(profile, assumeRole)
	if err != nil {
		fmt.Printf("Error while getting all the regions: %v\n", err)
		os.Exit(1)
	}

	// Get all the instances of all the regions
	var instances *[][]string
	instances, err = aggregation.Instances(profile, assumeRole, regions)
	if err != nil {
		fmt.Printf("Error while getting all the instances: %v\n", err)
		os.Exit(2)
	}

	// Aggregation
	var mapresult *map[string]int
	mapresult = aggregation.DoAggregate(instances)

	// Final Output
	var key string
	var count int
	var keys []string
	for key, count = range *mapresult {
		keys = strings.Split(key, ":")
		availavilityZone := keys[0]
		instanceType := keys[1]
		platform := keys[2]
		tenamcy := keys[3]
		fmt.Printf("%s,%s,%s,%s,%d\n", availavilityZone, instanceType, platform, tenamcy, count)
	}

	os.Exit(0)
}
