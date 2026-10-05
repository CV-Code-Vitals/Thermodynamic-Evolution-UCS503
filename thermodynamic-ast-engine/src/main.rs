// =============================================================================
// thermodynamic-ast-engine · src/main.rs  (v2.0 — Tree-sitter AST-driven)
//
// A structural thermodynamic entropy analyzer for Go and Python source files.
//
// Design pillars
// ──────────────
//   1. Zero unsafe code — all parsing through tree-sitter safe bindings.
//   2. Concrete AST visitors replace all regex heuristics.
//   3. Structural Energy formula: E = w1·M + w2·Dmax + w3·log2(S) + w4·C
//   4. File scanning runs in parallel via Rayon (feature-gated).
//   5. All public data types implement Serialize for JSON report emission.
// =============================================================================

use clap::Parser;
use colored::Colorize;
use serde::{Deserialize, Serialize};
use std::{
    fs,
    io,
    path::{Path, PathBuf},
};
use walkdir::WalkDir;

#[cfg(feature = "parallel")]
use rayon::prelude::*;

// =============================================================================
// § 1  CLI argument definition (clap derive)
// =============================================================================

/// Thermodynamic AST Engine v2 — calculates structural energy from AST analysis
#[derive(Parser, Debug)]
#[command(
    name        = "thermodynamic-ast-engine",
    version     = env!("CARGO_PKG_VERSION"),
    author      = env!("CARGO_PKG_AUTHORS"),
    about       = "Tree-sitter AST-driven thermodynamic energy analyzer for Go/Python source \
                   files. Emits a JSON report of structural entropy hotspots.",
    long_about  = None,
)]
struct Cli {
    /// Root directory to scan (recursively)
    #[arg(value_name = "DIR")]
    directory: PathBuf,

    /// Output file path for the JSON report
    #[arg(
        short,
        long,
        value_name = "FILE",
        default_value = "thermodynamic_report.json"
    )]
    output: PathBuf,

    /// Minimum energy score to include in the report
    #[arg(short, long, value_name = "SCORE", default_value_t = 0.0)]
    min_score: f64,

    /// Energy threshold for flagging hotspots
    #[arg(short = 't', long, value_name = "THRESHOLD", default_value_t = 15.0)]
    energy_threshold: f64,

    /// Show verbose per-file progress in stdout
    #[arg(short, long)]
    verbose: bool,
}

// =============================================================================
// § 2  Core data model
// =============================================================================

/// The vulnerability / entropy driver category detected by AST analysis.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub enum VulnerabilityType {
    /// Deeply nested control-flow blocks (O(n^k) risk)
    DeepNesting,
    /// Direct or mutual recursion without obvious guardrails
    RecursiveCall,
    /// Heap allocation inside a hot loop path
    HotAllocation,
    /// Blocking I/O or syscall on the critical path
    BlockingIO,
    /// High cyclomatic complexity (many decision points)
    CognitiveBranch,
    /// Unsafe synchronisation primitive (mutex inside loop, etc.)
    SyncContention,
    /// High structural energy from combined metrics
    HighEnergy,
}

impl std::fmt::Display for VulnerabilityType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::DeepNesting => write!(f, "DeepNesting"),
            Self::RecursiveCall => write!(f, "RecursiveCall"),
            Self::HotAllocation => write!(f, "HotAllocation"),
            Self::BlockingIO => write!(f, "BlockingIO"),
            Self::CognitiveBranch => write!(f, "CognitiveBranch"),
            Self::SyncContention => write!(f, "SyncContention"),
            Self::HighEnergy => write!(f, "HighEnergy"),
        }
    }
}

/// Detailed per-function metrics extracted from the AST.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FunctionMetrics {
    /// Cyclomatic complexity: decision points + 1
    pub cyclomatic_complexity: usize,
    /// Maximum AST depth of nested control-flow blocks
    pub max_nesting_depth: usize,
    /// Total named AST nodes within the function scope
    pub ast_node_count: usize,
    /// Count of concurrency / resource primitives
    pub concurrency_primitives: usize,
    /// Structural Thermodynamic Energy
    pub energy: f64,
}

/// A single entropy "hotspot" — one detected signal within a file.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Hotspot {
    /// Name of the containing function / method
    pub function_name: String,
    /// 1-indexed line number where the function starts
    pub line_number: usize,
    /// Byte range [start, end] within the source file
    pub byte_range: [usize; 2],
    /// The raw source snippet (first line of the function signature)
    pub source_snippet: String,
    /// Structural energy score for this function
    pub entropy_score: f64,
    /// Classification of the primary entropy driver
    pub vulnerability_type: VulnerabilityType,
    /// Human-readable explanation of why this is flagged
    pub description: String,
    /// Full metric breakdown
    pub metrics: FunctionMetrics,
}

/// Aggregated report for one source file.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FileReport {
    /// Path to the source file (relative to scanned directory)
    pub file_path: String,
    /// Programming language detected from the file extension
    pub language: String,
    /// Total number of non-blank, non-comment lines scanned
    pub lines_scanned: usize,
    /// Sum of all individual hotspot energy scores
    pub total_entropy: f64,
    /// Mean energy per detected hotspot (0 if no hotspots)
    pub mean_hotspot_entropy: f64,
    /// All identified hotspots, sorted highest-energy first
    pub hotspots: Vec<Hotspot>,
}

