package main

import (
	"internal/slurmapi"
	"time"
	"fmt"
	"flag"
)

/*
type PastJob struct {
	JobID    string
	UserName string
	Elapsed  time.Duration
	GPUInfo  NodeInfo
	GPUCount int
	CPUs     int
	Mem      int
}

*/

var days = flag.Int("days", 7, "the number of days to look back")

func printTimeComparisonBarChart(times map[string]time.Duration, unit ...string) {
	if len(unit) == 0 {
		unit = append(unit, "Hours")
	}
	// find the maximum time
	maxTime := time.Duration(0)
	for _, time := range times {
		if time > maxTime {
			maxTime = time
		}
	}

	// find the maximum username length for padding
	maxUserLen := 0
	for user := range times {
		if len(user) > maxUserLen {
			maxUserLen = len(user)
		}
	}

	// print the bar chart
	for user, time := range times {
		fmt.Printf("%s: ", user)
		for i := 0; i < maxUserLen-len(user); i++ {
			fmt.Print(" ")
		}
		for i := 0; i < int(time.Seconds())*60/int(maxTime.Seconds()); i++ {
			fmt.Print("=")
		}
		//print total time in Hours
		fmt.Printf(" %.2f %s\n", time.Hours(), unit[0])
	}
}

func main() {
	flag.Parse()
	pastJobs, err := slurmapi.GetPastJobs(time.Now().Add(-time.Hour*24*time.Duration(*days)), time.Now())
	if err != nil {
		panic(err)
	}

	// compute total compute time for each user
	userComputeTime := make(map[string]time.Duration)
	for _, job := range pastJobs {
		userComputeTime[job.UserName] += job.Elapsed
	}

	// compute total CPU time for each user
	userCPUTime := make(map[string]time.Duration)
	for _, job := range pastJobs {
		userCPUTime[job.UserName] += time.Duration(job.CPUs) * job.Elapsed
	}

	// compute total GPU time for each user
	userGPUTime := make(map[string]time.Duration)
	for _, job := range pastJobs {
		userGPUTime[job.UserName] += time.Duration(job.GPUCount) * job.Elapsed
	}

	// compute total VRAM time for each user
	userVRAMTime := make(map[string]time.Duration)
	for _, job := range pastJobs {
		userVRAMTime[job.UserName] += time.Duration(job.GPUInfo.GpuMem) * job.Elapsed
	}

	// print the results
	fmt.Printf("Cluster usage in the last %d days:\n", *days)
	fmt.Println("Compute Time:")
	printTimeComparisonBarChart(userComputeTime)
	fmt.Println("\nCPU Time:")
	printTimeComparisonBarChart(userCPUTime, "CPU Hours")
	fmt.Println("\nGPU Time:")
	printTimeComparisonBarChart(userGPUTime, "GPU Hours")
	fmt.Println("\nVRAM Time:")
	printTimeComparisonBarChart(userVRAMTime, "Gb.Hours")
}
