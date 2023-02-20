package pkg

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bcicen/termui"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

var (
	statusMap = map[ServiceStatus]string{
		Healthy:   "Healthy",
		Unhealthy: "Unhealthy",
	}
	updateInterval = 1 * time.Second
)

func RenderOutput(rows map[string]interface{}, service string) {
	t := table.NewWriter()
	t.SetColumnConfigs([]table.ColumnConfig{})
	t.SetStyle(table.StyleBold)
	t.SetRowPainter(table.RowPainter(func(row table.Row) text.Colors {
		if row[4].(string) == statusMap[Unhealthy] {
			return text.Colors{text.BgRed, text.FgBlack}
		}
		return text.Colors{}
	}))

	init := false // helps in not type casting again

	for k, v := range rows {
		// Filtering out the rows for different services
		// Kinda innefficient, but works for this mode
		if service != "" && k != service {
			continue
		}

		switch r := v.(type) {
		case []DefaultLsRow:
			if !init {
				t.AppendHeader(table.Row{"IP", "Service", "Cpu", "Memory", "Status"})
				init = true
			}

			for _, info := range r {
				t.AppendRow(table.Row{info.Ip, info.Name, info.Cpu, info.Mem, statusMap[info.Status]})
			}

		case MergedLsRow:
			if !init {
				t.AppendHeader(table.Row{"IPs", "Service", "Cpu_Avg", "Memory_Avg", "Replicas"})
				t.Style().Options.SeparateRows = true
				init = true
			}
			t.AppendRow(table.Row{strings.Join(r.Ips, "\n"), r.Name,
				strconv.FormatInt(int64(r.CpuAvg), 10) + "%",
				strconv.FormatInt(int64(r.MemAvg), 10) + "%",
				strconv.FormatInt(int64(r.Replicas), 10)},
			)
		}
	}

	fmt.Println(t.Render())
}

func RenderOutputFollow(row interface{}) {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()
	init := false // helps in not type casting again

	table := widgets.NewTable()

	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.SetRect(0, 0, 80, 100)

	switch r := row.(type) {
	case []DefaultLsRow:
		if !init {
			table.Rows = [][]string{{"IP", "Service", "Cpu", "Memory", "Status"}}
			init = true
		}

		// Create a slice of ips to pass it to the update function.
		// It is used to query further only the ips for that service.
		var ips []string
		// Index to query faster the rows based on ip.
		index := make(map[string]int)

		for _, info := range r {
			index[info.Ip] = len(table.Rows)
			table.Rows = append(table.Rows, []string{info.Ip, info.Name, info.Cpu, info.Mem, statusMap[info.Status]})
			ips = append(ips, info.Ip)
		}

		// Update the widget with new values on update interval
		go func() {
			for range time.NewTicker(updateInterval).C {
				updateDefault(table.Rows, ips, index)
				ui.Render(table)
			}
		}()

	case MergedLsRow:
		if !init {
			table.Rows = [][]string{{"IPs", "Service", "Cpu_Avg", "Memory_Avg", "Replicas"}}
			init = true
		}
		table.Rows = append(table.Rows, []string{strings.Join(r.Ips, ", "), r.Name,
			strconv.FormatInt(int64(r.CpuAvg), 10) + "%",
			strconv.FormatInt(int64(r.MemAvg), 10) + "%",
			strconv.FormatInt(int64(r.Replicas), 10)},
		)

		err := termui.Init()
		if err != nil {
			panic(err)
		}
		defer termui.Close()

		dataPoints := 80
		lc0 := termui.NewLineChart()
		lc0.BorderLabel = "CPU Average for " + r.Name
		lc0.Data = make([]float64, dataPoints)
		lc0.Width = 50
		lc0.Height = 12
		lc0.X = 80
		lc0.Y = 0
		lc0.AxesColor = termui.ColorWhite
		lc0.LineColor = termui.ColorGreen | termui.AttrBold

		lc1 := termui.NewLineChart()
		lc1.BorderLabel = "Memory Average for " + r.Name
		lc1.Data = make([]float64, dataPoints)
		lc1.Width = 50
		lc1.Height = 12
		lc1.X = 80
		lc1.Y = 12
		lc1.AxesColor = termui.ColorWhite
		lc1.LineColor = termui.ColorGreen | termui.AttrBold

		termui.Render(lc0, lc1)

		// Update the widget with new values on upd	ate interval
		go func() {
			lineIndex := 0

			for range time.NewTicker(updateInterval).C {
				updateMerged(table.Rows, r.Ips, lc0.Data, lc1.Data, lineIndex%dataPoints)
				ui.Render(table)
				termui.Render(lc0, lc1)
				lineIndex++
			}
		}()
	}

	ui.Render(table)
	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		}
	}
}

func updateDefault(rows [][]string, ips []string, index map[string]int) {
	for _, ip := range ips {
		s := GetService(ip)
		rowIndex := index[ip]
		rows[rowIndex][2] = s.Cpu
		rows[rowIndex][3] = s.Memory
		rows[rowIndex][4] = statusMap[isUnhealthy(s.Cpu, s.Memory)]
	}
}

func updateMerged(rows [][]string, ips []string, cpu []float64, mem []float64, lineIndex int) {
	cpuSum := 0
	memSum := 0

	for _, ip := range ips {
		s := GetService(ip)
		parsedCpu, err := strconv.Atoi(strings.Split(s.Cpu, "%")[0])
		if err != nil {
			log.Fatalf("could not parse cpu percentage")
		}

		parsedMem, err := strconv.Atoi(strings.Split(s.Memory, "%")[0])
		if err != nil {
			log.Fatalf("could not parse mem percentage")
		}

		cpuSum += parsedCpu
		memSum += parsedMem
	}

	cpuAvg := cpuSum / len(ips)
	memAvg := memSum / len(ips)

	rows[1][2] = strconv.FormatInt(int64(cpuAvg), 10) + "%"
	rows[1][3] = strconv.FormatInt(int64(memAvg), 10) + "%"

	cpu[lineIndex] = float64(cpuAvg)
	mem[lineIndex] = float64(memAvg)
}