/// Root report document written to disk.
#[derive(Debug, Serialize, Deserialize)]
pub struct ThermodynamicReport {
    /// Engine version for schema compatibility
    pub engine_version: String,
    /// ISO-8601 timestamp when the scan completed
    pub generated_at: String,
    /// Directory that was scanned
    pub scanned_directory: String,
    /// Total source files analyzed
    pub files_analyzed: usize,
    /// Total hotspots found across all files
    pub total_hotspots: usize,
    /// Aggregate entropy across the entire codebase
    pub global_entropy: f64,
    /// Per-file reports, sorted by `total_entropy` descending
    pub file_reports: Vec<FileReport>,
}

// =============================================================================
// § 3  Thermodynamic energy formula
// =============================================================================

/// Default weights for the structural energy formula.
/// E = w1·M + w2·Dmax + w3·log2(S) + w4·C
const W1_CYCLOMATIC: f64 = 0.45;
const W2_NESTING: f64 = 0.25;
const W3_VOLUME: f64 = 0.15;
const W4_CONCURRENCY: f64 = 0.15;

/// Compute structural thermodynamic energy for a function.
pub fn compute_energy(metrics: &FunctionMetrics) -> f64 {
    let m = metrics.cyclomatic_complexity as f64;
    let d = metrics.max_nesting_depth as f64;
    let s = if metrics.ast_node_count > 0 {
        (metrics.ast_node_count as f64).log2()
    } else {
        0.0
    };
    let c = metrics.concurrency_primitives as f64;

    let e = W1_CYCLOMATIC * m + W2_NESTING * d + W3_VOLUME * s + W4_CONCURRENCY * c;
    (e * 100.0).round() / 100.0
}

// =============================================================================
// § 4  Tree-sitter AST visitor engine
// =============================================================================

/// Supported language enum for the visitor dispatch.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Language {
    Go,
    Python,
}

/// Create a tree-sitter parser for the given language.
fn create_parser(lang: Language) -> Option<tree_sitter::Parser> {
    let mut parser = tree_sitter::Parser::new();
    let ts_lang = match lang {
        Language::Go => tree_sitter_go::LANGUAGE.into(),
        Language::Python => tree_sitter_python::LANGUAGE.into(),
    };
    parser.set_language(&ts_lang).ok()?;
    Some(parser)
}

/// Decision-point node kinds for cyclomatic complexity (per language).
fn is_decision_point(kind: &str, lang: Language) -> bool {
    match lang {
        Language::Go => matches!(
            kind,
            "if_statement"
                | "for_statement"
                | "switch_statement"
                | "select_statement"
                | "expression_case"
                | "default_case"
                | "type_case"
                | "communication_case"
                | "go_statement"
        ),
        Language::Python => matches!(
            kind,
            "if_statement"
                | "elif_clause"
                | "for_statement"
                | "while_statement"
                | "except_clause"
                | "with_statement"
                | "assert_statement"
                | "conditional_expression"
        ),
    }
}

/// Binary boolean operators that add to cyclomatic complexity.
fn is_boolean_operator(kind: &str, text: &str, lang: Language) -> bool {
    match lang {
        Language::Go => {
            kind == "binary_expression"
                && (text.contains("&&") || text.contains("||"))
        }
        Language::Python => {
            kind == "boolean_operator"
        }
    }
}

/// Check if a node represents a control-flow nesting block.
fn is_nesting_block(kind: &str, lang: Language) -> bool {
    match lang {
        Language::Go => matches!(
            kind,
            "if_statement"
                | "for_statement"
                | "switch_statement"
                | "select_statement"
                | "func_literal"
        ),
        Language::Python => matches!(
            kind,
            "if_statement"
                | "elif_clause"
                | "else_clause"
                | "for_statement"
                | "while_statement"
                | "with_statement"
                | "try_statement"
        ),
    }
}

/// Check if a node is a function/method declaration.
fn is_function_node(kind: &str, lang: Language) -> bool {
    match lang {
        Language::Go => matches!(kind, "function_declaration" | "method_declaration"),
        Language::Python => matches!(kind, "function_definition"),
    }
}

/// Extract function name from a function AST node.
fn extract_function_name(node: &tree_sitter::Node, source: &[u8], lang: Language) -> String {
    match lang {
        Language::Go => {
            // For method_declaration, the name is inside the receiver + name
            // For function_declaration, it's the `name` child
            node.child_by_field_name("name")
                .and_then(|n| n.utf8_text(source).ok())
                .unwrap_or("<anonymous>")
                .to_string()
        }
        Language::Python => {
            node.child_by_field_name("name")
                .and_then(|n| n.utf8_text(source).ok())
                .unwrap_or("<anonymous>")
                .to_string()
        }
    }
}

/// Check if a node represents a concurrency / resource primitive.
fn is_concurrency_primitive(kind: &str, text: &str, lang: Language) -> bool {
    match lang {
        Language::Go => {
            // goroutine spawn
            if kind == "go_statement" {
                return true;
            }
            // channel operations
            if kind == "send_statement" || kind == "receive_expression" {
                return true;
            }
            // defer statement (resource management)
            if kind == "defer_statement" {
                return true;
            }
            // sync.Mutex, sync.WaitGroup, atomic operations
            if kind == "selector_expression" || kind == "call_expression" {
                let t = text.to_lowercase();
                if t.contains("sync.mutex")
                    || t.contains("sync.rwmutex")
                    || t.contains("sync.waitgroup")
                    || t.contains("atomic.")
                    || t.contains(".lock()")
                    || t.contains(".unlock()")
                    || t.contains(".rlock()")
                    || t.contains(".runlock()")
                {
                    return true;
                }
            }
            false
        }
        Language::Python => {
            if kind == "call" || kind == "attribute" {
                let t = text.to_lowercase();
                if t.contains("threading.lock")
                    || t.contains("threading.rlock")
                    || t.contains("threading.semaphore")
                    || t.contains("asyncio.lock")
                    || t.contains("multiprocessing.lock")
                    || t.contains(".acquire()")
                    || t.contains(".release()")
                {
                    return true;
                }
            }
            if kind == "with_statement" {
                let t = text.to_lowercase();
                if t.contains("lock") || t.contains("semaphore") {
                    return true;
                }
            }
            false
        }
    }
}

