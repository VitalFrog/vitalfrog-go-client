package main

import (
	"encoding/json"
	"fmt"
	"github.com/VitalFrog/vitalfrog-go-client/vfrogapi"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/vitalfrog/termtable"
	"math/rand"
	"os"
	"strings"
	"time"
)

func main() {
	fmt.Println(vitalFrogHeaderText)

	//
	// Parse config struct
	cfg, err := parseEnvironmentToConfig()
	if err != nil {
		log.Fatalf(err.Error())
	}

	//
	// Create new report
	vfAPI := vfrogapi.New(cfg.APIBaseUrl, cfg.APIToken)
	reportConfig := cfg.ToReportConfig()

	metadata, err := vfAPI.CreateReport(reportConfig)
	if err != nil {
		log.Fatalf("Could not CreateReport: %s", err)
	}

	if metadata == nil {
		log.Fatalf("Did get nil metadata as response from VitalFrog API. This is not valid.")
	}

	//
	// Load reports performance budgets for later coloring of the cli
	var performanceBudgets *vfrogapi.PerformanceBudgets
	if metadata.Config.PerformanceBudgetsId != nil {
		performanceBudgets, err = vfAPI.GetPerformanceBudgets(*metadata.Config.PerformanceBudgetsId)
		if err != nil {
			log.Fatalf("could not GetPerformanceBudgets: %s", err)
		}
	}

	//
	// Print basic info
	fmt.Print("\n----------\n")
	fmt.Printf("\nCreated at %s\n", metadata.Created.Format(time.RFC822))
	fmt.Printf("Costs %d tokens\n", metadata.Cost)
	fmt.Printf("Report web url https://app.vitalfrog.com/report/%s\n", metadata.Uuid)
	fmt.Print("\n----------\n")

	//
	// Write performance report table to cli
	// Only write table if we have sync report
	if !cfg.RunAsync {
		tt := termtable.New(os.Stdout, " | ")
		tt.WriteHeader([]termtable.HeaderField{
			{
				Field: termtable.NewStringField("Path"),
			},
			{
				Field: termtable.NewStringField("Country"),
				Width: termtable.IntPointer(4),
			},
			{
				Field: termtable.NewStringField("Device"),
				Width: termtable.IntPointer(10),
			},
			{
				Field: termtable.NewStringField("Max First Input Delay"),
				Width: termtable.IntPointer(10),
			},
			{
				Field: termtable.NewStringField("Server response time"),
				Width: termtable.IntPointer(10),
			},
			{
				Field: termtable.NewStringField("Time to interactive"),
				Width: termtable.IntPointer(10),
			},
			{
				Field: termtable.NewStringField("Cumulative Layout Shift"),
				Width: termtable.IntPointer(40),
			},
			{
				Field: termtable.NewStringField("Largest Contentful Paint"),
				Width: termtable.IntPointer(40),
			},
		})
		tt.WriteRowDivider('=')

		//
		// Get budgets from channel and write them as table rows
		// If highestBudgetLevel is 2, return os.Exit(1). To trigger CI failure
		highestBudgetLevel, err := writeBudgetRows(tt, vfAPI, metadata.Uuid, performanceBudgets)
		if err != nil {
			log.Errorf("could not writeBudgetRows: %s", err)
		}
		defer func(highestBudgetLevel int) {
			switch highestBudgetLevel {
			case 0:
				color.New(color.FgGreen).Print("All metrics are in a good shape. Nothing to do.")
			case 1:
				color.New(color.FgYellow).Print("You have a few metrics which you should look at as they are in the warning state. Please check above table.")
			case 2:
				color.New(color.FgRed).Print("Got at least one metric which is not within an acceptable performance budget. Please check above table. (Marked with '✖')")
				os.Exit(1)
			}
		}(highestBudgetLevel)
	}

	//
	// Write report summary footer
	if jsonConfig, err := json.Marshal(metadata.Config); err == nil {
		fmt.Printf("\n----------\n\nConfig:\n%s\n", string(jsonConfig))
	} else {
		fmt.Printf("\n----------\n\nConfig:\n%+v\n", metadata.Config)
	}

	if performanceBudgets != nil {
		if jsonBudgets, err := json.Marshal(performanceBudgets.Budgets); err == nil {
			fmt.Printf("\nPerformance Budgets:\n%s\n\n----------\n", string(jsonBudgets))
		} else {
			fmt.Printf("\nPerformance Budgets:\n%+v\n\n----------\n", performanceBudgets.Budgets)
		}
	}

}

