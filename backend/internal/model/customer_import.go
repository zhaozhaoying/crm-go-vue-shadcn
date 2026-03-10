package model

type CustomerImportError struct {
	Row    int    `json:"row"`
	Name   string `json:"name,omitempty"`
	Phone  string `json:"phone,omitempty"`
	Reason string `json:"reason"`
}

type CustomerImportResult struct {
	TotalRows   int                   `json:"totalRows"`
	SuccessRows int                   `json:"successRows"`
	SkippedRows int                   `json:"skippedRows"`
	FailedRows  int                   `json:"failedRows"`
	DryRun      bool                  `json:"dryRun"`
	Errors      []CustomerImportError `json:"errors,omitempty"`
}
