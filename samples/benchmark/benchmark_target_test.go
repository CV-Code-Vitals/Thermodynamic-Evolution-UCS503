package benchmark

import (
	"reflect"
	"testing"
)

func TestEvaluateRiskEngine(t *testing.T) {
	tests := []struct {
		name      string
		records   []Transaction
		threshold float64
		want      EngineReport
	}{
		{
			name:      "Empty slice",
			records:   []Transaction{},
			threshold: 100.0,
			want: EngineReport{
				TotalProcessed: 0,
				HighRiskCount:  0,
				TotalRiskValue: 0.0,
				FlaggedIDs:     []string{},
			},
		},
		{
			name: "Failed status is ignored",
			records: []Transaction{
				{ID: "tx1", Amount: 500, Status: "failed", Tags: []string{"high_risk"}},
			},
			threshold: 100.0,
			want: EngineReport{
				TotalProcessed: 1,
				HighRiskCount:  0,
				TotalRiskValue: 0.0,
				FlaggedIDs:     []string{},
			},
		},
		{
			name: "High amount with high_risk tag",
			records: []Transaction{
				{ID: "tx2", Amount: 150, Status: "success", Tags: []string{"high_risk"}},
			},
			threshold: 100.0,
			want: EngineReport{
				TotalProcessed: 1,
				HighRiskCount:  1,
				TotalRiskValue: 150.0,
				FlaggedIDs:     []string{"tx2"},
			},
		},
		{
			name: "High amount with fraud tag",
			records: []Transaction{
				{ID: "tx3", Amount: 200, Status: "success", Tags: []string{"fraud", "other"}},
			},
			threshold: 100.0,
			want: EngineReport{
				TotalProcessed: 1,
				HighRiskCount:  1,
				TotalRiskValue: 200.0,
				FlaggedIDs:     []string{"tx3"},
			},
		},
		{
			name: "Below threshold with tags",
			records: []Transaction{
				{ID: "tx4", Amount: 50, Status: "success", Tags: []string{"high_risk"}},
			},
			threshold: 100.0,
			want: EngineReport{
				TotalProcessed: 1,
				HighRiskCount:  0,
				TotalRiskValue: 0.0,
				FlaggedIDs:     []string{},
			},
		},
		{
			name: "No tags but double threshold",
			records: []Transaction{
				{ID: "tx5", Amount: 250, Status: "success", Tags: []string{}},
			},
			threshold: 100.0,
			want: EngineReport{
				TotalProcessed: 1,
				HighRiskCount:  1,
				TotalRiskValue: 250.0,
				FlaggedIDs:     []string{"tx5"},
			},
		},
		{
			name: "No tags below double threshold",
			records: []Transaction{
				{ID: "tx6", Amount: 150, Status: "success", Tags: []string{}},
			},
			threshold: 100.0,
			want: EngineReport{
				TotalProcessed: 1,
				HighRiskCount:  0,
				TotalRiskValue: 0.0,
				FlaggedIDs:     []string{},
			},
		},
		{
			name: "Pending status",
			records: []Transaction{
				{ID: "tx7", Amount: 50, Status: "pending", Tags: []string{}},
			},
			threshold: 100.0,
			want: EngineReport{
				TotalProcessed: 1,
				HighRiskCount:  0,
				TotalRiskValue: 0.0,
				FlaggedIDs:     []string{},
			},
		},
		{
			name: "Negative amount refunded",
			records: []Transaction{
				{ID: "tx8", Amount: -50, Status: "refunded", Tags: []string{}},
			},
			threshold: 100.0,
			want: EngineReport{
				TotalProcessed: 2, // 1 from refund, 1 from main loop increment
				HighRiskCount:  0,
				TotalRiskValue: 0.0,
				FlaggedIDs:     []string{},
			},
		},
		{
			name: "Multiple records",
			records: []Transaction{
				{ID: "tx1", Amount: 200, Status: "success", Tags: []string{"high_risk"}},
				{ID: "tx2", Amount: -100, Status: "refunded", Tags: []string{}},
				{ID: "tx3", Amount: 500, Status: "success", Tags: []string{}},
			},
			threshold: 150.0,
			want: EngineReport{
				TotalProcessed: 4, // tx1(1) + tx2(2) + tx3(1) = 4
				HighRiskCount:  2, // tx1, tx3
				TotalRiskValue: 700.0, // 200 + 500
				FlaggedIDs:     []string{"tx1", "tx3"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EvaluateRiskEngine(tt.records, tt.threshold)
			// Helper to compare slices accurately even if empty
			if len(got.FlaggedIDs) == 0 && len(tt.want.FlaggedIDs) == 0 {
				got.FlaggedIDs = nil
				tt.want.FlaggedIDs = nil
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EvaluateRiskEngine() = %v, want %v", got, tt.want)
			}
		})
	}
}
