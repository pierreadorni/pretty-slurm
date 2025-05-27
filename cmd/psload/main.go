package main

import (
	"flag"
	"fmt"
	"slices"
	"sort"

	"github.com/acarl005/stripansi"
	"github.com/fatih/color"
	"github.com/mattn/go-runewidth"
	"github.com/rodaine/table"

	"internal/slurmapi"
)

var bestFlag = flag.Bool("best", false, "show only the best node")

func fullOutput(idleNodes, gpuNodes, noGpuNodes []string, nodesInfo map[string]slurmapi.NodeInfo, nodesLoad map[string]slurmapi.NodeLoad, bestNode string, bestNodeVram int) {
	bold := color.New(color.Bold)
	bold.Println("🔥 Unused Nodes")
	tbl := table.New("GPUs", "MEMORY", "CPUs", "NODE")
	greenBold := color.New(color.FgGreen, color.Bold)
	green := color.New(color.FgGreen)
	tbl.WithFirstColumnFormatter(tabulateFormatter).WithHeaderFormatter(tabulateFormatter).WithWidthFunc(widthOfColoredString)
	// sort by name alphabetically
	sort.Slice(idleNodes, func(i, j int) bool {
		return idleNodes[i] < idleNodes[j]
	})
	for _, nodeName := range idleNodes {
		info := nodesInfo[nodeName]
		bestNodeStr := ""
		printColor := green
		if nodeName == bestNode {
			bestNodeStr = fmt.Sprintf("\t⟵ Best Node (%dG VRAM)", bestNodeVram)
			printColor = greenBold
		}
		tbl.AddRow(
			printColor.Sprintf("%d×%s(%dG)", nodesLoad[nodeName].GpuFree, info.GpuName, info.GpuMem),
			printColor.Sprintf("%dG", nodesLoad[nodeName].MemFree),
			printColor.Sprintf("%d", nodesLoad[nodeName].CpuFree),
			printColor.Sprintf("%s", nodeName)+bestNodeStr,
		)
	}
	tbl.Print()

	bold.Println("\n✨ Free GPUs")
	tbl = table.New("GPUs", "MEMORY", "CPUs", "NODE")
	yellowBold := color.New(color.FgYellow, color.Bold)
	yellow := color.New(color.FgYellow)
	tbl.WithFirstColumnFormatter(tabulateFormatter).WithHeaderFormatter(tabulateFormatter).WithWidthFunc(widthOfColoredString)
	// sort by name alphabetically
	sort.Slice(gpuNodes, func(i, j int) bool {
		return gpuNodes[i] < gpuNodes[j]
	})
	for _, nodeName := range gpuNodes {
		info := nodesInfo[nodeName]
		bestNodeStr := ""
		printColor := yellow
		if nodeName == bestNode {
			bestNodeStr = fmt.Sprintf("\t⟵ Best Node (%dG VRAM)", bestNodeVram)
			printColor = yellowBold
		}
		tbl.AddRow(
			printColor.Sprintf("%d×%s(%dG)", nodesLoad[nodeName].GpuFree, info.GpuName, info.GpuMem),
			printColor.Sprintf("%dG", nodesLoad[nodeName].MemFree),
			printColor.Sprintf("%d", nodesLoad[nodeName].CpuFree),
			printColor.Sprintf("%s", nodeName)+bestNodeStr,
		)
	}
	tbl.Print()

	bold.Println("\n💀 No GPUs")
	tbl = table.New("GPUs", "MEMORY", "CPUs", "NODE")
	red := color.New(color.FgRed)
	tbl.WithFirstColumnFormatter(tabulateFormatter).WithHeaderFormatter(tabulateFormatter).WithWidthFunc(widthOfColoredString)
	// sort by name alphabetically
	sort.Slice(noGpuNodes, func(i, j int) bool {
		return noGpuNodes[i] < noGpuNodes[j]
	})
	for _, nodeName := range noGpuNodes {
		info := nodesInfo[nodeName]
		tbl.AddRow(
			red.Sprintf("%d×%s(%dG)", nodesLoad[nodeName].GpuFree, info.GpuName, info.GpuMem),
			red.Sprintf("%dG", nodesLoad[nodeName].MemFree),
			red.Sprintf("%d", nodesLoad[nodeName].CpuFree),
			red.Sprintf("%s", nodeName),
		)
	}
	tbl.Print()
}

