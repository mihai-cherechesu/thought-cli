package pkg

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

type ServiceStatus int64

// Set the number of workers as the number of processors available on
// the running machine.
var (
	workers = runtime.NumCPU()
)

const (
	Healthy ServiceStatus = iota
	Unhealthy
)

// Merged row in an output table from ls command (when --merged flag is specified).
// All servers that pertain to a single service are aggregated under
// that specific service.
// Cpu and Memory metrics are summed for all instances of the service and returned as
// an average.
type MergedLsRow struct {
	Ips      []string
	Name     string
	CpuAvg   int32
	MemAvg   int32
	Replicas int32
}

// Default row in an output table from ls command.
type DefaultLsRow struct {
	Ip     string
	Name   string
	Status ServiceStatus
	Cpu    string
	Mem    string
}

type RunOutput struct {
	Response ServiceResponse
	Ip       string
}

// Checks if cpu and memory are above some thresholds and returns true if so.
// Current thresholds are 90% for cpu and 80% for memory.
func isUnhealthy(cpu string, mem string) ServiceStatus {
	parsedCpu, err := strconv.Atoi(strings.Split(cpu, "%")[0])
	if err != nil {
		log.Fatalf("could not parse cpu percentage")
	}

	parsedMem, err := strconv.Atoi(strings.Split(mem, "%")[0])
	if err != nil {
		log.Fatalf("could not parse mem percentage")
	}

	status := Healthy
	if parsedCpu > 90 || parsedMem > 80 {
		status = Unhealthy
	}

	return status
}

func RunLs(follow bool, service string, merged bool) {
	// Fetch all servers.
	// Currently CPX maps all servers to a single service (shouldn't be the case in a
	// real world scenario as there could be multiple services running on the same server)
	servers := GetServers()

	bar := progressbar.NewOptions(len(servers),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(25),
		progressbar.OptionSetDescription("[cyan][1/1][reset] Fetching info for servers..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	// Create a channel to receive the stage names that finished
	// Buffered with a maximum capacity of len(servers)*
	doneCh := make(chan RunOutput, len(servers))
	numDone := 0

	// Set up the work pool
	work := make(chan string, len(servers))

	// Start the workers
	for i := 0; i < workers; i++ {

		go func() {
			for ip := range work {
				s := GetService(ip)
				doneCh <- RunOutput{Response: s, Ip: ip}
			}
		}()
	}

	// Queue up the work
	for _, server := range servers {
		work <- server
	}

	// Close the work channel to signal to the workers that they're done
	close(work)

	// Store each response in a map for faster retrieval.
	// We index the rows by their service name.
	// The value could be either MergedLsRow or []DefaultLsRow.
	rows := make(map[string]interface{})

	// Loop indefinitely and wait for responses on the channel
	for {
		select {
		case s := <-doneCh:
			numDone++
			bar.Add(1)

			_, ok := rows[s.Response.Service]

			if ok && merged {
				parsedCpu, err := strconv.Atoi(strings.Split(s.Response.Cpu, "%")[0])
				if err != nil {
					log.Fatalf("could not parse cpu percentage")
				}

				parsedMem, err := strconv.Atoi(strings.Split(s.Response.Memory, "%")[0])
				if err != nil {
					log.Fatalf("could not parse mem percentage")
				}

				mergedRow, _ := rows[s.Response.Service].(MergedLsRow)
				// In order to keep things simple, we always rebuild the averages as:
				// ((old_avg * old_num_ips) + curr_val) / new_num_ips
				mergedRow.CpuAvg *= int32(len(mergedRow.Ips))
				mergedRow.MemAvg *= int32(len(mergedRow.Ips))
				mergedRow.Ips = append(mergedRow.Ips, s.Ip)
				mergedRow.Replicas++
				mergedRow.CpuAvg += int32(parsedCpu)
				mergedRow.MemAvg += int32(parsedMem)
				mergedRow.CpuAvg /= int32(len(mergedRow.Ips))
				mergedRow.MemAvg /= int32(len(mergedRow.Ips))
				rows[s.Response.Service] = mergedRow

			} else if ok && !merged {
				defaultRow := DefaultLsRow{
					Ip:     s.Ip,
					Name:   s.Response.Service,
					Status: isUnhealthy(s.Response.Cpu, s.Response.Memory),
					Cpu:    s.Response.Cpu,
					Mem:    s.Response.Memory,
				}
				defaultRows, _ := rows[s.Response.Service].([]DefaultLsRow)
				defaultRows = append(defaultRows, defaultRow)
				rows[s.Response.Service] = defaultRows

			} else if !ok && merged {
				parsedCpu, err := strconv.Atoi(strings.Split(s.Response.Cpu, "%")[0])
				if err != nil {
					log.Fatalf("could not parse cpu percentage")
				}

				parsedMem, err := strconv.Atoi(strings.Split(s.Response.Memory, "%")[0])
				if err != nil {
					log.Fatalf("could not parse mem percentage")
				}

				mergedRow := MergedLsRow{
					Ips:      []string{s.Ip},
					Name:     s.Response.Service,
					CpuAvg:   int32(parsedCpu),
					MemAvg:   int32(parsedMem),
					Replicas: 1,
				}
				rows[s.Response.Service] = mergedRow

				// !ok && !merged
			} else {
				defaultRow := DefaultLsRow{
					Ip:     s.Ip,
					Name:   s.Response.Service,
					Status: isUnhealthy(s.Response.Cpu, s.Response.Memory),
					Cpu:    s.Response.Cpu,
					Mem:    s.Response.Memory,
				}
				rows[s.Response.Service] = []DefaultLsRow{defaultRow}
			}

		default:
			// All work is finished
			if numDone == len(servers) {
				fmt.Println()
				// Render the tables
				if follow {
					RenderOutputFollow(rows[service])
				} else {
					RenderOutput(rows, service)
				}

				// Exit clean with 0
				os.Exit(0)
			}

			// Wait half a second if nothing is received on channel
			time.Sleep(300 * time.Millisecond)
		}
	}
}
