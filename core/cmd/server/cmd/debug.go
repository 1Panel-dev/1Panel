package cmd

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/spf13/cobra"
)

var (
	cpuProfile    string
	memProfile    string
	duration      time.Duration
	showMemStats  bool
	showGoroutine bool
)

func init() {
	RootCmd.AddCommand(debugCmd)

	debugCmd.Flags().StringVar(&cpuProfile, "cpuprofile", "", "write cpu profile to file")
	debugCmd.Flags().StringVar(&memProfile, "memprofile", "", "write memory profile to file")
	debugCmd.Flags().DurationVar(&duration, "duration", 30*time.Second, "duration for cpu profiling")
	debugCmd.Flags().BoolVar(&showMemStats, "memstats", false, "show memory statistics")
	debugCmd.Flags().BoolVar(&showGoroutine, "goroutine", false, "show goroutine profile")
}

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Performance debugging with pprof",
	Long: `Debug command provides CPU and memory profiling capabilities using pprof.
	
Examples:
  # CPU profiling for 30 seconds
  1panel-core debug --cpuprofile=cpu.prof --duration=30s
  
  # Memory profiling
  1panel-core debug --memprofile=mem.prof
  
  # Show memory statistics
  1panel-core debug --memstats
  
  # Show goroutine information
  1panel-core debug --goroutine
  
  # Combined profiling
  1panel-core debug --cpuprofile=cpu.prof --memprofile=mem.prof --memstats`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if cpuProfile != "" {
			if err := startCPUProfile(cpuProfile); err != nil {
				return fmt.Errorf("failed to start CPU profile: %v", err)
			}
			defer stopCPUProfile()
			time.Sleep(duration)
			fmt.Printf("CPU profile saved to: %s\n", cpuProfile)
		}

		if memProfile != "" {
			if err := writeMemProfile(memProfile); err != nil {
				return fmt.Errorf("failed to write memory profile: %v", err)
			}
			fmt.Printf("Memory profile saved to: %s\n", memProfile)
		}

		if showMemStats {
			printMemStats()
		}

		if showGoroutine {
			printGoroutineInfo()
		}

		if cpuProfile == "" && memProfile == "" && !showMemStats && !showGoroutine {
			printBasicInfo()
		}
		return nil
	},
}

func startCPUProfile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		return err
	}

	return nil
}

func stopCPUProfile() {
	pprof.StopCPUProfile()
}

func writeMemProfile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	runtime.GC()

	if err := pprof.WriteHeapProfile(f); err != nil {
		return err
	}

	return nil
}

func printMemStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Println("\n=== Memory Statistics ===")
	fmt.Printf("Allocated memory: %d KB (%d MB)\n", m.Alloc/1024, m.Alloc/1024/1024)
	fmt.Printf("Total allocated: %d KB (%d MB)\n", m.TotalAlloc/1024, m.TotalAlloc/1024/1024)
	fmt.Printf("System memory: %d KB (%d MB)\n", m.Sys/1024, m.Sys/1024/1024)
	fmt.Printf("Heap allocated: %d KB (%d MB)\n", m.HeapAlloc/1024, m.HeapAlloc/1024/1024)
	fmt.Printf("Heap system: %d KB (%d MB)\n", m.HeapSys/1024, m.HeapSys/1024/1024)
	fmt.Printf("Heap idle: %d KB (%d MB)\n", m.HeapIdle/1024, m.HeapIdle/1024/1024)
	fmt.Printf("Heap in use: %d KB (%d MB)\n", m.HeapInuse/1024, m.HeapInuse/1024/1024)
	fmt.Printf("Heap released: %d KB (%d MB)\n", m.HeapReleased/1024, m.HeapReleased/1024/1024)
	fmt.Printf("Heap objects: %d\n", m.HeapObjects)
	fmt.Printf("GC runs: %d\n", m.NumGC)
	fmt.Printf("GC pause total: %d ns\n", m.PauseTotalNs)
	fmt.Printf("Next GC: %d KB (%d MB)\n", m.NextGC/1024, m.NextGC/1024/1024)
	fmt.Printf("Last GC: %s\n", time.Unix(0, int64(m.LastGC)).Format("2006-01-02 15:04:05"))
}

func printGoroutineInfo() {
	fmt.Println("\n=== Goroutine Information ===")
	fmt.Printf("Number of goroutines: %d\n", runtime.NumGoroutine())
	fmt.Printf("Number of CPUs: %d\n", runtime.NumCPU())
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))

	buf := make([]byte, 1024*1024)
	stackSize := runtime.Stack(buf, true)
	fmt.Printf("\nGoroutine stack trace (first 2048 bytes):\n")
	if stackSize > 2048 {
		fmt.Printf("%s...\n", buf[:2048])
		fmt.Printf("(truncated, total size: %d bytes)\n", stackSize)
	} else {
		fmt.Printf("%s\n", buf[:stackSize])
	}
}

func printBasicInfo() {
	fmt.Println("=== Performance Debug Info ===")
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("CPUs: %d\n", runtime.NumCPU())
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("Goroutines: %d\n", runtime.NumGoroutine())

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Memory allocated: %d KB\n", m.Alloc/1024)
	fmt.Printf("Total allocations: %d KB\n", m.TotalAlloc/1024)
	fmt.Printf("GC runs: %d\n", m.NumGC)

	fmt.Println("\nUse --help to see available profiling options.")
}
