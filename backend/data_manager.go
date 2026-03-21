package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"time"
)

const (
	schemesFile  = "data/sif_schemes.json"
	dataFile     = "data/sif_data.csv"
	holidaysFile = "data/holidays.json"
	startDateStr = "2025-09-01"
)

type SchemeMeta struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
}

type NAVRecord struct {
	ID        string
	Date      string
	NAV       float64
	IsPrevDay bool
}

// Holiday structure matching NSE API Response (subset)
type HolidayResp struct {
	CM []struct {
		TradingDate string `json:"tradingDate"` // Format: "dd-MMM-yyyy"
	} `json:"CM"`
}

func updateData(client *http.Client) error {
	// 1. Load Holidays
	holidays, err := loadHolidays()
	if err != nil {
		fmt.Println("Warning: Could not load holidays, assuming no holidays:", err)
		holidays = make(map[string]bool)
	}

	// 2. Load Schemes
	schemes, err := loadSchemes()
	if err != nil {
		schemes = make(map[string]SchemeMeta)
	}

	// 3. Load Existing CSV Data
	// Map: Date -> ID -> Record
	dataMap, err := loadCSV()
	if err != nil {
		// If error, assume empty
		dataMap = make(map[string]map[string]NAVRecord)
	}

	// 4. Determine Date Range
	start, _ := time.Parse("2006-01-02", startDateStr)

	// Find last date in data
	lastDate := start.AddDate(0, 0, -1)
	for dStr := range dataMap {
		d, _ := time.Parse("2006-01-02", dStr)
		if d.After(lastDate) {
			lastDate = d
		}
	}

	// Start from next day
	curr := lastDate.AddDate(0, 0, 1)
	today := time.Now().Truncate(24 * time.Hour) // handled in local time

	fmt.Printf("Updating data from %s to %s\n", curr.Format("2006-01-02"), today.Format("2006-01-02"))

	// 5. Loop through days
	// We need to keep track of "Active Schemes" to copy-forward on holidays
	// Initialize active schemes from the last known data point (if any)
	activeSchemes := make(map[string]float64)
	if !lastDate.Before(start) {
		lastDateStr := lastDate.Format("2006-01-02")
		if records, ok := dataMap[lastDateStr]; ok {
			for id, rec := range records {
				activeSchemes[id] = rec.NAV
			}
		}
	}

	for !curr.After(today) {
		dateStr := curr.Format("2006-01-02")
		fmt.Printf("Processing %s... ", dateStr)

		isWeekend := curr.Weekday() == time.Saturday || curr.Weekday() == time.Sunday
		isHol := holidays[dateStr]

		if isWeekend || isHol {
			fmt.Println("Holiday/Weekend. Copying previous values.")
			// Copy from activeSchemes
			currentRecords := make(map[string]NAVRecord)
			for id, val := range activeSchemes {
				currentRecords[id] = NAVRecord{
					ID:        id,
					Date:      dateStr,
					NAV:       val,
					IsPrevDay: true,
				}
			}
			dataMap[dateStr] = currentRecords
		} else {
			fmt.Println("Trading Day. Fetching API...")
			// Fetch from API
			navs, names, err := fetchNAVData(client, dateStr)
			if err != nil {
				fmt.Printf("Error fetching data for %s: %v. Skipping.\n", dateStr, err)
				// Critical logic decision: If API fails, do we skip or copy?
				// For now, let's just skip and retry next run.
			} else {
				currentRecords := make(map[string]NAVRecord)
				for id, val := range navs {
					// Update Active Schemes
					activeSchemes[id] = val
					currentRecords[id] = NAVRecord{
						ID:        id,
						Date:      dateStr,
						NAV:       val,
						IsPrevDay: false,
					}

					// Check New Scheme
					if _, exists := schemes[id]; !exists {
						name := names[id]
						// Apply cleaning
						name = cleanName(name)
						if !shouldSkip(name) {
							schemes[id] = SchemeMeta{
								ID:        id,
								Name:      name,
								StartDate: dateStr,
							}
							fmt.Printf("  [NEW] Found Scheme: %s\n", name)
						}
					}
				}
				dataMap[dateStr] = currentRecords
			}
			// Rate limit
			time.Sleep(200 * time.Millisecond)
		}

		curr = curr.AddDate(0, 0, 1)
	}

	// 6. Save Data
	if err := saveSchemes(schemes); err != nil {
		return err
	}
	if err := saveCSV(dataMap); err != nil {
		return err
	}

	fmt.Println("Update Complete.")
	return nil
}

// Persist Logic

func loadSchemes() (map[string]SchemeMeta, error) {
	f, err := os.Open(schemesFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var list []SchemeMeta
	if err := json.NewDecoder(f).Decode(&list); err != nil {
		return nil, err
	}

	m := make(map[string]SchemeMeta)
	for _, s := range list {
		m[s.ID] = s
	}
	return m, nil
}

func saveSchemes(m map[string]SchemeMeta) error {
	var list []SchemeMeta
	for _, s := range m {
		list = append(list, s)
	}
	// Sort for stability
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })

	b, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(schemesFile, b, 0644)
}

func loadCSV() (map[string]map[string]NAVRecord, error) {
	f, err := os.Open(dataFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	data := make(map[string]map[string]NAVRecord)
	// Header: sid_id, date, nav, is_previous_day_nav
	if len(rows) < 1 {
		return data, nil
	}

	for i, row := range rows {
		if i == 0 {
			continue
		} // Header
		if len(row) < 4 {
			continue
		}

		id := row[0]
		date := row[1]
		nav := 0.0
		fmt.Sscanf(row[2], "%f", &nav)
		isPrev := (row[3] == "Y")

		if data[date] == nil {
			data[date] = make(map[string]NAVRecord)
		}
		data[date][id] = NAVRecord{
			ID:        id,
			Date:      date,
			NAV:       nav,
			IsPrevDay: isPrev,
		}
	}
	return data, nil
}

func saveCSV(data map[string]map[string]NAVRecord) error {
	f, err := os.Create(dataFile)
	if err != nil {
		return err
	}
	defer f.Close()

	var dates []string
	for d := range data {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	w := csv.NewWriter(f)
	w.Write([]string{"sid_id", "date", "nav", "is_previous_day_nav"})

	for _, d := range dates {
		records := data[d]
		// Sort records by ID for stability
		var ids []string
		for id := range records {
			ids = append(ids, id)
		}
		sort.Strings(ids)

		for _, id := range ids {
			rec := records[id]
			prevFlag := ""
			if rec.IsPrevDay {
				prevFlag = "Y"
			}
			w.Write([]string{
				rec.ID,
				rec.Date,
				fmt.Sprintf("%.4f", rec.NAV),
				prevFlag,
			})
		}
	}
	w.Flush()
	return nil
}

func loadHolidays() (map[string]bool, error) {
	f, err := os.Open(holidaysFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var h HolidayResp
	if err := json.NewDecoder(f).Decode(&h); err != nil {
		return nil, err
	}

	holidayMap := make(map[string]bool)
	for _, day := range h.CM {
		// API Date format: "15-Jan-2026"
		t, err := time.Parse("02-Jan-2006", day.TradingDate)
		if err == nil {
			holidayMap[t.Format("2006-01-02")] = true
		}
	}
	return holidayMap, nil
}
