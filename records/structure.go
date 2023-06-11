package records

import "time"

type AllRecords struct {
	Scans map[string]ScanRecords `json:"scans"`
	Diffs map[string]DiffRecords `json:"diffs"`
}

type ScanRecords struct {
	Records     []Record `json:"records"`
	CurrScanNum int      `json:"currScanNum"`
}

type DiffRecords struct {
	Records []Record `json:"records"`
}

type Record struct {
	IsComprehensive bool
	TimeCompleted   time.Time
}