/// Check if a call node represents a blocking I/O operation.
fn is_blocking_io(text: &str, lang: Language) -> bool {
    let t = text.to_lowercase();
    match lang {
        Language::Go => {
            t.contains("os.open")
                || t.contains("os.create")
                || t.contains("ioutil.readfile")
                || t.contains("ioutil.writefile")
                || t.contains("http.get")
                || t.contains("http.post")
                || t.contains("net.dial")
                || t.contains("time.sleep")
                || t.contains("bufio.newreader")
                || t.contains("sql.open")
                || t.contains(".readstring(")
                || t.contains(".readline(")
        }
        Language::Python => {
            t.contains("requests.get")
                || t.contains("requests.post")
                || t.contains("requests.put")
                || t.contains("requests.delete")
                || t.contains("time.sleep")
                || t.contains("open(")
                || t.contains("subprocess.call")
                || t.contains("subprocess.run")
                || t.contains("socket.recv")
                || t.contains("urllib")
        }
    }
}

/// Check if a call node represents a heap allocation in a hot path.
fn is_allocation(text: &str, lang: Language) -> bool {
    let t = text.to_lowercase();
    match lang {
        Language::Go => {
            t.starts_with("make(") || t.starts_with("new(") || t.contains("append(")
        }
        Language::Python => {
            t.contains("list(")
                || t.contains("dict(")
                || t.contains("set(")
                || t.contains("np.zeros")
                || t.contains("np.ones")
                || t.contains("numpy.zeros")
                || t.contains("numpy.ones")
                || t.contains("torch.zeros")
                || t.contains("torch.ones")
                || t.contains("bytearray(")
        }
    }
}

/// Recursively collect metrics from AST nodes within a function scope.
struct FunctionVisitor<'a> {
    source: &'a [u8],
    lang: Language,
    func_name: String,
    cyclomatic: usize,
    max_depth: usize,
    node_count: usize,
    concurrency: usize,
    has_recursive_call: bool,
    blocking_io_count: usize,
    allocation_in_loop: bool,
    loop_depth: usize,
    current_depth: usize,
}

impl<'a> FunctionVisitor<'a> {
    fn new(source: &'a [u8], lang: Language, func_name: String) -> Self {
        Self {
            source,
            lang,
            func_name,
            cyclomatic: 1, // base complexity
            max_depth: 0,
            node_count: 0,
            concurrency: 0,
            has_recursive_call: false,
            blocking_io_count: 0,
            allocation_in_loop: false,
            loop_depth: 0,
            current_depth: 0,
        }
    }

    fn visit(&mut self, node: tree_sitter::Node) {
        let kind = node.kind();

        // Count named AST nodes (volume metric)
        if node.is_named() {
            self.node_count += 1;
        }

        // Decision points → cyclomatic complexity
        if is_decision_point(kind, self.lang) {
            self.cyclomatic += 1;
        }

        // Boolean operators → additional cyclomatic branches
        if let Ok(text) = node.utf8_text(self.source) {
            if is_boolean_operator(kind, text, self.lang) {
                // Count each && or || as one additional path
                let count = text.matches("&&").count()
                    + text.matches("||").count()
                    + if kind == "boolean_operator" { 1 } else { 0 };
                // Avoid double counting: only add extras beyond the first
                if count > 0 {
                    self.cyclomatic += count.saturating_sub(1).max(1);
                }
            }
        }

        // Track nesting depth
        let is_nesting = is_nesting_block(kind, self.lang);
        let is_loop = matches!(kind, "for_statement" | "while_statement");

        if is_nesting {
            self.current_depth += 1;
            if self.current_depth > self.max_depth {
                self.max_depth = self.current_depth;
            }
        }
        if is_loop {
            self.loop_depth += 1;
        }

        // Concurrency primitives
        if let Ok(text) = node.utf8_text(self.source) {
            if is_concurrency_primitive(kind, text, self.lang) {
                self.concurrency += 1;
            }

            // Blocking I/O detection
            if kind == "call_expression" || kind == "call" {
                if is_blocking_io(text, self.lang) {
                    self.blocking_io_count += 1;
                }
                // Allocation detection (flag especially if inside loops)
                if is_allocation(text, self.lang) && self.loop_depth > 0 {
                    self.allocation_in_loop = true;
                }
            }

            // Recursive call detection
            if kind == "call_expression" || kind == "call" {
                if text.contains(&format!("{}(", self.func_name))
                    || text.contains(&format!("self.{}(", self.func_name))
                {
                    self.has_recursive_call = true;
                }
            }
        }

        // Recurse into children
        let child_count = node.child_count();
        for i in 0..child_count {
            if let Some(child) = node.child(i) {
                // Don't recurse into nested function definitions
                if !is_function_node(child.kind(), self.lang) {
                    self.visit(child);
                }
            }
        }

        // Pop nesting tracking
        if is_nesting {
            self.current_depth -= 1;
        }
        if is_loop {
            self.loop_depth -= 1;
        }
    }

    fn into_metrics(self) -> FunctionMetrics {
        let mut m = FunctionMetrics {
            cyclomatic_complexity: self.cyclomatic,
            max_nesting_depth: self.max_depth,
            ast_node_count: self.node_count,
            concurrency_primitives: self.concurrency,
            energy: 0.0,
        };
        m.energy = compute_energy(&m);
        m
    }
}