func bestOutput(idleNodes, gpuNodes, noGpuNodes []string, nodesInfo map[string]slurmapi.NodeInfo, nodesLoad map[string]slurmapi.NodeLoad, bestNode string, bestNodeVram int) {
	tbl := table.New("GPUs", "MEMORY", "CPUs", "NODE")
	printColor := color.FgGreen
	if slices.Contains(gpuNodes, bestNode) {
		printColor = color.FgYellow
	}
	if slices.Contains(noGpuNodes, bestNode) {
		printColor = color.FgRed
	}
	printColorBold := color.New(printColor, color.Bold)
	tbl.WithFirstColumnFormatter(tabulateFormatter).WithHeaderFormatter(tabulateFormatter).WithWidthFunc(widthOfColoredString)
	// sort by name alphabetically

	info := nodesInfo[bestNode]
	bestNodeStr := fmt.Sprintf("\t⟵ Best Node (%dG VRAM)", bestNodeVram)

	tbl.AddRow(
		printColorBold.Sprintf("%d×%s(%dG)", nodesLoad[bestNode].GpuFree, info.GpuName, info.GpuMem),
		printColorBold.Sprintf("%dG", nodesLoad[bestNode].MemFree),
		printColorBold.Sprintf("%d", nodesLoad[bestNode].CpuFree),
		printColorBold.Sprintf("%s", bestNode)+bestNodeStr,
	)

	tbl.Print()
}

func main() {
	flag.Parse()
	nodesInfo, err := slurmapi.GetNodesInfo()
	if err != nil {
		fmt.Println(err)
		return
	}

	nodesLoad, err := slurmapi.GetNodesLoad()
	if err != nil {
		fmt.Println(err)
		return
	}

	// find all nodes that are idle
	idleNodes := []string{}
	for nodeName, nodeLoad := range nodesLoad {
		if slices.Contains(nodeLoad.State, "IDLE") {
			idleNodes = append(idleNodes, nodeName)
		}
	}

	// find all nodes with available GPU
	gpuNodes := []string{}
	for nodeName, nodeLoad := range nodesLoad {
		if !slices.Contains(nodeLoad.State, "IDLE") && nodeLoad.GpuFree > 0 {
			gpuNodes = append(gpuNodes, nodeName)
		}
	}

	// all nodes with no GPU available
	noGpuNodes := []string{}
	for nodeName, nodeLoad := range nodesLoad {
		if nodeLoad.GpuFree == 0 {
			noGpuNodes = append(noGpuNodes, nodeName)
		}
	}

	// find the best node currently (max free cumulative VRAM)
	bestNode := ""
	bestNodeVram := 0
	for nodeName, nodeLoad := range nodesLoad {
		nodeFreeVram := nodeLoad.GpuFree * nodesInfo[nodeName].GpuMem
		if nodeFreeVram > bestNodeVram {
			bestNodeVram = nodeFreeVram
		}
	}

	// find all nodes that have the same cumulative VRAM as the best node
	bestNodes := []string{}
	for nodeName, nodeLoad := range nodesLoad {
		nodeFreeVram := nodeLoad.GpuFree * nodesInfo[nodeName].GpuMem
		if nodeFreeVram == bestNodeVram {
			bestNodes = append(bestNodes, nodeName)
		}
	}

	// order the best nodes by: 1. avail RAM (max), 2. number of gpus (min), 3. number of cpus (max)
	if len(bestNodes) > 1 {
		sort.Slice(bestNodes, func(i, j int) bool {
			nodeLoadI := nodesLoad[bestNodes[i]]
			nodeLoadJ := nodesLoad[bestNodes[j]]
			if nodeLoadI.MemFree != nodeLoadJ.MemFree {
				return nodeLoadI.MemFree > nodeLoadJ.MemFree
			}
			if nodeLoadI.GpuFree != nodeLoadJ.GpuFree {
				return nodeLoadI.GpuFree < nodeLoadJ.GpuFree
			}
			return nodeLoadI.CpuFree > nodeLoadJ.CpuFree
		})
	}
	bestNode = bestNodes[0]

	if *bestFlag {
		bestOutput(idleNodes, gpuNodes, noGpuNodes, nodesInfo, nodesLoad, bestNode, bestNodeVram)
	} else {
		fullOutput(idleNodes, gpuNodes, noGpuNodes, nodesInfo, nodesLoad, bestNode, bestNodeVram)
	}
}

func tabulateFormatter(format string, vals ...interface{}) string {
	return "\t" + fmt.Sprintf(format, vals...)
}

func widthOfColoredString(s string) int {
	// remove escape sequences
	s = stripansi.Strip(s)
	return runewidth.StringWidth(s)
}
