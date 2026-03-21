package main

import (
	"fmt"
	"os"
	"text/tabwriter"
)

// Result struct for our processed data
type SIFPerformance struct {
	Name     string
	Scheme   string // Parent scheme name if needed
	NavID    string
	Current  float64
	OneDay   *float64
	OneWeek  *float64
	OneMonth *float64
	SixMonth *float64
}

func printTable(perfs []SIFPerformance) {
	fmt.Println("")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Scheme Name\tNAV\t1D %\t1W %\t1M %\t6M %")
	fmt.Fprintln(w, "-----------\t---\t----\t----\t----\t----")

	for _, p := range perfs {
		fmt.Fprintf(w, "%s\t%.4f\t%s\t%s\t%s\t%s\n",
			truncate(p.Name, 50),
			p.Current,
			formatPct(p.OneDay),
			formatPct(p.OneWeek),
			formatPct(p.OneMonth),
			formatPct(p.SixMonth),
		)
	}
	w.Flush()
}

func formatPct(val *float64) string {
	if val == nil {
		return "-"
	}
	color := ""
	reset := ""

	// Simple ANSI coloring
	if *val > 0 {
		color = "\033[32m" // Green
		reset = "\033[0m"
	} else if *val < 0 {
		color = "\033[31m" // Red
		reset = "\033[0m"
	}

	return fmt.Sprintf("%s%.2f%%%s", color, *val, reset)
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}