// =============================================================================
// § 5  File language detection & Security Filters
// =============================================================================

/// Maximum file size scanned by the engine (512 KB).
pub const MAX_FILE_SIZE_BYTES: u64 = 512 * 1024;

/// Check if a file extension is a compiled binary, asset, or archive.
pub fn is_excluded_extension(ext: &str) -> bool {
    matches!(
        ext.to_ascii_lowercase().as_str(),
        "exe" | "dll" | "so" | "dylib" | "bin" | "o" | "a" | "lib" | "class" | "jar" | "war"
        | "pyc" | "pyo" | "pyd" | "wasm"
        | "png" | "jpg" | "jpeg" | "gif" | "bmp" | "ico" | "webp" | "tiff" | "psd" | "raw" | "svg"
        | "mp3" | "mp4" | "wav" | "ogg" | "flac" | "mkv" | "avi" | "mov" | "webm"
        | "zip" | "tar" | "gz" | "bz2" | "xz" | "7z" | "rar" | "iso" | "dmg" | "pkg"
        | "pdf" | "doc" | "docx" | "ppt" | "pptx" | "xls" | "xlsx"
        | "woff" | "woff2" | "ttf" | "eot" | "otf"
        | "db" | "sqlite" | "sqlite3" | "parquet" | "arrow" | "avro"
        | "lock" | "sum" | "map"
    )
}

/// Security: inspect first 512 bytes for null bytes to detect binary files.
pub fn is_binary_file(path: &Path) -> bool {
    use std::io::Read;
    let mut file = match fs::File::open(path) {
        Ok(f) => f,
        Err(_) => return true,
    };
    let mut buf = [0u8; 512];
    let n = match file.read(&mut buf) {
        Ok(n) => n,
        Err(_) => return true,
    };
    if n == 0 {
        return false;
    }
    buf[..n].contains(&0)
}

/// Directory filter: ignore dependency caches, build outputs, and venvs.
pub fn is_ignored_dir(entry: &walkdir::DirEntry) -> bool {
    if !entry.file_type().is_dir() {
        return false;
    }
    let name = entry.file_name().to_string_lossy();
    (name.starts_with('.') && name != "." && name != "..")
        || matches!(
            name.as_ref(),
            "node_modules"
                | "vendor"
                | "target"
                | "dist"
                | "build"
                | "venv"
                | ".venv"
                | "env"
                | "__pycache__"
                | ".git"
                | ".idea"
                | ".vscode"
        )
}

/// Returns the language enum for supported file extensions.
pub fn detect_language(path: &Path) -> Option<Language> {
    let file_name = path.file_name()?.to_str()?;
    if file_name.ends_with(".min.js") || file_name.ends_with(".min.css") {
        return None;
    }
    if file_name == "package-lock.json"
        || file_name == "Cargo.lock"
        || file_name == "yarn.lock"
        || file_name == "pnpm-lock.yaml"
    {
        return None;
    }

    let ext = path.extension().and_then(|e| e.to_str()).unwrap_or("");
    if !ext.is_empty() && is_excluded_extension(ext) {
        return None;
    }

    match ext.to_ascii_lowercase().as_str() {
        "go" => Some(Language::Go),
        "py" | "pyw" => Some(Language::Python),
        _ => None,
    }
}

fn language_name(lang: Language) -> &'static str {
    match lang {
        Language::Go => "Go",
        Language::Python => "Python",
    }
}

// =============================================================================
// § 6  File-level AST scanner
// =============================================================================

