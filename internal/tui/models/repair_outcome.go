package models

import "github.com/petergi/ebook-mechanic-lib/pkg/ebmlib"

// RepairOutcome bundles a repair result with optional validation output.
type RepairOutcome struct {
	Result *ebmlib.RepairResult
	Report *ebmlib.ValidationReport
}
