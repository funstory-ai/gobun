package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/funstory-ai/gobun/adaptors/xiangongyun"
)

func main() {
	pool := xiangongyun.NewPool("Bearer " + os.Getenv("XGY_TOKEN"))
	pods, err := pool.ListPods()
	if err != nil {
		panic(err)
	}

	// Create a new tab writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "ID\tNAME\tSTATUS\tGPU\tGPU MODEL\tMEMORY")

	for _, pod := range pods {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
			pod.ID,
			pod.Name,
			pod.Status,
			pod.GPUCount,
			pod.GPUModel,
			humanReadableMemory(pod.MemorySize),
		)
	}

	// Flush the writer
	w.Flush()
}

// humanReadableMemory converts bytes to a human-readable format
func humanReadableMemory(bytes int64) string {
	const (
		KB = 1 << 10
		MB = 1 << 20
		GB = 1 << 30
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