/// Analyze a single source file using tree-sitter AST parsing.
///
/// Extracts all function/method definitions, computes structural metrics for
/// each, calculates thermodynamic energy, and returns detected hotspots.
pub fn scan_file(
    path: &Path,
    min_score: f64,
    energy_threshold: f64,
) -> io::Result<Option<FileReport>> {
    // ── Security: File Size Limit (512 KB) ─────────────────────────────────
    if let Ok(meta) = fs::metadata(path) {
        if meta.len() > MAX_FILE_SIZE_BYTES {
            return Ok(None);
        }
    }

    // ── Security: Binary File Safety ───────────────────────────────────────
    if is_binary_file(path) {
        return Ok(None);
    }

    // ── Language detection ──────────────────────────────────────────────────
    let lang = match detect_language(path) {
        Some(l) => l,
        None => return Ok(None),
    };

    // ── Read source ────────────────────────────────────────────────────────
    let source = fs::read(path)?;
    let source_text = String::from_utf8_lossy(&source);

    // Count non-blank, non-comment lines
    let lines_scanned = source_text
        .lines()
        .filter(|line| {
            let trimmed = line.trim();
            !trimmed.is_empty()
                && !trimmed.starts_with('#')
                && !trimmed.starts_with("//")
                && !trimmed.starts_with("/*")
        })
        .count();

    // ── Parse with tree-sitter ─────────────────────────────────────────────
    let mut parser = match create_parser(lang) {
        Some(p) => p,
        None => return Ok(None),
    };

    let tree = match parser.parse(&source, None) {
        Some(t) => t,
        None => return Ok(None),
    };

    let root = tree.root_node();

    // ── Walk root to find function definitions ─────────────────────────────
    let mut hotspots: Vec<Hotspot> = Vec::new();

    fn collect_functions(
        node: tree_sitter::Node,
        source: &[u8],
        lang: Language,
        min_score: f64,
        energy_threshold: f64,
        hotspots: &mut Vec<Hotspot>,
    ) {
        if is_function_node(node.kind(), lang) {
            let func_name = extract_function_name(&node, source, lang);
            let start_line = node.start_position().row + 1; // 1-indexed
            let byte_range = [node.start_byte(), node.end_byte()];

            // Extract first line of function for snippet
            let snippet = node
                .utf8_text(source)
                .ok()
                .and_then(|t| t.lines().next())
                .unwrap_or("")
                .chars()
                .take(120)
                .collect::<String>();

            // Visit the function AST subtree
            let mut visitor = FunctionVisitor::new(source, lang, func_name.clone());

            // Visit children (not the function node itself to avoid re-counting)
            let child_count = node.child_count();
            for i in 0..child_count {
                if let Some(child) = node.child(i) {
                    visitor.visit(child);
                }
            }

            let metrics = visitor.into_metrics();

            if metrics.energy < min_score {
                return;
            }

            // Determine primary entropy driver
            let (vuln_type, description) = determine_primary_driver(
                &metrics,
                visitor_has_recursive_call(&node, source, lang, &func_name),
                visitor_blocking_io_count(&node, source, lang),
                visitor_allocation_in_loop(&node, source, lang),
                energy_threshold,
            );

            hotspots.push(Hotspot {
                function_name: func_name,
                line_number: start_line,
                byte_range,
                source_snippet: snippet,
                entropy_score: metrics.energy,
                vulnerability_type: vuln_type,
                description,
                metrics,
            });
        }

        // Recurse into children to find nested function definitions
        let child_count = node.child_count();
        for i in 0..child_count {
            if let Some(child) = node.child(i) {
                collect_functions(child, source, lang, min_score, energy_threshold, hotspots);
            }
        }
    }

    collect_functions(root, &source, lang, min_score, energy_threshold, &mut hotspots);

    // ── Sort hotspots by energy descending ─────────────────────────────────
    hotspots.sort_by(|a, b| {
        b.entropy_score
            .partial_cmp(&a.entropy_score)
            .unwrap_or(std::cmp::Ordering::Equal)
    });

    // ── Aggregate metrics ──────────────────────────────────────────────────
    let total_entropy: f64 = hotspots.iter().map(|h| h.entropy_score).sum();
    let mean_entropy = if hotspots.is_empty() {
        0.0
    } else {
        (total_entropy / hotspots.len() as f64 * 100.0).round() / 100.0
    };

    Ok(Some(FileReport {
        file_path: path.to_string_lossy().into_owned(),
        language: language_name(lang).to_owned(),
        lines_scanned,
        total_entropy: (total_entropy * 100.0).round() / 100.0,
        mean_hotspot_entropy: mean_entropy,
        hotspots,
    }))
}

/// Helper: re-walk function body to check for recursive calls.
fn visitor_has_recursive_call(
    func_node: &tree_sitter::Node,
    source: &[u8],
    lang: Language,
    func_name: &str,
) -> bool {
    fn check(node: tree_sitter::Node, source: &[u8], lang: Language, name: &str) -> bool {
        let kind = node.kind();
        if kind == "call_expression" || kind == "call" {
            if let Ok(text) = node.utf8_text(source) {
                if text.contains(&format!("{}(", name))
                    || text.contains(&format!("self.{}(", name))
                {
                    return true;
                }
            }
        }
        // Don't descend into nested function defs
        if is_function_node(kind, lang) {
            return false;
        }
        let cc = node.child_count();
        for i in 0..cc {
            if let Some(child) = node.child(i) {
                if check(child, source, lang, name) {
                    return true;
                }
            }
        }
        false
    }
    let cc = func_node.child_count();
    for i in 0..cc {
        if let Some(child) = func_node.child(i) {
            if check(child, source, lang, func_name) {
                return true;
            }
        }
    }
    false
}

/// Helper: count blocking I/O calls inside a function body.
fn visitor_blocking_io_count(
    func_node: &tree_sitter::Node,
    source: &[u8],
    lang: Language,
) -> usize {
    fn count(node: tree_sitter::Node, source: &[u8], lang: Language) -> usize {
        let kind = node.kind();
        let mut total = 0;
        if kind == "call_expression" || kind == "call" {
            if let Ok(text) = node.utf8_text(source) {
                if is_blocking_io(text, lang) {
                    total += 1;
                }
            }
        }
        if is_function_node(kind, lang) {
            return 0; // don't count nested funcs
        }
        let cc = node.child_count();
        for i in 0..cc {
            if let Some(child) = node.child(i) {
                total += count(child, source, lang);
            }
        }
        total
    }
    let cc = func_node.child_count();
    let mut total = 0;
    for i in 0..cc {
        if let Some(child) = func_node.child(i) {
            total += count(child, source, lang);
        }
    }
    total
}

/// Helper: check for allocations inside loops.
fn visitor_allocation_in_loop(
    func_node: &tree_sitter::Node,
    source: &[u8],
    lang: Language,
) -> bool {
    fn check(node: tree_sitter::Node, source: &[u8], lang: Language, in_loop: bool) -> bool {
        let kind = node.kind();
        let is_loop = matches!(kind, "for_statement" | "while_statement");
        let inside = in_loop || is_loop;

        if (kind == "call_expression" || kind == "call") && inside {
            if let Ok(text) = node.utf8_text(source) {
                if is_allocation(text, lang) {
                    return true;
                }
            }
        }
        if is_function_node(kind, lang) {
            return false;
        }
        let cc = node.child_count();
        for i in 0..cc {
            if let Some(child) = node.child(i) {
                if check(child, source, lang, inside) {
                    return true;
                }
            }
        }
        false
    }
    let cc = func_node.child_count();
    for i in 0..cc {
        if let Some(child) = func_node.child(i) {
            if check(child, source, lang, false) {
                return true;
            }
        }
    }
    false
}

