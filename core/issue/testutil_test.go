package issue

import "time"

// timePtr is a shared test helper function to create time.Time pointers for tests.
// This helper is used across all safe accessor tests (issue, worklog, comment, attachment).
func timePtr(year, month, day, hour, min, sec int) *time.Time {
	t := time.Date(year, time.Month(month), day, hour, min, sec, 0, time.UTC)
	return &t
}