func writeBudgetRows(tt *termtable.TermTable,
	vfAPI vfrogapi.Client,
	uuid string,
	performanceBudgets *vfrogapi.PerformanceBudgets) (int, error) {
	highestBudgetLevel := 0
	seenReports := map[int32]struct{}{}
	for {
		time.Sleep(time.Duration(rand.Intn(5000-1000)+1000) * time.Millisecond)
		report, err := vfAPI.GetReport(uuid)
		if err != nil {
			return -1, fmt.Errorf("could not GetReport: %w", err)
		}

		for _, report := range report.Data {
			if _, seen := seenReports[report.Id]; seen {
				continue
			}
			seenReports[report.Id] = struct{}{}

			lcp := fmt.Sprintf("%dms", report.LargestContentfulPaint.ValueMs)
			fid := fmt.Sprintf("%dms", report.MaxPotentialFidMs)
			cls := fmt.Sprintf("%f", report.CumulativeLayoutShift.Value)
			serverResponseTime := fmt.Sprintf("%dms", report.ServerResponseTimeMs)
			interactive := fmt.Sprintf("%dms", report.InteractiveMs)

			lcpColor := white
			fidColor := white
			clsColor := white
			serverResponseTimeColor := white
			interactiveColor := white

			for _, budget := range performanceBudgets.Budgets {
				compare := func(value int32) (*color.Color, int) {
					return compareValueAgainstBudget(value, budget)
				}
				var highestBudget int
				switch budget.Metric {
				case vfrogapi.PerformanceBudgetMetricLargestContentfulPaintMs:
					lcpColor, highestBudget = compare(report.LargestContentfulPaint.ValueMs)
					if highestBudget == 2 {
						lcp = fmt.Sprintf("✖ %s", lcp)
					}
				case vfrogapi.PerformanceBudgetMetricMaxPotentialFidMs:
					fidColor, highestBudget = compare(report.MaxPotentialFidMs)
					if highestBudget == 2 {
						fid = fmt.Sprintf("✖ %s", fid)
					}
				case vfrogapi.PerformanceBudgetMetricCumulativeLayoutShift:
					clsColor, highestBudget = compare(int32(report.CumulativeLayoutShift.Value * 100))
					if highestBudget == 2 {
						cls = fmt.Sprintf("✖ %s", cls)
					}
				case vfrogapi.PerformanceBudgetMetricServerResponseTimeMs:
					serverResponseTimeColor, highestBudget = compare(report.ServerResponseTimeMs)
					if highestBudget == 2 {
						serverResponseTime = fmt.Sprintf("✖ %s", serverResponseTime)
					}
				case vfrogapi.PerformanceBudgetMetricInteractiveMs:
					interactiveColor, highestBudget = compare(report.InteractiveMs)
					if highestBudget == 2 {
						interactive = fmt.Sprintf("✖ %s", interactive)
					}
				}

				if highestBudget > highestBudgetLevel {
					highestBudgetLevel = highestBudget
				}
			}

			tt.WriteRow([]termtable.Field{
				termtable.NewStringField(report.Path),
				termtable.NewStringField(report.Country.Code),
				termtable.NewStringField(string(report.Device.Name)),
				termtable.NewColorField(fid, fidColor),
				termtable.NewColorField(serverResponseTime, serverResponseTimeColor),
				termtable.NewColorField(interactive, interactiveColor),
				termtable.NewColorField(cls, clsColor),
				termtable.NewColorField(lcp, lcpColor),
			})

			lcpSelectorElements := strings.Split(report.LargestContentfulPaint.Element.Selector, ">")
			for k, v := range lcpSelectorElements {
				arrow := ">"
				if k == 0 {
					arrow = ""
				}
				lcpSelectorElements[k] = fmt.Sprintf("%s%s%s", termtable.WhiteSpace(k), arrow, strings.TrimSpace(v))
			}
			var clsSelectorElements []string

			if report.CumulativeLayoutShift.Elements != nil {
				for _, el := range *report.CumulativeLayoutShift.Elements {
					elements := strings.Split(el.Selector, ">")
					for k, v := range elements {
						arrow := ">"
						if k == 0 {
							arrow = ""
						}
						clsSelectorElements = append(clsSelectorElements, fmt.Sprintf("%s%s%s", termtable.WhiteSpace(k), arrow, strings.TrimSpace(v)))
					}
				}
			}

			for k := 0; k < maxInt(len(lcpSelectorElements), len(clsSelectorElements)); k++ {
				switch {
				case k < len(lcpSelectorElements) && k < len(clsSelectorElements):
					// Both still have values
					tt.WriteRow([]termtable.Field{
						termtable.NewEmptyField(),
						termtable.NewEmptyField(),
						termtable.NewEmptyField(),
						termtable.NewEmptyField(),
						termtable.NewEmptyField(),
						termtable.NewEmptyField(),
						termtable.NewStringField(clsSelectorElements[k]),
						termtable.NewStringField(lcpSelectorElements[k]),
					})
				case k >= len(lcpSelectorElements) && k < len(clsSelectorElements):
					//CLS still has values
					tt.WriteRow([]termtable.Field{
						termtable.NewEmptyField(),
						termtable.NewEmptyField(),
						termtable.NewEmptyField(),
						termtable.NewEmptyField(),
						termtable.NewEmptyField(),
						termtable.NewEmptyField(),
						termtable.NewStringField(clsSelectorElements[k]),
						termtable.NewEmptyField(),
					})
				case k < len(lcpSelectorElements) && k >= len(clsSelectorElements):
					// LCP still have values
					tt.WriteRow([]termtable.Field{
						termtable.NewEmptyField(),
						termtable.NewEmptyField(),
						termtable.NewEmptyField(),
						termtable.NewEmptyField(),
						termtable.NewEmptyField(),
						termtable.NewEmptyField(),
						termtable.NewEmptyField(),
						termtable.NewStringField(lcpSelectorElements[k]),
					})
				}
			}

			tt.WriteRowDivider('-')

		}
		if report.Metadata.Finished != nil {
			break
		}
	}
	return highestBudgetLevel, nil
}