/// Determine the primary entropy driver based on computed metrics.
fn determine_primary_driver(
    metrics: &FunctionMetrics,
    has_recursive: bool,
    blocking_io: usize,
    alloc_in_loop: bool,
    threshold: f64,
) -> (VulnerabilityType, String) {
    // Priority ordering: highest-impact issues first
    if has_recursive {
        return (
            VulnerabilityType::RecursiveCall,
            format!(
                "Recursive call detected — stack-depth risk (M={}, D={})",
                metrics.cyclomatic_complexity, metrics.max_nesting_depth
            ),
        );
    }
    if metrics.concurrency_primitives > 0 && metrics.max_nesting_depth >= 2 {
        return (
            VulnerabilityType::SyncContention,
            format!(
                "Synchronization primitive at nesting depth {} — lock contention risk (C={})",
                metrics.max_nesting_depth, metrics.concurrency_primitives
            ),
        );
    }
    if blocking_io > 0 && metrics.max_nesting_depth >= 1 {
        return (
            VulnerabilityType::BlockingIO,
            format!(
                "Blocking I/O call(s) ({}) inside control flow depth {} — latency spike risk",
                blocking_io, metrics.max_nesting_depth
            ),
        );
    }
    if alloc_in_loop {
        return (
            VulnerabilityType::HotAllocation,
            format!(
                "Heap allocation inside loop body — O(n) memory pressure (S={})",
                metrics.ast_node_count
            ),
        );
    }
    if metrics.max_nesting_depth >= 4 {
        return (
            VulnerabilityType::DeepNesting,
            format!(
                "Nesting depth {} exceeds safe threshold — O(n^k) complexity risk (M={})",
                metrics.max_nesting_depth, metrics.cyclomatic_complexity
            ),
        );
    }
    if metrics.cyclomatic_complexity >= 10 {
        return (
            VulnerabilityType::CognitiveBranch,
            format!(
                "Cyclomatic complexity {} — high cognitive load and branch exhaustion risk",
                metrics.cyclomatic_complexity
            ),
        );
    }
    if metrics.energy >= threshold {
        return (
            VulnerabilityType::HighEnergy,
            format!(
                "Structural energy {:.2} exceeds threshold {:.2} (M={}, D={}, S={}, C={})",
                metrics.energy,
                threshold,
                metrics.cyclomatic_complexity,
                metrics.max_nesting_depth,
                metrics.ast_node_count,
                metrics.concurrency_primitives
            ),
        );
    }

    // Default — low energy, informational
    (
        VulnerabilityType::CognitiveBranch,
        format!(
            "Structural energy {:.2} — nominal (M={}, D={})",
            metrics.energy, metrics.cyclomatic_complexity, metrics.max_nesting_depth
        ),
    )
}

// =============================================================================
// § 7  Directory walker
// =============================================================================

/// Walk `root` recursively, scan every supported source file, and collect
/// `FileReport` values. Uses Rayon for parallel execution when enabled.
pub fn scan_directory(
    root: &Path,
    min_score: f64,
    energy_threshold: f64,
    verbose: bool,
) -> Vec<FileReport> {
    let candidates: Vec<PathBuf> = WalkDir::new(root)
        .follow_links(false)
        .into_iter()
        .filter_entry(|e| !is_ignored_dir(e))
        .filter_map(|e| e.ok())
        .filter(|e| e.file_type().is_file())
        .map(|e| e.into_path())
        .filter(|p| {
            if let Ok(meta) = fs::metadata(p) {
                if meta.len() > MAX_FILE_SIZE_BYTES {
                    return false;
                }
            }
            if is_binary_file(p) {
                return false;
            }
            detect_language(p).is_some()
        })
        .collect();

    if verbose {
        println!(
            "{} {} source files to analyze …",
            "→".cyan().bold(),
            candidates.len()
        );
    }

    // ── Parallel scan (Rayon) ──────────────────────────────────────────────
    #[cfg(feature = "parallel")]
    let reports: Vec<FileReport> = {
        use rayon::iter::IntoParallelIterator;
        candidates
            .into_par_iter()
            .filter_map(|path| {
                if verbose {
                    println!("  {} {}", "⚡".yellow(), path.display());
                }
                match scan_file(&path, min_score, energy_threshold) {
                    Ok(Some(report)) => Some(report),
                    Ok(None) => None,
                    Err(e) => {
                        eprintln!("{} {}: {}", "ERR".red().bold(), path.display(), e);
                        None
                    }
                }
            })
            .collect()
    };

    // ── Sequential fallback ────────────────────────────────────────────────
    #[cfg(not(feature = "parallel"))]
    let reports: Vec<FileReport> = candidates
        .into_iter()
        .filter_map(|path| {
            if verbose {
                println!("  {} {}", "→", path.display());
            }
            match scan_file(&path, min_score, energy_threshold) {
                Ok(Some(report)) => Some(report),
                Ok(None) => None,
                Err(e) => {
                    eprintln!("ERR {}: {}", path.display(), e);
                    None
                }
            }
        })
        .collect();

    reports
}

// =============================================================================
// § 8  Report builder & JSON serializer
// =============================================================================

