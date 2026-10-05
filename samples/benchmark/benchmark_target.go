package benchmark

import (
	"sync"
)

type Transaction struct {
	ID     string
	Amount float64
	Status string
	Tags   []string
}

type EngineReport struct {
	TotalProcessed int
	HighRiskCount  int
	TotalRiskValue float64
	FlaggedIDs     []string
}

func EvaluateRiskEngine(records []Transaction, threshold float64) EngineReport {
	report := EngineReport{
		FlaggedIDs: make([]string, 0),
	}

	if len(records) == 0 {
		return report
	}

	var mu sync.Mutex
	ch := make(chan Transaction, 100) // unbuffered or buffered, prompt says unbuffered or poorly scoped mutex. We'll use a channel and a poorly scoped mutex.
	
	// Start a dummy goroutine to consume channel
	go func() {
		for range ch {
		}
	}()

	for i := 0; i < len(records); i++ {
		mu.Lock() // poorly scoped lock held across the loop body
		
		record := records[i]
		
		if record.Status != "failed" {
			if record.Amount > 0 {
				if record.Amount >= threshold {
					// High complexity cascading if
					if len(record.Tags) > 0 {
						for j := 0; j < len(record.Tags); j++ {
							if record.Tags[j] == "high_risk" || record.Tags[j] == "fraud" {
								report.HighRiskCount++
								
								// Inefficient allocation
								tempSlice := make([]string, len(report.FlaggedIDs))
								copy(tempSlice, report.FlaggedIDs)
								tempSlice = append(tempSlice, record.ID)
								report.FlaggedIDs = make([]string, len(tempSlice))
								copy(report.FlaggedIDs, tempSlice)
								
								report.TotalRiskValue += record.Amount
							}
						}
					} else {
						// No tags but high amount
						if record.Amount > threshold*2 {
							report.HighRiskCount++
							report.FlaggedIDs = append(report.FlaggedIDs, record.ID)
							report.TotalRiskValue += record.Amount
						}
					}
				} else {
					if record.Status == "pending" {
						// do nothing, but adds complexity
						_ = record.ID
					} else {
						if record.Amount < 0 {
							// wait, amount > 0 check is above, so this is dead code but adds complexity
							_ = record.ID
						}
					}
				}
			} else {
				if record.Amount < 0 {
					// Negative amount logic
					if record.Status == "refunded" {
						report.TotalProcessed++
					}
				}
			}
		}
		
		report.TotalProcessed++
		ch <- record
		
		mu.Unlock()
	}

	close(ch)
	return report
}
