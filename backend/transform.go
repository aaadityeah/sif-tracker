package main

import "strings"

func cleanName(name string) string {
	// Name Cleaning
	reshaped := name

	// Case-insensitive replacement would be more robust but matching user request specific patterns
	reshaped = strings.ReplaceAll(reshaped, "Long-Short Fund", "LSF")
	reshaped = strings.ReplaceAll(reshaped, "Long Short Fund", "LSF")

	// Remove "Plan" from "Growth Plan" -> "Growth"
	// Remove "Option" from "Growth Option" -> "Growth"
	reshaped = strings.ReplaceAll(reshaped, "Growth Plan", "Growth")
	reshaped = strings.ReplaceAll(reshaped, "Growth Option", "Growth")

	// Remove "Plan" from "Direct Plan" -> "Direct"
	reshaped = strings.ReplaceAll(reshaped, "Direct Plan", "Direct")
	reshaped = strings.ReplaceAll(reshaped, "Direct  Plan", "Direct")

	return reshaped
}

func shouldSkip(name string) bool {
	lowerName := strings.ToLower(name)
	// Filter out "Regular" and "IDCW" and long-form IDCW
	if strings.Contains(lowerName, "regular") ||
		strings.Contains(lowerName, "idcw") ||
		strings.Contains(lowerName, "income distribution cum capital withdrawal") {
		return true
	}
	return false
}

func calcChange(current, previous float64) float64 {
	if previous == 0 {
		return 0
	}
	return ((current - previous) / previous) * 100
}