/// Assemble the root `ThermodynamicReport`, sort by entropy, and write JSON.
pub fn build_and_write_report(
    mut file_reports: Vec<FileReport>,
    scanned_dir: &Path,
    output_path: &Path,
) -> io::Result<ThermodynamicReport> {
    file_reports.sort_by(|a, b| {
        b.total_entropy
            .partial_cmp(&a.total_entropy)
            .unwrap_or(std::cmp::Ordering::Equal)
    });

    let total_hotspots = file_reports.iter().map(|r| r.hotspots.len()).sum();
    let global_entropy = file_reports.iter().map(|r| r.total_entropy).sum::<f64>();

    let report = ThermodynamicReport {
        engine_version: env!("CARGO_PKG_VERSION").to_owned(),
        generated_at: chrono::Utc::now().to_rfc3339(),
        scanned_directory: scanned_dir.to_string_lossy().into_owned(),
        files_analyzed: file_reports.len(),
        total_hotspots,
        global_entropy: (global_entropy * 100.0).round() / 100.0,
        file_reports,
    };

    let json = serde_json::to_string_pretty(&report)
        .map_err(|e| io::Error::new(io::ErrorKind::Other, e))?;

    fs::write(output_path, json)?;

    Ok(report)
}

// =============================================================================
// § 9  Terminal summary banner
// =============================================================================

fn print_summary(report: &ThermodynamicReport) {
    println!();
    println!(
        "{}",
        "╔══════════════════════════════════════════════════════╗".cyan()
    );
    println!(
        "{}",
        "║   Thermodynamic AST Engine v2 — Summary             ║".cyan()
    );
    println!(
        "{}",
        "║   (Tree-sitter AST-driven structural analysis)      ║".cyan()
    );
    println!(
        "{}",
        "╚══════════════════════════════════════════════════════╝".cyan()
    );
    println!(
        "  Files analyzed   : {}",
        report.files_analyzed.to_string().yellow()
    );
    println!(
        "  Total hotspots   : {}",
        report.total_hotspots.to_string().yellow()
    );
    println!(
        "  Global entropy   : {}",
        format!("{:.2}", report.global_entropy).red().bold()
    );

    println!();
    println!("{}", "  Top 5 Energy Hotspots:".bold());
    println!(
        "  {:<8}  {:<25}  {:<6}  {:<6}  {:<6}  {}",
        "Energy", "Function", "M", "D", "S", "Driver"
    );
    println!("  {}", "─".repeat(78));

    // Flatten and collect top 5 across all files
    let mut all_hotspots: Vec<&Hotspot> = report
        .file_reports
        .iter()
        .flat_map(|r| r.hotspots.iter())
        .collect();

    all_hotspots.sort_by(|a, b| {
        b.entropy_score
            .partial_cmp(&a.entropy_score)
            .unwrap_or(std::cmp::Ordering::Equal)
    });

    for h in all_hotspots.iter().take(5) {
        println!(
            "  {:<8.2}  {:<25}  {:<6}  {:<6}  {:<6}  {}",
            h.entropy_score,
            &h.function_name,
            h.metrics.cyclomatic_complexity,
            h.metrics.max_nesting_depth,
            h.metrics.ast_node_count,
            h.vulnerability_type.to_string().red()
        );
    }

    println!();
}

// =============================================================================
// § 10  Entry point
// =============================================================================

fn main() {
    let cli = Cli::parse();

    // ── Validate input directory ───────────────────────────────────────────
    if !cli.directory.exists() {
        eprintln!(
            "{} Directory '{}' does not exist.",
            "ERROR:".red().bold(),
            cli.directory.display()
        );
        std::process::exit(1);
    }
    if !cli.directory.is_dir() {
        eprintln!(
            "{} '{}' is not a directory.",
            "ERROR:".red().bold(),
            cli.directory.display()
        );
        std::process::exit(1);
    }

    println!(
        "\n{} Scanning: {}\n",
        "▶".green().bold(),
        cli.directory.display().to_string().cyan()
    );

    // ── Scan ───────────────────────────────────────────────────────────────
    let file_reports = scan_directory(
        &cli.directory,
        cli.min_score,
        cli.energy_threshold,
        cli.verbose,
    );

    if file_reports.is_empty() {
        println!(
            "{} No supported source files (Go/Python) found in '{}'.",
            "WARN:".yellow().bold(),
            cli.directory.display()
        );
        let _ = build_and_write_report(Vec::new(), &cli.directory, &cli.output);
        std::process::exit(0);
    }

    // ── Build & persist report ─────────────────────────────────────────────
    let report = match build_and_write_report(file_reports, &cli.directory, &cli.output) {
        Ok(r) => r,
        Err(e) => {
            eprintln!("{} Failed to write report: {}", "ERROR:".red().bold(), e);
            std::process::exit(1);
        }
    };

    // ── Print summary banner ───────────────────────────────────────────────
    print_summary(&report);

    println!(
        "{} Report written to: {}\n",
        "✓".green().bold(),
        cli.output.display().to_string().green()
    );
}

