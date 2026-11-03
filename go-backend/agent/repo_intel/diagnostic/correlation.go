package diagnostic

import (
	"fmt"
	"sort"
	"strings"
)

// FindingCorrelator clusters related findings, identifies parent-child hierarchies,
// and identifies duplicate or coordinated remediation needs.
type FindingCorrelator struct{}

// NewFindingCorrelator creates a new FindingCorrelator.
func NewFindingCorrelator() *FindingCorrelator {
	return &FindingCorrelator{}
}

// CorrelateFindings analyzes findings to find shared locations, symbols, and root causes.
func (fc *FindingCorrelator) CorrelateFindings(findings []ValidatedFinding) ([]FindingCorrelation, map[string][]string) {
	var correlations []FindingCorrelation
	relatedMap := make(map[string][]string)

	// 1. Group by file and symbol for Parent-Child / Shared Cause AST clustering
	type symbolKey struct {
		file   string
		symbol string
	}
	symbolGroups := make(map[symbolKey][]string)

	// 2. Group by file for co-located findings
	fileGroups := make(map[string][]string)

	// 3. Exact evidence duplication detection
	type evidenceKey struct {
		file      string
		lineStart int
		category  string
	}
	evidenceGroups := make(map[evidenceKey][]string)

	for _, f := range findings {
		for _, ev := range f.Evidence {
			k := evidenceKey{
				file:      ev.File,
				lineStart: ev.LineStart,
				category:  string(f.Category),
			}
			evidenceGroups[k] = append(evidenceGroups[k], f.FindingID)

			fileGroups[ev.File] = append(fileGroups[ev.File], f.FindingID)
		}

		if len(f.AffectedComponents) > 0 {
			file := f.AffectedComponents[0]
			fileGroups[file] = append(fileGroups[file], f.FindingID)

			// Check symbol
			for _, sym := range extractSymbolsFromID(f.FindingID) {
				sk := symbolKey{file: file, symbol: sym}
				symbolGroups[sk] = append(symbolGroups[sk], f.FindingID)
			}
		}
	}

	corrIdx := 1

	// Check duplicates (exact same file, line, and category)
	for _, ids := range evidenceGroups {
		dedupedIDs := uniqueStrings(ids)
		if len(dedupedIDs) > 1 {
			primary := dedupedIDs[0]
			related := dedupedIDs[1:]
			corr := FindingCorrelation{
				CorrelationID:     fmt.Sprintf("CORR-DUP-%04d", corrIdx),
				Type:              CorrelationDuplicate,
				PrimaryFindingID:  primary,
				RelatedFindingIDs: related,
				Rationale:         fmt.Sprintf("Multiple findings targeting identical file location and category"),
			}
			correlations = append(correlations, corr)
			corrIdx++

			for _, r := range related {
				relatedMap[primary] = append(relatedMap[primary], r)
				relatedMap[r] = append(relatedMap[r], primary)
			}
		}
	}

	// Check Parent-Child (same symbol with multiple structural symptoms: nesting + complexity + volume)
	for sk, ids := range symbolGroups {
		dedupedIDs := uniqueStrings(ids)
		if len(dedupedIDs) >= 2 {
			// Find primary (e.g. HIGH-COMPLEXITY or LARGE-AST-VOLUME as parent)
			sort.Slice(dedupedIDs, func(i, j int) bool {
				return strings.Contains(dedupedIDs[i], "COMPLEXITY") || strings.Contains(dedupedIDs[i], "VOLUME")
			})
			primary := dedupedIDs[0]
			related := dedupedIDs[1:]

			corr := FindingCorrelation{
				CorrelationID:     fmt.Sprintf("CORR-PARENT-%04d", corrIdx),
				Type:              CorrelationParentChild,
				PrimaryFindingID:  primary,
				RelatedFindingIDs: related,
				Rationale:         fmt.Sprintf("Function %q in %s exhibits compound structural complexity symptoms (nesting, cyclomatic complexity, volume) stemming from monolithic design", sk.symbol, sk.file),
			}
			correlations = append(correlations, corr)
			corrIdx++

			for _, r := range related {
				relatedMap[primary] = append(relatedMap[primary], r)
				relatedMap[r] = append(relatedMap[r], primary)
			}
		}
	}

	// Check Coordinated Resolution for Architectural Cycles
	var cycleFindingIDs []string
	for _, f := range findings {
		if strings.Contains(f.FindingID, "CYCLE") {
			cycleFindingIDs = append(cycleFindingIDs, f.FindingID)
		}
	}
	if len(cycleFindingIDs) >= 2 {
		primary := cycleFindingIDs[0]
		related := cycleFindingIDs[1:]
		corr := FindingCorrelation{
			CorrelationID:     fmt.Sprintf("CORR-COORD-%04d", corrIdx),
			Type:              CorrelationCoordinatedResolution,
			PrimaryFindingID:  primary,
			RelatedFindingIDs: related,
			Rationale:         "Mutually dependent architectural cycle components requiring coordinated interface extraction",
		}
		correlations = append(correlations, corr)
		corrIdx++

		for _, r := range related {
			relatedMap[primary] = append(relatedMap[primary], r)
			relatedMap[r] = append(relatedMap[r], primary)
		}
	}

	return correlations, relatedMap
}

func extractSymbolsFromID(id string) []string {
	parts := strings.Split(id, "-")
	var symbols []string
	for i, p := range parts {
		if (p == "NEST" || p == "COMPLEX" || p == "VOLUME" || p == "HIGH_NESTING" || p == "HIGH_COMPLEXITY") && i+2 < len(parts) {
			symbols = append(symbols, parts[i+2])
		}
	}
	return symbols
}

func uniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	var res []string
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			res = append(res, s)
		}
	}
	return res
}
