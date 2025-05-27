package slurmapi

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type NodeLoad struct {
	Load    int
	CpuFree int
	CpuTot  int
	MemFree int
	MemTot  int
	GpuFree int
	GpuTot  int
	State   []string
}

type NodeInfo struct {
	GpuName string
	GpuMem  int
}

type NodesLoad map[string]NodeLoad

type scontrolShowNodesCpuLoad struct {
	Number int `json:"number"`
}

type scontrolShowNodesNode struct {
	Tres     string                   `json:"tres"`
	TresUsed string                   `json:"tres_used"`
	Name     string                   `json:"name"`
	CpuLoad  scontrolShowNodesCpuLoad `json:"cpu_load"`
	State    []string                 `json:"state"`
}

type scontrolShowNodes struct {
	Nodes []scontrolShowNodesNode `json:"nodes"`
}

type PastJob struct {
	JobID    string
	UserName string
	Elapsed  time.Duration
	GPUInfo  NodeInfo
	GPUCount int
	CPUs     int
	Mem      int
}

func GetNodesLoad() (NodesLoad, error) {
	out, err := exec.Command("scontrol", "show", "nodes", "--json").Output()
	if err != nil {
		return nil, err
	}

	scontrolShowNodesData := scontrolShowNodes{}
	err = json.Unmarshal(out, &scontrolShowNodesData)
	if err != nil {
		return nil, err
	}

	nodesLoad := NodesLoad{}
	for _, node := range scontrolShowNodesData.Nodes {
		cpuTot := 0
		cpuUsed := 0
		memTot := 0
		memUsed := 0
		gpuTot := 0
		gpuUsed := 0
		tres := strings.Split(node.Tres, ",")
		tresUsed := strings.Split(node.TresUsed, ",")

		for _, t := range tres {
			if strings.HasPrefix(t, "cpu=") {
				cpuTot, err = strconv.Atoi(strings.TrimPrefix(t, "cpu="))
				if err != nil {
					return nil, err
				}
			}

			if strings.HasPrefix(t, "mem=") {
				value := strings.TrimPrefix(t, "mem=")

				if strings.HasSuffix(value, "G") {
					memTot, err = strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(t, "mem="), "G"))
				}

				if strings.HasSuffix(value, "M") {
					memTot, err = strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(t, "mem="), "M"))
					memTot = memTot / 1024
				}

				if err != nil {
					return nil, err
				}
			}

			if strings.HasPrefix(t, "gres/gpu=") {
				gpuTot, err = strconv.Atoi(strings.TrimPrefix(t, "gres/gpu="))
				if err != nil {
					return nil, err
				}
			}
		}

		for _, t := range tresUsed {
			if strings.HasPrefix(t, "cpu=") {
				cpuUsed, err = strconv.Atoi(strings.TrimPrefix(t, "cpu="))
				if err != nil {
					return nil, err
				}
			}

			if strings.HasPrefix(t, "mem=") {
				value := strings.TrimPrefix(t, "mem=")
				if strings.HasSuffix(value, "G") {
					memUsed, err = strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(t, "mem="), "G"))
				}
				if strings.HasSuffix(value, "M") {
					memUsed, err = strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(t, "mem="), "M"))
					memUsed = memUsed / 1024
				}
				if err != nil {
					return nil, err
				}
			}

			if strings.HasPrefix(t, "gres/gpu=") {
				gpuUsed, err = strconv.Atoi(strings.TrimPrefix(t, "gres/gpu="))
				if err != nil {
					return nil, err
				}
			}
		}

		nodesLoad[node.Name] = NodeLoad{
			Load:    0,
			CpuFree: cpuTot - cpuUsed,
			CpuTot:  cpuTot,
			MemFree: memTot - memUsed,
			MemTot:  memTot,
			GpuFree: gpuTot - gpuUsed,
			GpuTot:  gpuTot,
			State:   node.State,
		}
	}

	return nodesLoad, nil
}

func GetNodesInfo() (map[string]NodeInfo, error) {
	out, err := exec.Command("sinfo", "-o", "%n %G %f").Output()
	if err != nil {
		return nil, err
	}

	nodesInfo := map[string]NodeInfo{}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines[1:] {
		values := strings.Fields(line)
		if len(values) < 3 {
			continue
		}

		nodeName := values[0]
		gpuName := strings.ToUpper(strings.Split(values[1], ":")[1])
		nodeFeatures := strings.Split(values[2], ",")

		var gpuMem int
		for _, feature := range nodeFeatures {
			if strings.HasPrefix(feature, "m") {
				mem, err := strconv.Atoi(strings.TrimPrefix(feature, "m"))
				if err != nil {
					return nil, err
				}
				if mem > gpuMem {
					gpuMem = mem
				}
			}
		}

		nodesInfo[nodeName] = NodeInfo{
			GpuName: gpuName,
			GpuMem:  gpuMem,
		}
	}

	return nodesInfo, nil
}

func parseDuration(duration string) (time.Duration, error) {
	days := 0
	hours := 0
	minutes := 0
	seconds := 0
	err := error(nil)

	parts := strings.Split(duration, "-")
	if len(parts) == 2 {
		days, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0, err
		}
		duration = parts[1]
	}

	parts = strings.Split(duration, ":")
	if len(parts) == 3 {
		hours, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0, err
		}
		minutes, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, err
		}
		seconds, err = strconv.Atoi(parts[2])
		if err != nil {
			return 0, err
		}
	}

	return time.Duration(days*24+hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second, nil
} 

func GetPastJobs(start time.Time, end time.Time) ([]PastJob, error) {
	nodesInfo, err := GetNodesInfo()
	if err != nil {
		return nil, err
	}
	out, err := exec.Command("sacct", "-a", "-S", start.Format("2006-01-02-15:04"), "-E", end.Format("2006-01-02-15:04"), "-X", "-o", "JobID,User,Elapsed,NodeList,AllocTRES", "--noheader", "--parsable2").Output()
	if err != nil {
		return nil, err
	}

	pastJobs := []PastJob{}
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		pastJob := PastJob{}
		values := strings.Split(line, "|")
		if len(values) < 5 {
			continue
		}
		pastJob.JobID = values[0]
		pastJob.UserName = values[1]
		elapsed, err := parseDuration(values[2])
		if err != nil {
			fmt.Println("Error parsing duration:", values[2])
			return nil, err
		}
		pastJob.Elapsed = elapsed
		nodes := strings.Split(values[3], ",")
		for _, node := range nodes {
			if _, ok := nodesInfo[node]; !ok {
				continue
			}
			pastJob.GPUInfo = nodesInfo[node]
		}
		tresString := values[4]
		tres := strings.Split(tresString, ",")
		for _, t := range tres {
			parts := strings.Split(t, "=")
			if len(parts) != 2 {
				continue
			}
			key, value := parts[0], parts[1]
			if key == "cpu" {
				pastJob.CPUs, err = strconv.Atoi(value)
				if err != nil {
					return nil, err
				}
			}
			if key == "mem" {
				if strings.HasSuffix(value, "G") {
					pastJob.Mem, err = strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(t, "mem="), "G"))
				}
				if strings.HasSuffix(value, "M") {
					pastJob.Mem, err = strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(t, "mem="), "M"))
					pastJob.Mem = pastJob.Mem / 1024
				}
				if err != nil {
					return nil, err
				}
			}
			if key == "gres/gpu" {
				pastJob.GPUCount, err = strconv.Atoi(value)
				if err != nil {
					return nil, err
				}
			}
		}
		pastJobs = append(pastJobs, pastJob)
	}

	return pastJobs, nil
}
