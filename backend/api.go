package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// API constants
const apiURL = "https://www.amfiindia.com/api/sif-nav-history?query_type=all_for_date&from_date=%s"

// Data structures matching the API response
type APIResponse struct {
	Data []FundHouse `json:"data"`
}

type FundHouse struct {
	MFName  string   `json:"mfName"`
	Schemes []Scheme `json:"schemes"`
}

type Scheme struct {
	SchemeName string `json:"schemeName"`
	Navs       []NAV  `json:"navs"`
}

type NAV struct {
	SD_ID        string `json:"SD_ID"`
	NAV_Name     string `json:"NAV_Name"`
	HNAV_Amt     string `json:"hNAV_Amt"`
	ISIN_RI      string `json:"ISIN_RI"`
	ISIN_PO      string `json:"ISIN_PO"`
	HNAV_Date    string `json:"hNAV_Date"`
	HNAV_Dtstamp string `json:"hNAV_Dtstamp"`
}

// Helper to retrieve data
func fetchNAVData(client *http.Client, dateStr string) (map[string]float64, map[string]string, error) {
	url := fmt.Sprintf(apiURL, dateStr)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.amfiindia.com/sif/latest-nav/nav-history")

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, nil, err
	}

	navMap := make(map[string]float64)
	nameMap := make(map[string]string)

	for _, fh := range apiResp.Data {
		for _, scheme := range fh.Schemes {
			for _, nav := range scheme.Navs {
				val, err := strconv.ParseFloat(nav.HNAV_Amt, 64)
				if err == nil {
					navMap[nav.SD_ID] = val
					nameMap[nav.SD_ID] = nav.NAV_Name
				}
			}
		}
	}

	// If no data, return error to indicate empty for this date
	if len(navMap) == 0 {
		return nil, nil, fmt.Errorf("no data")
	}

	return navMap, nameMap, nil
}

// Map to store date -> SD_ID -> NAV Float value
type DateData map[string]float64

// Map to store ID -> Name (from the latest available data)
type IDNameMap map[string]string

func getTargetDates(anchorDate time.Time) map[string]time.Time {
	return map[string]time.Time{
		"current": anchorDate,
		"1d":      anchorDate.AddDate(0, 0, -1),
		"1w":      anchorDate.AddDate(0, 0, -7),
		"1m":      anchorDate.AddDate(0, -1, 0),
		"6m":      anchorDate.AddDate(0, -6, 0),
	}
}
