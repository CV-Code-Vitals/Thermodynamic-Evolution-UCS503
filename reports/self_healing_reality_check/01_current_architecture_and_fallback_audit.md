# Architecture and baseline audit
Starting local and remote branch HEAD: 4b992caf951aef397ec9ef043bb9f1e3755b0cae.
Branch: release/milestone-3.8-mvp-rc. Existing report modifications and Rust source changes were preserved; starting-status.txt and collaborator-starting.patch capture the baseline working tree.

Public URL -> scanner.go git clone -> Rust AST binary -> retained workspace -> remediation targets -> semantic UnifiedOrchestrator -> snapshot-bound target -> AnalyzeTarget -> ProposeRepair -> AST synthesizer -> approval -> isolated copy -> oracle suite -> audit/rollback.

Baseline ListTargets appended five demos without source and swallowed discovery errors. UI selected INJ-EXT2-RES-02. Live test used fixed SQL/command IDs and an unavailable runner. Repository-provided thermodynamic_report.json could manufacture hotspot claims. The AST fallback classified allocations without proving missing cleanup. These paths were corrected or removed from live discovery.

Live semantic passes inspect real Go ASTs, but receiver inference and control-flow analysis remain bounded heuristics. The public executor supplies empty reproducer/invariance commands and no regression commands. Therefore its complete independently verified lifecycle is not established. Thermodynamic scores describe source structure, not measured runtime improvements.

The Rust binary was rebuilt from the current working tree, including the collaborator's existing main.rs changes. This is not a pristine starting-commit binary.