// =============================================================================
// § 11  Unit tests
// =============================================================================

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_energy_formula_basic() {
        let m = FunctionMetrics {
            cyclomatic_complexity: 10,
            max_nesting_depth: 4,
            ast_node_count: 128,
            concurrency_primitives: 2,
            energy: 0.0,
        };
        let e = compute_energy(&m);
        // E = 0.45*10 + 0.25*4 + 0.15*log2(128) + 0.15*2
        // E = 4.5 + 1.0 + 0.15*7 + 0.3 = 4.5 + 1.0 + 1.05 + 0.3 = 6.85
        assert!((e - 6.85).abs() < 0.1, "Expected ~6.85, got {}", e);
    }

    #[test]
    fn test_energy_zero_nodes() {
        let m = FunctionMetrics {
            cyclomatic_complexity: 1,
            max_nesting_depth: 0,
            ast_node_count: 0,
            concurrency_primitives: 0,
            energy: 0.0,
        };
        let e = compute_energy(&m);
        // E = 0.45*1 + 0 + 0 + 0 = 0.45
        assert!((e - 0.45).abs() < 0.01, "Expected ~0.45, got {}", e);
    }

    #[test]
    fn test_language_detection() {
        assert_eq!(detect_language(Path::new("foo.py")), Some(Language::Python));
        assert_eq!(detect_language(Path::new("bar.go")), Some(Language::Go));
        assert_eq!(detect_language(Path::new("baz.rs")), None); // Only Go/Python now
        assert_eq!(detect_language(Path::new("image.png")), None);
        assert_eq!(detect_language(Path::new("package-lock.json")), None);
    }

    #[test]
    fn test_is_binary_file_detection() {
        let temp_dir = std::env::temp_dir();
        let text_file = temp_dir.join("test_ast_plain_text.txt");
        let bin_file = temp_dir.join("test_ast_binary_blob.bin");

        let _ = fs::write(&text_file, "Hello, this is a plain text file!\nNo null bytes here.");
        let _ = fs::write(&bin_file, b"Hello\x00Binary\x00Data");

        assert!(!is_binary_file(&text_file), "Text file should not be binary");
        assert!(is_binary_file(&bin_file), "File with nulls must be binary");

        let _ = fs::remove_file(&text_file);
        let _ = fs::remove_file(&bin_file);
    }

    #[test]
    fn test_go_function_parsing() {
        let source = br#"
package main

import "fmt"

func ProcessData(items []int) int {
    total := 0
    for _, item := range items {
        if item > 0 {
            total += item
        }
    }
    return total
}

func SimpleFunc() {
    fmt.Println("hello")
}
"#;
        let mut parser = create_parser(Language::Go).unwrap();
        let tree = parser.parse(source.as_slice(), None).unwrap();
        let root = tree.root_node();

        // Count function declarations
        let mut func_count = 0;
        fn count_funcs(node: tree_sitter::Node, count: &mut usize) {
            if is_function_node(node.kind(), Language::Go) {
                *count += 1;
            }
            let cc = node.child_count();
            for i in 0..cc {
                if let Some(child) = node.child(i) {
                    count_funcs(child, count);
                }
            }
        }
        count_funcs(root, &mut func_count);
        assert_eq!(func_count, 2, "Expected 2 Go functions");
    }

    #[test]
    fn test_python_function_parsing() {
        let source = br#"
import time

def fetch_data(urls):
    results = []
    for url in urls:
        if url.startswith("http"):
            results.append(url)
    return results

def simple():
    pass
"#;
        let mut parser = create_parser(Language::Python).unwrap();
        let tree = parser.parse(source.as_slice(), None).unwrap();
        let root = tree.root_node();

        let mut func_count = 0;
        fn count_funcs(node: tree_sitter::Node, count: &mut usize) {
            if is_function_node(node.kind(), Language::Python) {
                *count += 1;
            }
            let cc = node.child_count();
            for i in 0..cc {
                if let Some(child) = node.child(i) {
                    count_funcs(child, count);
                }
            }
        }
        count_funcs(root, &mut func_count);
        assert_eq!(func_count, 2, "Expected 2 Python functions");
    }

    #[test]
    fn test_go_cyclomatic_complexity() {
        let source = br#"
package main

func Complex(x int) int {
    if x > 0 {
        if x > 10 {
            for i := 0; i < x; i++ {
                if i % 2 == 0 {
                    return i
                }
            }
        }
    }
    return 0
}
"#;
        let mut parser = create_parser(Language::Go).unwrap();
        let tree = parser.parse(source.as_slice(), None).unwrap();
        let root = tree.root_node();

        // Find the function and visit it
        fn find_func(node: tree_sitter::Node) -> Option<tree_sitter::Node> {
            if is_function_node(node.kind(), Language::Go) {
                return Some(node);
            }
            let cc = node.child_count();
            for i in 0..cc {
                if let Some(child) = node.child(i) {
                    if let Some(found) = find_func(child) {
                        return Some(found);
                    }
                }
            }
            None
        }

        let func_node = find_func(root).expect("Should find Complex function");
        let func_name = extract_function_name(&func_node, source, Language::Go);
        assert_eq!(func_name, "Complex");

        let mut visitor = FunctionVisitor::new(source, Language::Go, func_name);
        let cc = func_node.child_count();
        for i in 0..cc {
            if let Some(child) = func_node.child(i) {
                visitor.visit(child);
            }
        }
        let metrics = visitor.into_metrics();

        // 3 if + 1 for = 4 decision points + 1 base = 5
        assert!(
            metrics.cyclomatic_complexity >= 4,
            "Expected cyclomatic >= 4, got {}",
            metrics.cyclomatic_complexity
        );
        assert!(
            metrics.max_nesting_depth >= 3,
            "Expected max_depth >= 3, got {}",
            metrics.max_nesting_depth
        );
        assert!(metrics.energy > 0.0, "Energy should be positive");
    }

    #[test]
    fn test_scan_test_sample_go() {
        let path = Path::new("test_samples/crawler.go");
        if !path.exists() {
            return; // skip if not in the right directory
        }
        let report = scan_file(path, 0.0, 15.0).unwrap();
        assert!(report.is_some(), "crawler.go should produce a report");
        let r = report.unwrap();
        assert!(!r.hotspots.is_empty(), "crawler.go should have hotspots");
        assert!(r.total_entropy > 0.0, "Total entropy should be positive");
    }

    #[test]
    fn test_scan_test_sample_python() {
        let path = Path::new("test_samples/data_pipeline.py");
        if !path.exists() {
            return;
        }
        let report = scan_file(path, 0.0, 15.0).unwrap();
        assert!(report.is_some(), "data_pipeline.py should produce a report");
        let r = report.unwrap();
        assert!(!r.hotspots.is_empty(), "Should have hotspots");
    }
}
