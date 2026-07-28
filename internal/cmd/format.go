package cmd

import "time"

// formatTime renders Linear's ISO timestamps as local "2006-01-02 15:04".
func formatTime(iso string) string {
	if iso == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return iso
	}
	return t.Local().Format("2006-01-02 15:04")
}