func maxInt(a, v int) int {
	if a > v {
		return a
	}
	return v
}

var (
	green  = color.New(color.FgGreen)
	yellow = color.New(color.FgYellow)
	red    = color.New(color.FgRed)
	white  = color.New(color.FgWhite)
)

func compareValueAgainstBudget(value int32, budget vfrogapi.PerformanceBudget) (*color.Color, int) {
	if budget.Mode == nil || *budget.Mode == vfrogapi.Above {
		// Check above values
		switch {
		case value < budget.Warning:
			// all alright
			return green, 0
		case value >= budget.Warning && value < budget.Error:
			return yellow, 1
		default:
			return red, 2
		}
	}
	// Check below values
	switch {
	case value > budget.Warning:
		// all alright
		return green, 0
	case value <= budget.Warning && value > budget.Error:
		return yellow, 1
	default:
		return red, 2
	}

}

var vitalFrogHeaderText = color.New(color.FgGreen).SprintFunc()(`                                                                                         
 _|      _|   _|     _|                  _|   _|_|_|_|                                   
 _|      _|        _|_|_|_|     _|_|_|   _|   _|         _|  _|_|     _|_|       _|_|_|  
 _|      _|   _|     _|       _|    _|   _|   _|_|_|     _|_|       _|    _|   _|    _|  
   _|  _|     _|     _|       _|    _|   _|   _|         _|         _|    _|   _|    _|  
     _|       _|       _|_|     _|_|_|   _|   _|         _|           _|_|       _|_|_|  
                                                                                     _|  
                                                                                 _|_|    `)
