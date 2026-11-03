package benchmark

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// GenerateBenchmarkSuiteV3 generates the 144-fixture expanded benchmark suite v3 and its ground truth manifest.
func GenerateBenchmarkSuiteV3(baseDir string, manifestPath string) (*GroundTruthManifest, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("create baseDir: %w", err)
	}

	manifest := &GroundTruthManifest{
		BenchmarkName:       "Self-Healing Autonomous Benchmark Suite v3 (Independent Multi-Category)",
		Version:             "3.0.0",
		CreatedAt:           time.Now(),
		TargetRepoPath:      filepath.ToSlash(baseDir),
		DefectFixturesCount: 96,
		CleanFixturesCount:  48,
		TotalFixtures:       144,
		DefectFixtures:      make([]DefectFixture, 0, 96),
		CleanFixtures:       make([]CleanFixture, 0, 48),
	}

	writeFile := func(relPath, content string) (string, error) {
		full := filepath.Join(baseDir, filepath.FromSlash(relPath))
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			return "", err
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			return "", err
		}
		h := sha256.Sum256([]byte(content))
		return hex.EncodeToString(h[:]), nil
	}

	// Root go.mod
	_, err := writeFile("go.mod", `module github.com/saksham/self-healing-benchmark-v3

go 1.25.6
`)
	if err != nil {
		return nil, err
	}

	// Root README.md
	_, _ = writeFile("README.md", "# Self-Healing Suite v3\nIndependent 144-fixture multi-category benchmark with zero synthetic template leakage.\n")

	// =========================================================================
	// 1. ARCHITECTURE (12 defects, 6 clean)
	// =========================================================================

	// 1.1 Layer bypass defects (6 fixtures)
	bypassConfigs := []struct {
		file, fn, pkg, driver string
	}{
		{"architecture/payment_gateway_handler.go", "ProcessPaymentCheckout", "architecture", "database/sql"},
		{"architecture/user_profile_handler.go", "UpdateProfileDetails", "architecture", "gorm.io/gorm"},
		{"architecture/subscription_handler.go", "CancelSubscriptionPlan", "architecture", "github.com/jmoiron/sqlx"},
		{"architecture/billing_invoice_handler.go", "GenerateBillingInvoice", "architecture", "database/sql"},
		{"architecture/auth_session_handler.go", "InvalidateSessionToken", "architecture", "gorm.io/gorm"},
		{"architecture/inventory_order_handler.go", "ReserveInventoryStock", "architecture", "database/sql"},
	}

	for idx, cfg := range bypassConfigs {
		i := idx + 1
		code := fmt.Sprintf(`package %s

import (
	%q
)

type %sRequest struct {
	ID    string
	Param int
}

func %s(req *%sRequest) error {
	_ = %s.ErrNoRows
	return nil
}
`, cfg.pkg, cfg.driver, cfg.fn, cfg.fn, cfg.fn, filepath.Base(cfg.driver))
		if cfg.driver == "gorm.io/gorm" {
			code = fmt.Sprintf(`package %s

import (
	"gorm.io/gorm"
)

type %sRequest struct {
	ID string
}

func %s(db *gorm.DB, req *%sRequest) error {
	return db.Error
}
`, cfg.pkg, cfg.fn, cfg.fn, cfg.fn)
		} else if cfg.driver == "github.com/jmoiron/sqlx" {
			code = fmt.Sprintf(`package %s

import (
	"github.com/jmoiron/sqlx"
)

type %sRequest struct {
	ID string
}

func %s(db *sqlx.DB, req *%sRequest) error {
	return db.Ping()
}
`, cfg.pkg, cfg.fn, cfg.fn, cfg.fn)
		}

		h, _ := writeFile(cfg.file, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID:                        fmt.Sprintf("FIX-V3-ARC-BYPASS-%02d", i),
			Category:                         repo_intel.CategoryArchitecture,
			DefectDescription:                "Presentation layer handler directly imports database driver, violating layered architecture",
			TargetFile:                       cfg.file,
			TargetSymbol:                     cfg.fn,
			GroundTruthLocation:              GroundTruthLocation{File: cfg.file, StartLine: 3, EndLine: 18},
			ExpectedDiagnosticClassification: "ARCHITECTURE",
			ExpectedSeverity:                 repo_intel.SeverityMedium,
			ExpectedConfidenceRange:          [2]float64{0.90, 1.0},
			ExpectedRemediationEligibility:   "MANUAL_REVIEW_REQUIRED",
			ExpectedExecutionStrategy:        "ARCHITECTURAL_DECOUPLING",
			FileSHA256:                       h,
		})
	}

	// 1.2 Architectural cyclic coupling defects (6 fixtures)
	cycleConfigs := []struct {
		file, typeA, typeB string
	}{
		{"architecture/order_customer_cycle.go", "OrderModelV3", "CustomerModelV3"},
		{"architecture/account_ledger_cycle.go", "AccountNodeV3", "LedgerNodeV3"},
		{"architecture/workflow_step_cycle.go", "WorkflowGraphV3", "StepGraphV3"},
		{"architecture/cluster_member_cycle.go", "ClusterCoordinatorV3", "MemberPeerV3"},
		{"architecture/session_device_cycle.go", "SessionContextV3", "DeviceRegistryV3"},
		{"architecture/ticket_agent_cycle.go", "SupportTicketV3", "AssignedAgentV3"},
	}

	for idx, cfg := range cycleConfigs {
		i := idx + 7
		code := fmt.Sprintf(`package architecture

type %s struct {
	ID       string
	Partner  *%s
	Metadata map[string]string
}

type %s struct {
	ID       string
	Parent   *%s
	Capacity int
}
`, cfg.typeA, cfg.typeB, cfg.typeB, cfg.typeA)
		h, _ := writeFile(cfg.file, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID:                        fmt.Sprintf("FIX-V3-ARC-CYCLE-%02d", i),
			Category:                         repo_intel.CategoryArchitecture,
			DefectDescription:                fmt.Sprintf("Direct bidirectional circular type coupling between %s and %s", cfg.typeA, cfg.typeB),
			TargetFile:                       cfg.file,
			TargetSymbol:                     cfg.typeA,
			GroundTruthLocation:              GroundTruthLocation{File: cfg.file, StartLine: 3, EndLine: 15},
			ExpectedDiagnosticClassification: "ARCHITECTURE",
			ExpectedSeverity:                 repo_intel.SeverityMedium,
			ExpectedConfidenceRange:          [2]float64{0.88, 1.0},
			ExpectedRemediationEligibility:   "SUPPORTED_WITH_REVIEW",
			ExpectedExecutionStrategy:        "INTERFACE_EXTRACTION",
			FileSHA256:                       h,
		})
	}

	// 1.3 Clean architectural controls (6 fixtures)
	cleanArchFiles := []struct {
		file, fn, domain string
	}{
		{"architecture/clean_layered_user_handler.go", "HandleUserQueryClean", "user"},
		{"architecture/clean_layered_product_catalog.go", "HandleCatalogQueryClean", "catalog"},
		{"architecture/clean_layered_telemetry.go", "HandleTelemetryQueryClean", "telemetry"},
		{"architecture/clean_layered_notifier.go", "HandleNotificationQueryClean", "notifier"},
		{"architecture/clean_layered_audit_trail.go", "HandleAuditQueryClean", "audit"},
		{"architecture/clean_layered_health_probe.go", "HandleHealthProbeClean", "health"},
	}

	for idx, cfg := range cleanArchFiles {
		code := fmt.Sprintf(`package architecture

type %sService interface {
	Execute(id string) (string, error)
}

type %sHandler struct {
	svc %sService
}

func (h *%sHandler) %s(id string) (string, error) {
	if id == "" { return "", nil }
	return h.svc.Execute(id)
}
`, cfg.domain, cfg.domain, cfg.domain, cfg.domain, cfg.fn)
		h, _ := writeFile(cfg.file, code)
		manifest.CleanFixtures = append(manifest.CleanFixtures, CleanFixture{
			FixtureID:    fmt.Sprintf("CLEAN-V3-ARC-%02d", idx+1),
			Category:     repo_intel.CategoryArchitecture,
			Description:  "Clean presentation layer communicating through abstracted service interface",
			TargetFile:   cfg.file,
			TargetSymbol: cfg.fn,
			FileSHA256:   h,
		})
	}

	// =========================================================================
	// 2. DEPENDENCY (12 defects, 6 clean)
	// =========================================================================

	depNames := []string{
		"auth_client", "billing_core", "shipping_calc", "analytics_agg",
		"search_indexer", "notification_hub", "etl_pipeline", "gateway_proxy",
		"reporting_worker", "ingest_agent", "cache_cluster", "sync_daemon",
	}

	for idx, name := range depNames {
		i := idx + 1
		relMod := fmt.Sprintf("dependency/%s/go.mod", name)
		content := fmt.Sprintf(`module github.com/saksham/v3/%s

go 1.25.6

require github.com/stretchr/testify v1.8.4
`, name)
		h, _ := writeFile(relMod, content)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID:                        fmt.Sprintf("FIX-V3-DEP-LOCK-%02d", i),
			Category:                         repo_intel.CategoryDependency,
			DefectDescription:                "Module go.mod has unpinned dependencies and is missing cryptographic go.sum",
			TargetFile:                       relMod,
			TargetSymbol:                     name,
			GroundTruthLocation:              GroundTruthLocation{File: relMod, StartLine: 1, EndLine: 5},
			ExpectedDiagnosticClassification: "DEPENDENCY",
			ExpectedSeverity:                 repo_intel.SeverityMedium,
			ExpectedConfidenceRange:          [2]float64{0.90, 1.0},
			ExpectedRemediationEligibility:   "FULLY_AUTONOMOUS",
			ExpectedExecutionStrategy:        "MANIFEST_TIDY",
			FileSHA256:                       h,
		})
	}

	cleanDepNames := []string{
		"order_client", "account_client", "email_client",
		"storage_client", "queue_client", "metrics_client",
	}

	for idx, name := range cleanDepNames {
		relMod := fmt.Sprintf("dependency/clean_%s/go.mod", name)
		relSum := fmt.Sprintf("dependency/clean_%s/go.sum", name)
		modContent := fmt.Sprintf(`module github.com/saksham/v3/clean_%s

go 1.25.6

require github.com/stretchr/testify v1.8.4
`, name)
		sumContent := fmt.Sprintf(`github.com/stretchr/testify v1.8.4 h1:m2GDUYFhQ7qU3s%04dxYz1=
`, idx)
		h, _ := writeFile(relMod, modContent)
		_, _ = writeFile(relSum, sumContent)
		manifest.CleanFixtures = append(manifest.CleanFixtures, CleanFixture{
			FixtureID:    fmt.Sprintf("CLEAN-V3-DEP-%02d", idx+1),
			Category:     repo_intel.CategoryDependency,
			Description:  "Clean module with cryptographically pinned dependencies and populated go.sum",
			TargetFile:   relMod,
			TargetSymbol: name,
			FileSHA256:   h,
		})
	}

	// =========================================================================
	// 3. DOCUMENTATION (12 defects, 6 clean)
	// =========================================================================

	docDefectTypes := []struct {
		file, typeName, fnName string
	}{
		{"documentation/router_dispatcher.go", "RouterDispatcherV3", "DispatchEventV3"},
		{"documentation/webhook_subscriber.go", "WebhookSubscriberV3", "DeliverEventV3"},
		{"documentation/rate_limiter_registry.go", "RateLimiterRegistryV3", "CheckQuotaV3"},
		{"documentation/token_revoker.go", "TokenRevokerV3", "RevokeTokenV3"},
		{"documentation/batch_stream_flusher.go", "BatchStreamFlusherV3", "FlushBufferV3"},
		{"documentation/cache_partition_manager.go", "CachePartitionManagerV3", "EvictStaleV3"},
		{"documentation/circuit_breaker_panel.go", "CircuitBreakerPanelV3", "RecordTripV3"},
		{"documentation/storage_blob_client.go", "StorageBlobClientV3", "UploadStreamV3"},
		{"documentation/index_bloom_filter.go", "IndexBloomFilterV3", "TestBitmaskV3"},
		{"documentation/query_ast_optimizer.go", "QueryAstOptimizerV3", "RewritePlanV3"},
		{"documentation/payload_proto_encoder.go", "PayloadProtoEncoderV3", "EncodeFrameV3"},
		{"documentation/tenant_policy_resolver.go", "TenantPolicyResolverV3", "ResolvePolicyV3"},
	}

	for idx, cfg := range docDefectTypes {
		i := idx + 1
		code := fmt.Sprintf(`package documentation

type %s struct {
	Active bool
	Zone   string
}

func (c *%s) %s(val int) int {
	if !c.Active { return 0 }
	return val * 3
}
`, cfg.typeName, cfg.typeName, cfg.fnName)
		h, _ := writeFile(cfg.file, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID:                        fmt.Sprintf("FIX-V3-DOC-UNDOC-%02d", i),
			Category:                         repo_intel.CategoryDocumentation,
			DefectDescription:                "Exported public struct and method missing required godoc comments",
			TargetFile:                       cfg.file,
			TargetSymbol:                     cfg.typeName,
			GroundTruthLocation:              GroundTruthLocation{File: cfg.file, StartLine: 3, EndLine: 12},
			ExpectedDiagnosticClassification: "DOCUMENTATION",
			ExpectedSeverity:                 repo_intel.SeverityLow,
			ExpectedConfidenceRange:          [2]float64{0.88, 1.0},
			ExpectedRemediationEligibility:   "FULLY_AUTONOMOUS",
			ExpectedExecutionStrategy:        "GODOC_SYNTHESIS",
			FileSHA256:                       h,
		})
	}

	cleanDocFiles := []struct {
		file, typeName, fnName string
	}{
		{"documentation/clean_documented_audit_logger.go", "CleanAuditLoggerV3", "RecordEntry"},
		{"documentation/clean_documented_crypto_signer.go", "CleanCryptoSignerV3", "SignMessage"},
		{"documentation/clean_documented_config_loader.go", "CleanConfigLoaderV3", "ParseConfig"},
		{"documentation/clean_documented_session_store.go", "CleanSessionStoreV3", "FetchSession"},
		{"documentation/clean_documented_metrics_exporter.go", "CleanMetricsExporterV3", "PublishMetrics"},
		{"documentation/clean_documented_task_scheduler.go", "CleanTaskSchedulerV3", "EnqueueJob"},
	}

	for idx, cfg := range cleanDocFiles {
		code := fmt.Sprintf(`package documentation

// %s coordinates auditable operations across subsystems.
type %s struct {
	Enabled bool
}

// %s processes the input parameter according to validated contracts.
func (c *%s) %s(input string) bool {
	return len(input) > 0 && c.Enabled
}
`, cfg.typeName, cfg.typeName, cfg.fnName, cfg.typeName, cfg.fnName)
		h, _ := writeFile(cfg.file, code)
		manifest.CleanFixtures = append(manifest.CleanFixtures, CleanFixture{
			FixtureID:    fmt.Sprintf("CLEAN-V3-DOC-%02d", idx+1),
			Category:     repo_intel.CategoryDocumentation,
			Description:  "Clean exported struct and method with comprehensive GoDoc documentation comments",
			TargetFile:   cfg.file,
			TargetSymbol: cfg.typeName,
			FileSHA256:   h,
		})
	}

	// =========================================================================
	// 4. MAINTAINABILITY (12 defects, 6 clean)
	// =========================================================================

	mntClones := []struct {
		name, fnA, fnB string
	}{
		{"token_normalizer", "SanitizeBearerTokenA", "SanitizeBearerTokenB"},
		{"ip_subnet_matcher", "MatchCIDRRangeA", "MatchCIDRRangeB"},
		{"hex_digest_parser", "ParseHexFingerprintA", "ParseHexFingerprintB"},
		{"url_slug_generator", "GenerateSafeSlugA", "GenerateSafeSlugB"},
		{"cron_interval_parser", "CalculateNextIntervalA", "CalculateNextIntervalB"},
		{"semantic_version_cmp", "CompareSemVerComponentsA", "CompareSemVerComponentsB"},
	}

	for idx, cfg := range mntClones {
		i1 := idx*2 + 1
		i2 := idx*2 + 2
		file1 := fmt.Sprintf("maintainability/%s_1.go", cfg.name)
		file2 := fmt.Sprintf("maintainability/%s_2.go", cfg.name)

		code1 := fmt.Sprintf(`package maintainability

import "strings"

func %s(input string) string {
	raw := strings.TrimSpace(input)
	clean := strings.ToLower(raw)
	segments := strings.Split(clean, "/")
	if len(segments) == 0 {
		return ""
	}
	filtered := strings.Join(segments, "-")
	return filtered + "_v3"
}
`, cfg.fnA)

		code2 := fmt.Sprintf(`package maintainability

import "strings"

func %s(input string) string {
	raw := strings.TrimSpace(input)
	clean := strings.ToLower(raw)
	segments := strings.Split(clean, "/")
	if len(segments) == 0 {
		return ""
	}
	filtered := strings.Join(segments, "-")
	return filtered + "_v3"
}
`, cfg.fnB)

		h1, _ := writeFile(file1, code1)
		h2, _ := writeFile(file2, code2)

		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID:                        fmt.Sprintf("FIX-V3-MNT-DUPLICATE-%02d", i1),
			Category:                         repo_intel.CategoryMaintainability,
			DefectDescription:                "Duplicated 7-statement AST implementation clone across files",
			TargetFile:                       file1,
			TargetSymbol:                     cfg.fnA,
			GroundTruthLocation:              GroundTruthLocation{File: file1, StartLine: 5, EndLine: 16},
			ExpectedDiagnosticClassification: "MAINTAINABILITY",
			ExpectedSeverity:                 repo_intel.SeverityLow,
			ExpectedConfidenceRange:          [2]float64{0.90, 1.0},
			ExpectedRemediationEligibility:   "FULLY_AUTONOMOUS",
			ExpectedExecutionStrategy:        "EXTRACT_HELPER_FUNCTION",
			FileSHA256:                       h1,
		})

		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID:                        fmt.Sprintf("FIX-V3-MNT-DUPLICATE-%02d", i2),
			Category:                         repo_intel.CategoryMaintainability,
			DefectDescription:                "Duplicated 7-statement AST implementation clone across files",
			TargetFile:                       file2,
			TargetSymbol:                     cfg.fnB,
			GroundTruthLocation:              GroundTruthLocation{File: file2, StartLine: 5, EndLine: 16},
			ExpectedDiagnosticClassification: "MAINTAINABILITY",
			ExpectedSeverity:                 repo_intel.SeverityLow,
			ExpectedConfidenceRange:          [2]float64{0.90, 1.0},
			ExpectedRemediationEligibility:   "FULLY_AUTONOMOUS",
			ExpectedExecutionStrategy:        "EXTRACT_HELPER_FUNCTION",
			FileSHA256:                       h2,
		})
	}

	cleanMntFiles := []struct {
		file, fn string
	}{
		{"maintainability/clean_slice_lookup.go", "SafeSliceLookupClean"},
		{"maintainability/clean_prefix_trimmer.go", "TrimPrefixClean"},
		{"maintainability/clean_port_parser.go", "ParsePortClean"},
		{"maintainability/clean_int_clamper.go", "ClampIntegerClean"},
		{"maintainability/clean_time_formatter.go", "FormatTimestampClean"},
		{"maintainability/clean_email_masker.go", "MaskEmailClean"},
	}

	for idx, cfg := range cleanMntFiles {
		code := fmt.Sprintf(`package maintainability

func %s(val int) int {
	if val < 0 { return 0 }
	return val + 10
}
`, cfg.fn)
		h, _ := writeFile(cfg.file, code)
		manifest.CleanFixtures = append(manifest.CleanFixtures, CleanFixture{
			FixtureID:    fmt.Sprintf("CLEAN-V3-MNT-%02d", idx+1),
			Category:     repo_intel.CategoryMaintainability,
			Description:  "Clean concise helper with no duplicate clones across the repository",
			TargetFile:   cfg.file,
			TargetSymbol: cfg.fn,
			FileSHA256:   h,
		})
	}

	// =========================================================================
	// 5. RELIABILITY (12 defects, 6 clean)
	// =========================================================================

	relLeakFiles := []struct {
		file, fn, op string
	}{
		{"reliability/stream_file_leak_1.go", "ProcessStreamLeak1", "os.Open(p)"},
		{"reliability/stream_file_leak_2.go", "ProcessStreamLeak2", "os.Open(p)"},
		{"reliability/db_conn_leak_1.go", "ProcessDBLeak1", "sql.Open(\"postgres\", p)"},
		{"reliability/db_conn_leak_2.go", "ProcessDBLeak2", "sql.Open(\"postgres\", p)"},
		{"reliability/socket_drain_leak_1.go", "ProcessSocketLeak1", "net.Dial(\"tcp\", p)"},
		{"reliability/socket_drain_leak_2.go", "ProcessSocketLeak2", "net.Dial(\"tcp\", p)"},
	}

	for idx, cfg := range relLeakFiles {
		i := idx + 1
		code := fmt.Sprintf(`package reliability

import (
	"database/sql"
	"net"
	"os"
)

func %s(paths []string) error {
	for _, p := range paths {
		res, err := %s
		if err != nil {
			return err
		}
		_ = res
	}
	return nil
}
`, cfg.fn, cfg.op)
		h, _ := writeFile(cfg.file, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID:                        fmt.Sprintf("FIX-V3-REL-LEAK-%02d", i),
			Category:                         repo_intel.CategoryReliability,
			DefectDescription:                "Resource handle opened in loop without Close() call, leaking descriptors",
			TargetFile:                       cfg.file,
			TargetSymbol:                     cfg.fn,
			GroundTruthLocation:              GroundTruthLocation{File: cfg.file, StartLine: 9, EndLine: 19},
			ExpectedDiagnosticClassification: "RELIABILITY",
			ExpectedSeverity:                 repo_intel.SeverityHigh,
			ExpectedConfidenceRange:          [2]float64{0.92, 1.0},
			ExpectedRemediationEligibility:   "FULLY_AUTONOMOUS",
			ExpectedExecutionStrategy:        "INSERT_CLOSE_CALL",
			FileSHA256:                       h,
		})
	}

	relDeferFiles := []struct {
		file, fn, op string
	}{
		{"reliability/defer_loop_file_1.go", "ProcessDeferLoop1", "os.Open(p)"},
		{"reliability/defer_loop_file_2.go", "ProcessDeferLoop2", "os.Open(p)"},
		{"reliability/defer_loop_db_1.go", "ProcessDeferLoopDB1", "sql.Open(\"sqlite3\", p)"},
		{"reliability/defer_loop_db_2.go", "ProcessDeferLoopDB2", "sql.Open(\"sqlite3\", p)"},
		{"reliability/defer_loop_net_1.go", "ProcessDeferLoopNet1", "net.Dial(\"udp\", p)"},
		{"reliability/defer_loop_net_2.go", "ProcessDeferLoopNet2", "net.Dial(\"udp\", p)"},
	}

	for idx, cfg := range relDeferFiles {
		i := idx + 7
		code := fmt.Sprintf(`package reliability

import (
	"database/sql"
	"net"
	"os"
)

func %s(items []string) error {
	for _, p := range items {
		res, err := %s
		if err != nil {
			return err
		}
		defer res.Close()
	}
	return nil
}
`, cfg.fn, cfg.op)
		h, _ := writeFile(cfg.file, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID:                        fmt.Sprintf("FIX-V3-REL-DEFER-%02d", i),
			Category:                         repo_intel.CategoryReliability,
			DefectDescription:                "Resource defer placed directly inside loop body, deferring cleanup until outer function returns",
			TargetFile:                       cfg.file,
			TargetSymbol:                     cfg.fn,
			GroundTruthLocation:              GroundTruthLocation{File: cfg.file, StartLine: 9, EndLine: 20},
			ExpectedDiagnosticClassification: "RELIABILITY",
			ExpectedSeverity:                 repo_intel.SeverityMedium,
			ExpectedConfidenceRange:          [2]float64{0.90, 1.0},
			ExpectedRemediationEligibility:   "FULLY_AUTONOMOUS",
			ExpectedExecutionStrategy:        "EXTRACT_LOOP_CLOSURE",
			FileSHA256:                       h,
		})
	}

	cleanRelFiles := []struct {
		file, fn string
	}{
		{"reliability/clean_closure_stream.go", "StreamFileWithClosureClean"},
		{"reliability/clean_batch_socket.go", "ProcessBatchedSocketsClean"},
		{"reliability/clean_payload_downloader.go", "DownloadPayloadSafelyClean"},
		{"reliability/clean_config_reader.go", "ReadConfigFileStrictClean"},
		{"reliability/clean_temp_lock.go", "AcquireTempLockClean"},
		{"reliability/clean_db_transaction.go", "TransactDatabaseSafeClean"},
	}

	for idx, cfg := range cleanRelFiles {
		code := fmt.Sprintf(`package reliability

import "os"

func %s(p string) ([]byte, error) {
	return os.ReadFile(p)
}
`, cfg.fn)
		h, _ := writeFile(cfg.file, code)
		manifest.CleanFixtures = append(manifest.CleanFixtures, CleanFixture{
			FixtureID:    fmt.Sprintf("CLEAN-V3-REL-%02d", idx+1),
			Category:     repo_intel.CategoryReliability,
			Description:  "Clean resource management using atomic helper function with automatic descriptor cleanup",
			TargetFile:   cfg.file,
			TargetSymbol: cfg.fn,
			FileSHA256:   h,
		})
	}

	// =========================================================================
	// 6. SECURITY (12 defects, 6 clean)
	// =========================================================================

	sqlDefects := []struct {
		file, fn, table string
	}{
		{"security/sql_user_lookup.go", "QueryUserByIDUnsafe", "users"},
		{"security/sql_order_fetch.go", "FetchOrderByCodeUnsafe", "orders"},
		{"security/sql_billing_search.go", "SearchBillingPlanUnsafe", "billing_plans"},
		{"security/sql_audit_filter.go", "FilterAuditRecordsUnsafe", "audit_logs"},
	}

	for idx, cfg := range sqlDefects {
		i := idx + 1
		code := fmt.Sprintf(`package security

import "database/sql"

func %s(db *sql.DB, inputID string) (*sql.Rows, error) {
	query := "SELECT * FROM %s WHERE id = '" + inputID + "'"
	return db.Query(query)
}
`, cfg.fn, cfg.table)
		h, _ := writeFile(cfg.file, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID:                        fmt.Sprintf("FIX-V3-SEC-SQL-%02d", i),
			Category:                         repo_intel.CategorySecurity,
			DefectDescription:                "SQL query dynamically constructed via string binary concatenation without parameterized binding",
			TargetFile:                       cfg.file,
			TargetSymbol:                     cfg.fn,
			GroundTruthLocation:              GroundTruthLocation{File: cfg.file, StartLine: 5, EndLine: 9},
			ExpectedDiagnosticClassification: "SECURITY",
			ExpectedSeverity:                 repo_intel.SeverityHigh,
			ExpectedConfidenceRange:          [2]float64{0.95, 1.0},
			ExpectedRemediationEligibility:   "FULLY_AUTONOMOUS",
			ExpectedExecutionStrategy:        "SQL_PARAMETERIZE",
			FileSHA256:                       h,
		})
	}

	cmdDefects := []struct {
		file, fn, cmd string
	}{
		{"security/cmd_ping_target.go", "ExecutePingUnsafe", "ping -c 1"},
		{"security/cmd_dns_lookup.go", "ExecuteDNSQueryUnsafe", "nslookup"},
		{"security/cmd_system_backup.go", "ExecuteBackupUnsafe", "tar -czf"},
		{"security/cmd_log_rotate.go", "ExecuteLogRotateUnsafe", "logrotate"},
	}

	for idx, cfg := range cmdDefects {
		i := idx + 5
		code := fmt.Sprintf(`package security

import (
	"fmt"
	"os/exec"
)

func %s(host string) error {
	cmdStr := fmt.Sprintf("%s %%s", host)
	cmd := exec.Command("sh", "-c", cmdStr)
	return cmd.Run()
}
`, cfg.fn, cfg.cmd)
		h, _ := writeFile(cfg.file, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID:                        fmt.Sprintf("FIX-V3-SEC-CMD-%02d", i),
			Category:                         repo_intel.CategorySecurity,
			DefectDescription:                "OS command executed via shell with dynamic format interpolation, risking command injection",
			TargetFile:                       cfg.file,
			TargetSymbol:                     cfg.fn,
			GroundTruthLocation:              GroundTruthLocation{File: cfg.file, StartLine: 8, EndLine: 14},
			ExpectedDiagnosticClassification: "SECURITY",
			ExpectedSeverity:                 repo_intel.SeverityHigh,
			ExpectedConfidenceRange:          [2]float64{0.94, 1.0},
			ExpectedRemediationEligibility:   "FULLY_AUTONOMOUS",
			ExpectedExecutionStrategy:        "SANITIZE_COMMAND_ARGS",
			FileSHA256:                       h,
		})
	}

	secDefects := []struct {
		file, varName, token string
	}{
		{"security/hardcoded_aws_token_1.go", "MasterAWSKeyV3", "AKIA9988776655443322"},
		{"security/hardcoded_aws_token_2.go", "ProductionRootAKIA", "AKIA5566778899001122"},
		{"security/hardcoded_aws_token_3.go", "ClientAWSAccessKey", "AKIA3344556677889900"},
		{"security/hardcoded_aws_token_4.go", "S3BackupTokenAKIA", "AKIA1234432156788765"},
	}

	for idx, cfg := range secDefects {
		i := idx + 9
		code := fmt.Sprintf(`package security

// Hardcoded credential defect
const %s = %q
`, cfg.varName, cfg.token)
		h, _ := writeFile(cfg.file, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID:                        fmt.Sprintf("FIX-V3-SEC-SECRET-%02d", i),
			Category:                         repo_intel.CategorySecurity,
			DefectDescription:                "Hardcoded plaintext AWS Access Key ID embedded directly in source code",
			TargetFile:                       cfg.file,
			TargetSymbol:                     cfg.varName,
			GroundTruthLocation:              GroundTruthLocation{File: cfg.file, StartLine: 3, EndLine: 5},
			ExpectedDiagnosticClassification: "SECURITY",
			ExpectedSeverity:                 repo_intel.SeverityHigh,
			ExpectedConfidenceRange:          [2]float64{0.95, 1.0},
			ExpectedRemediationEligibility:   "FULLY_AUTONOMOUS",
			ExpectedExecutionStrategy:        "SECRET_ENV_SUBSTITUTION",
			FileSHA256:                       h,
		})
	}

	cleanSecFiles := []struct {
		file, fn string
	}{
		{"security/clean_sql_parameterized_1.go", "QueryUserSafeClean"},
		{"security/clean_sql_parameterized_2.go", "FetchOrderSafeClean"},
		{"security/clean_cmd_static_args_1.go", "ExecuteGitStatusSafeClean"},
		{"security/clean_cmd_static_args_2.go", "ExecuteVersionSafeClean"},
		{"security/clean_secret_env_1.go", "FetchTokenFromEnvSafeClean"},
		{"security/clean_secret_env_2.go", "FetchKeyFromEnvSafeClean"},
	}

	for idx, cfg := range cleanSecFiles {
		var code string
		if idx < 2 {
			code = fmt.Sprintf(`package security

import "database/sql"

func %s(db *sql.DB, id string) (*sql.Rows, error) {
	return db.Query("SELECT * FROM items WHERE id = ?", id)
}
`, cfg.fn)
		} else if idx < 4 {
			code = fmt.Sprintf(`package security

import "os/exec"

func %s() error {
	cmd := exec.Command("git", "status")
	return cmd.Run()
}
`, cfg.fn)
		} else {
			code = fmt.Sprintf(`package security

import "os"

func %s() string {
	return os.Getenv("AWS_SECRET_ACCESS_KEY")
}
`, cfg.fn)
		}

		h, _ := writeFile(cfg.file, code)
		manifest.CleanFixtures = append(manifest.CleanFixtures, CleanFixture{
			FixtureID:    fmt.Sprintf("CLEAN-V3-SEC-%02d", idx+1),
			Category:     repo_intel.CategorySecurity,
			Description:  "Clean security implementation utilizing parameterized queries, static args, or environment variables",
			TargetFile:   cfg.file,
			TargetSymbol: cfg.fn,
			FileSHA256:   h,
		})
	}

	// =========================================================================
	// 7. STRUCTURAL (12 defects, 6 clean)
	// =========================================================================

	for i := 1; i <= 6; i++ {
		rel := fmt.Sprintf("structural/deep_nest_tree_%d.go", i)
		fnName := fmt.Sprintf("EvaluateDeepNestingTreeV3_%d", i)
		code := fmt.Sprintf(`package structural

func %s(data map[string]interface{}) int {
	res := 0
	if data != nil {
		if val, ok := data["payload"]; ok {
			if list, isList := val.([]int); isList {
				for _, num := range list {
					if num > 5 {
						if num%%2 == 0 {
							if num < 100 {
								res += num * %d
							}
						}
					}
				}
			}
		}
	}
	return res
}
`, fnName, i)
		h, _ := writeFile(rel, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID:                        fmt.Sprintf("FIX-V3-STR-NEST-%02d", i),
			Category:                         repo_intel.CategoryStructural,
			DefectDescription:                "Excessive control flow nesting depth exceeding threshold (>= 5 levels)",
			TargetFile:                       rel,
			TargetSymbol:                     fnName,
			GroundTruthLocation:              GroundTruthLocation{File: rel, StartLine: 3, EndLine: 22},
			ExpectedDiagnosticClassification: "STRUCTURAL",
			ExpectedSeverity:                 repo_intel.SeverityMedium,
			ExpectedConfidenceRange:          [2]float64{0.85, 1.0},
			ExpectedRemediationEligibility:   "SUPPORTED_WITH_REVIEW",
			ExpectedExecutionStrategy:        "AST_TRANSFORMATION",
			FileSHA256:                       h,
		})
	}

	for i := 1; i <= 6; i++ {
		idx := i + 6
		rel := fmt.Sprintf("structural/high_complexity_calc_%d.go", i)
		fnName := fmt.Sprintf("ComputeMultiDecisionV3_%d", i)
		code := fmt.Sprintf(`package structural

func %s(a, b, c, d int, mode string) int {
	res := 0
	if a > 0 { res += 1 } else { res -= 1 }
	if b > 10 { res += 2 } else { res -= 2 }
	if c > 20 { res += 3 } else { res -= 3 }
	if d > 30 { res += 4 } else { res -= 4 }
	switch mode {
	case "alpha": res *= 2
	case "beta": res *= 3
	case "gamma": res *= 4
	case "delta": res *= 5
	default: res += 0
	}
	if a == b || c == d { res += 10 }
	if a > c && b > d { res += 20 }
	return res
}
`, fnName)
		h, _ := writeFile(rel, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID:                        fmt.Sprintf("FIX-V3-STR-COMPLEX-%02d", idx),
			Category:                         repo_intel.CategoryStructural,
			DefectDescription:                "Excessive cyclomatic complexity exceeding cognitive load threshold (>= 10)",
			TargetFile:                       rel,
			TargetSymbol:                     fnName,
			GroundTruthLocation:              GroundTruthLocation{File: rel, StartLine: 3, EndLine: 20},
			ExpectedDiagnosticClassification: "STRUCTURAL",
			ExpectedSeverity:                 repo_intel.SeverityMedium,
			ExpectedConfidenceRange:          [2]float64{0.85, 1.0},
			ExpectedRemediationEligibility:   "SUPPORTED_WITH_REVIEW",
			ExpectedExecutionStrategy:        "REFACTOR_GUARD_CLAUSES",
			FileSHA256:                       h,
		})
	}

	cleanStrFiles := []struct {
		file, fn string
	}{
		{"structural/clean_flat_guard_1.go", "EvaluatePrivilegeClean1"},
		{"structural/clean_flat_guard_2.go", "EvaluatePrivilegeClean2"},
		{"structural/clean_simple_calc_1.go", "SanitizeHeaderClean1"},
		{"structural/clean_simple_calc_2.go", "SanitizeHeaderClean2"},
		{"structural/clean_filter_records_1.go", "FilterValidRecordsClean1"},
		{"structural/clean_filter_records_2.go", "FilterValidRecordsClean2"},
	}

	for idx, cfg := range cleanStrFiles {
		code := fmt.Sprintf(`package structural

func %s(data map[string]int) int {
	if data == nil { return 0 }
	val, ok := data["tier"]
	if !ok || val <= 0 { return 0 }
	return val * 10
}
`, cfg.fn)
		h, _ := writeFile(cfg.file, code)
		manifest.CleanFixtures = append(manifest.CleanFixtures, CleanFixture{
			FixtureID:    fmt.Sprintf("CLEAN-V3-STR-%02d", idx+1),
			Category:     repo_intel.CategoryStructural,
			Description:  "Clean linear logic with early guard-clause returns and low cyclomatic complexity",
			TargetFile:   cfg.file,
			TargetSymbol: cfg.fn,
			FileSHA256:   h,
		})
	}

	// =========================================================================
	// 8. TEST_QUALITY (12 defects, 6 clean)
	// =========================================================================

	for i := 1; i <= 6; i++ {
		rel := fmt.Sprintf("testing/untested_billing_calc_%d.go", i)
		fnName := fmt.Sprintf("CalculateTieredDiscountV3_%d", i)
		code := fmt.Sprintf(`package testing

func %s(baseAmount float64, tierLevel int) float64 {
	if tierLevel == 1 {
		return baseAmount * 0.90
	} else if tierLevel == 2 {
		return baseAmount * 0.80
	} else if tierLevel == 3 {
		return baseAmount * 0.70
	}
	return baseAmount
}
`, fnName)
		h, _ := writeFile(rel, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID:                        fmt.Sprintf("FIX-V3-TST-UNTESTED-%02d", i),
			Category:                         repo_intel.CategoryTestQuality,
			DefectDescription:                "Exported business logic calculation function has zero automated unit test invocations in package test suite",
			TargetFile:                       rel,
			TargetSymbol:                     fnName,
			GroundTruthLocation:              GroundTruthLocation{File: rel, StartLine: 3, EndLine: 13},
			ExpectedDiagnosticClassification: "TEST_QUALITY",
			ExpectedSeverity:                 repo_intel.SeverityMedium,
			ExpectedConfidenceRange:          [2]float64{0.88, 1.0},
			ExpectedRemediationEligibility:   "FULLY_AUTONOMOUS",
			ExpectedExecutionStrategy:        "SYNTHESIZE_UNIT_TEST",
			FileSHA256:                       h,
		})
	}

	for i := 1; i <= 6; i++ {
		idx := i + 6
		rel := fmt.Sprintf("testing/untested_auth_engine_%d.go", i)
		fnName := fmt.Sprintf("EvaluateTokenScopesV3_%d", i)
		code := fmt.Sprintf(`package testing

func %s(scopes []string, targetScope string) bool {
	for _, s := range scopes {
		if s == targetScope || s == "admin" {
			return true
		}
	}
	return false
}
`, fnName)
		h, _ := writeFile(rel, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID:                        fmt.Sprintf("FIX-V3-TST-UNTESTED-%02d", idx),
			Category:                         repo_intel.CategoryTestQuality,
			DefectDescription:                "Exported authorization scope checker has zero automated unit test invocations in package test suite",
			TargetFile:                       rel,
			TargetSymbol:                     fnName,
			GroundTruthLocation:              GroundTruthLocation{File: rel, StartLine: 3, EndLine: 12},
			ExpectedDiagnosticClassification: "TEST_QUALITY",
			ExpectedSeverity:                 repo_intel.SeverityMedium,
			ExpectedConfidenceRange:          [2]float64{0.88, 1.0},
			ExpectedRemediationEligibility:   "FULLY_AUTONOMOUS",
			ExpectedExecutionStrategy:        "SYNTHESIZE_UNIT_TEST",
			FileSHA256:                       h,
		})
	}

	// Create test file in testing/ so package has tests, making untested symbols properly detected
	_, _ = writeFile("testing/suite_runner_test.go", `package testing

import "testing"

func TestBaseSuite(t *testing.T) {
	// Base placeholder test
}
`)

	cleanTestServices := []struct {
		file, testFile, fnName string
	}{
		{"testing/clean_auth_service.go", "testing/clean_auth_service_test.go", "ValidateAuthSessionClean"},
		{"testing/clean_payment_service.go", "testing/clean_payment_service_test.go", "ValidatePaymentAuthClean"},
		{"testing/clean_user_service.go", "testing/clean_user_service_test.go", "ValidateUserProfileClean"},
		{"testing/clean_order_service.go", "testing/clean_order_service_test.go", "ValidateOrderDetailsClean"},
		{"testing/clean_token_service.go", "testing/clean_token_service_test.go", "ValidateTokenIssuerClean"},
		{"testing/clean_cache_service.go", "testing/clean_cache_service_test.go", "ValidateCacheEntryClean"},
	}

	for idx, cfg := range cleanTestServices {
		srcCode := fmt.Sprintf(`package testing

func %s(key string) bool {
	return len(key) >= 4
}
`, cfg.fnName)
		testCode := fmt.Sprintf(`package testing

import "testing"

func Test%s(t *testing.T) {
	if !%s("valid_key") {
		t.Errorf("Expected true for valid_key")
	}
}
`, cfg.fnName, cfg.fnName)

		h, _ := writeFile(cfg.file, srcCode)
		_, _ = writeFile(cfg.testFile, testCode)

		manifest.CleanFixtures = append(manifest.CleanFixtures, CleanFixture{
			FixtureID:    fmt.Sprintf("CLEAN-V3-TST-%02d", idx+1),
			Category:     repo_intel.CategoryTestQuality,
			Description:  "Clean business service fully exercised and validated by automated unit test suite",
			TargetFile:   cfg.file,
			TargetSymbol: cfg.fnName,
			FileSHA256:   h,
		})
	}

	// Persist manifest
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal manifest: %w", err)
	}

	if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
		return nil, fmt.Errorf("write manifest to %s: %w", manifestPath, err)
	}

	h := sha256.Sum256(manifestData)
	manifest.ManifestSHA256 = hex.EncodeToString(h[:])

	return manifest, nil
}
