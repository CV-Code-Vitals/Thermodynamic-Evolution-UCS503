// =============================================================================
// thermodynamic-ast-engine · src/main.rs
//
// A heuristic "thermodynamic entropy" analyzer for Go and Python source files.
//
// Design pillars
// ──────────────
//   1. Zero unsafe code.
//   2. Compiled regexes are built exactly once (once_cell::sync::Lazy).
//   3. File scanning runs in parallel via Rayon (feature-gated).
//   4. All public data types implement Serialize so the JSON report is trivial.
//   5. Every logical stage is a separate module for testability.
// =============================================================================

// ── External crate imports ────────────────────────────────────────────────────
use clap::Parser;
use colored::Colorize;
use once_cell::sync::Lazy;
use regex::Regex;
use serde::{Deserialize, Serialize};
use std::{
    fs,
    io::{self, BufRead},
    path::{Path, PathBuf},
};
use walkdir::WalkDir;

#[cfg(feature = "parallel")]
use rayon::prelude::*;

// =============================================================================
// § 1  CLI argument definition (clap derive)
// =============================================================================

/// Thermodynamic AST Engine — identifies entropy hotspots in your codebase
#[derive(Parser, Debug)]
#[command(
    name        = "thermodynamic-ast-engine",
    version     = env!("CARGO_PKG_VERSION"),
    author      = env!("CARGO_PKG_AUTHORS"),
    about       = "Calculates thermodynamic entropy scores for Go/Python source files \
                   and emits a JSON report of scaling bottlenecks.",
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

    /// Minimum entropy score to include in the report (0.0 – 100.0)
    #[arg(short, long, value_name = "SCORE", default_value_t = 0.0)]
    min_score: f64,

    /// Show verbose per-file progress in stdout
    #[arg(short, long)]
    verbose: bool,
}

// =============================================================================
// § 2  Core data model
// =============================================================================

/// The vulnerability category detected by the heuristic engine.
/// Each variant maps to a distinct set of regex patterns.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub enum VulnerabilityType {
    /// Deeply nested loops (O(n^k) risk)
    DeepNesting,
    /// Direct or mutual recursion without obvious base case
    RecursiveCall,
    /// Heap allocation inside a hot loop
    HotAllocation,
    /// Blocking I/O or syscall on the critical path
    BlockingIO,
    /// High cognitive complexity (many boolean operators / branches)
    CognitiveBranch,
    /// Unsafe synchronisation primitive (mutex inside loop, etc.)
    SyncContention,
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
        }
    }
}

/// A single entropy "hotspot" — one detected signal within a file.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Hotspot {
    /// Name of the containing function / method (best-effort heuristic)
    pub function_name: String,

    /// 1-indexed line number where the pattern was detected
    pub line_number: usize,

    /// The raw source line that triggered the signal
    pub source_snippet: String,

    /// Weighted entropy score for this single signal (0.0 – 100.0)
    pub entropy_score: f64,

    /// Classification of the detected risk
    pub vulnerability_type: VulnerabilityType,

    /// Human-readable explanation of why this is flagged
    pub description: String,
}

/// Aggregated report for one source file.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FileReport {
    /// Absolute path to the source file
    pub file_path: String,

    /// Programming language detected from the file extension
    pub language: String,

    /// Total number of non-blank, non-comment lines scanned
    pub lines_scanned: usize,

    /// Sum of all individual hotspot entropy scores
    pub total_entropy: f64,

    /// Mean entropy per detected hotspot (0 if no hotspots)
    pub mean_hotspot_entropy: f64,

    /// All identified hotspots, sorted highest-score first
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
// § 3  Regex pattern registry (compiled once, shared across threads)
// =============================================================================

/// Internal representation of one pattern rule.
pub struct PatternRule {
    pub regex: &'static Lazy<Regex>,
    pub vulnerability: VulnerabilityType,
    pub base_score: f64, // base entropy contribution per match
    pub description_tmpl: &'static str,
}

// ── Python patterns ───────────────────────────────────────────────────────────

static PY_FOR_WHILE: Lazy<Regex> = Lazy::new(|| Regex::new(r"^\s*(for|while)\s+").unwrap());
// Note: the `regex` crate does not support backreferences; we detect recursion
// via two independent signals:
//   1. `self.method()` — object calling its own method (Python)
//   2. A standalone identifier call on its own line that is NOT a dotted method
//      call (e.g. `flatten(items)` rather than `obj.flatten(items)`)
static PY_RECURSIVE: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"\bself\s*\.\s*\w+\s*\(|(?:^|\s)(\w+)\s*\([^)]*\)\s*$").unwrap());
static PY_ALLOC: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"\b(list|dict|set|bytearray|numpy\.zeros|numpy\.ones|np\.zeros|np\.ones|torch\.zeros|torch\.ones)\s*[\(\[]").unwrap()
});
static PY_BLOCKING_IO: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"\b(open\s*\(|requests\.(get|post|put|delete|patch)|urllib|subprocess\.(call|run|Popen)|time\.sleep|socket\.recv|socket\.accept)\b").unwrap()
});
static PY_BRANCH: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"\b(if|elif|and|or|not|assert)\b").unwrap());
static PY_MUTEX: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"\b(threading\.(Lock|RLock|Semaphore)|asyncio\.Lock|multiprocessing\.Lock)\b")
        .unwrap()
});
static PY_FUNC: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"^\s*(?:async\s+)?def\s+(\w+)\s*\(").unwrap());

// ── Go patterns ───────────────────────────────────────────────────────────────

static GO_FOR: Lazy<Regex> = Lazy::new(|| Regex::new(r"^\s*for\s+").unwrap());
// Match a bare function call (not a dotted method call like obj.Method()).
// The negative lookbehind equivalent in `regex` isn't supported, so we anchor
// on the pattern starting after whitespace or at line start, without a dot.
static GO_RECURSIVE: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"(?:^|[\s,=(])([A-Z]\w*)\s*\(").unwrap()); // Capital = exported fn, likely recursive
static GO_ALLOC: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"\bmake\s*\(|\bnew\s*\(|\[\][\w\*]+\{").unwrap());
static GO_BLOCKING_IO: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"\b(os\.Open|os\.Create|ioutil\.(ReadFile|WriteFile)|http\.(Get|Post)|net\.Dial|time\.Sleep|bufio\.NewReader|sql\.Open)\b").unwrap()
});
static GO_BRANCH: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"\b(if|else|switch|case|&&|\|\||select)\b").unwrap());
static GO_MUTEX: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"\b(sync\.(Mutex|RWMutex|WaitGroup|Once)|atomic\.(Add|Load|Store|Swap))").unwrap()
});
static GO_FUNC: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"^\s*func\s+(?:\([^)]+\)\s+)?(\w+)\s*\(").unwrap());

// ── JavaScript / TypeScript patterns ──────────────────────────────────────────

static JS_FOR_WHILE: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"\b(for\s*\(|while\s*\(|for\s+await|\.forEach\(|\.map\()").unwrap());
static JS_RECURSIVE: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"(?:^|[\s,=(])([a-zA-Z_$][\w$]*)\s*\(").unwrap());
static JS_ALLOC: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"\b(new\s+(Array|Buffer|Uint8Array|Map|Set|Object)|Array\.from)\b").unwrap());
static JS_BLOCKING_IO: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"\b(fs\.(readFileSync|writeFileSync|appendFileSync|existsSync)|execSync|spawnSync|Atomics\.wait)\b").unwrap()
});
static JS_BRANCH: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"\b(if|else|switch|case|catch|&&|\|\|)\b").unwrap());
static JS_MUTEX: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"\b(Mutex|Semaphore|Lock|AsyncLock|atomics)\b").unwrap());
static JS_FUNC: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"^\s*(?:export\s+)?(?:async\s+)?function\s+(\w+)|(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s*)?\(").unwrap()
});

// ── Rust patterns ─────────────────────────────────────────────────────────────

static RUST_FOR_WHILE: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"^\s*(for\s+\w+\s+in|while\s+|loop\s*\{)").unwrap());
static RUST_RECURSIVE: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"(?:^|[\s,=(])([a-zA-Z_]\w*)\s*\(").unwrap());
static RUST_ALLOC: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"\b(Vec::with_capacity|Vec::new|Box::new|String::from|vec!\[)").unwrap());
static RUST_BLOCKING_IO: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"\b(std::fs::|File::open|File::create|thread::sleep|TcpStream::connect|Command::new)\b").unwrap()
});
static RUST_BRANCH: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"\b(if|else|match|&&|\|\|)\b").unwrap());
static RUST_MUTEX: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"\b(Mutex::|RwLock::|AtomicBool|AtomicUsize|Arc::new|mpsc::channel|barrier)\b").unwrap()
});
static RUST_FUNC: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"^\s*(?:pub(?:\([^)]+\))?\s+)?(?:async\s+)?fn\s+(\w+)").unwrap());

// ── C / C++ patterns ──────────────────────────────────────────────────────────

static C_FOR_WHILE: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"^\s*(for\s*\(|while\s*\(|do\s*\{)").unwrap());
static C_RECURSIVE: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"(?:^|[\s,=(])([a-zA-Z_]\w*)\s*\(").unwrap());
static C_ALLOC: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"\b(malloc\s*\(|calloc\s*\(|realloc\s*\(|new\s+\w+)").unwrap());
static C_BLOCKING_IO: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"\b(fopen|fread|fwrite|sleep|usleep|recv|send|connect|system)\b").unwrap()
});
static C_BRANCH: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"\b(if|else|switch|case|&&|\|\|)\b").unwrap());
static C_MUTEX: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"\b(pthread_mutex_|std::mutex|std::lock_guard|std::unique_lock|atomic)\b").unwrap()
});
static C_FUNC: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"^\s*(?:[\w:*&<>]+\s+)+(\w+)\s*\([^;]*\)\s*\{?$").unwrap());

// ── Java / C# / Kotlin / Scala patterns ───────────────────────────────────────

static JAVA_FOR_WHILE: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"^\s*(for\s*\(|while\s*\(|foreach\s*\()").unwrap());
static JAVA_RECURSIVE: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"(?:^|[\s,=(])([a-zA-Z_]\w*)\s*\(").unwrap());
static JAVA_ALLOC: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"\b(new\s+\w+\[|new\s+ArrayList|new\s+HashMap|new\s+byte\[)").unwrap());
static JAVA_BLOCKING_IO: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"\b(Thread\.sleep|FileInputStream|FileOutputStream|Socket|HttpClient|File\.read)\b").unwrap()
});
static JAVA_BRANCH: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"\b(if|else|switch|case|catch|&&|\|\|)\b").unwrap());
static JAVA_MUTEX: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"\b(synchronized|ReentrantLock|Semaphore|CountDownLatch|Monitor\.Enter)\b").unwrap()
});
static JAVA_FUNC: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"^\s*(?:public|private|protected|static|final|native|synchronized|abstract|\s)+[\w<>\[\]]+\s+(\w+)\s*\(").unwrap()
});

// ── Shell patterns ────────────────────────────────────────────────────────────

static SH_LOOP: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"^\s*(for\s+\w+\s+in|while\s+|until\s+)").unwrap());
static SH_BLOCKING_IO: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"\b(sleep|curl|wget|scp|rsync|ssh|nc|netcat)\b").unwrap()
});
static SH_BRANCH: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"\b(if\s+|elif\s+|then|else|case\s+|&&|\|\|)").unwrap());
static SH_FUNC: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"^\s*(?:function\s+)?(\w+)\s*\(\)").unwrap());

// ── Config & Data patterns ────────────────────────────────────────────────────

static CONFIG_SECRET: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r#"(?i)\b(password|secret|api_key|token|private_key)\s*[:=]\s*['"][^'"]{4,}"#).unwrap()
});
static CONFIG_BLOCKING: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"(?i)\b(timeout\s*:\s*0|unlimited|keep_alive|max_connections\s*:\s*\d{5,})").unwrap()
});
static CONFIG_BRANCH: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"^\s*-\s+name:|\b(when|condition|assert):").unwrap()
});

// ── Generic / Fallback patterns ───────────────────────────────────────────────

static GENERIC_LOOP: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"\b(for\s+|while\s+|loop\s*\{|repeat\s+)").unwrap());
static GENERIC_RECURSIVE: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"(?:^|[\s,=(])([a-zA-Z_]\w*)\s*\(").unwrap());
static GENERIC_ALLOC: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"\b(malloc|alloc|new\s+\w+|clone\(\))\b").unwrap());
static GENERIC_BLOCKING: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"\b(sleep|delay|wait|read|write|connect|recv|send)\s*\(").unwrap());
static GENERIC_BRANCH: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"\b(if|else|switch|case|when|unless|catch)\b").unwrap());
static GENERIC_MUTEX: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"\b(mutex|lock|semaphore|atomic|critical_section)\b").unwrap());
static GENERIC_FUNC: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"^\s*(?:def|func|fn|function|sub|void|int|bool|string)\s+(\w+)").unwrap());

// =============================================================================
// § 4  Language-specific rule tables
// =============================================================================

fn python_rules() -> Vec<PatternRule> {
    vec![
        PatternRule {
            regex: &PY_FOR_WHILE,
            vulnerability: VulnerabilityType::DeepNesting,
            base_score: 15.0,
            description_tmpl: "Loop construct detected - nesting depth multiplier applied",
        },
        PatternRule {
            regex: &PY_RECURSIVE,
            vulnerability: VulnerabilityType::RecursiveCall,
            base_score: 20.0,
            description_tmpl: "Possible recursive invocation - stack-depth risk",
        },
        PatternRule {
            regex: &PY_ALLOC,
            vulnerability: VulnerabilityType::HotAllocation,
            base_score: 12.0,
            description_tmpl: "Heap allocation inside potentially hot path",
        },
        PatternRule {
            regex: &PY_BLOCKING_IO,
            vulnerability: VulnerabilityType::BlockingIO,
            base_score: 18.0,
            description_tmpl: "Blocking I/O call on critical path - latency spike risk",
        },
        PatternRule {
            regex: &PY_BRANCH,
            vulnerability: VulnerabilityType::CognitiveBranch,
            base_score: 5.0,
            description_tmpl: "Branch/boolean operator increases cyclomatic complexity",
        },
        PatternRule {
            regex: &PY_MUTEX,
            vulnerability: VulnerabilityType::SyncContention,
            base_score: 22.0,
            description_tmpl: "Synchronisation primitive - potential lock contention hotspot",
        },
    ]
}

fn go_rules() -> Vec<PatternRule> {
    vec![
        PatternRule {
            regex: &GO_FOR,
            vulnerability: VulnerabilityType::DeepNesting,
            base_score: 15.0,
            description_tmpl: "Go for-loop - nesting depth multiplier applied",
        },
        PatternRule {
            regex: &GO_RECURSIVE,
            vulnerability: VulnerabilityType::RecursiveCall,
            base_score: 20.0,
            description_tmpl: "Exported function call - checked for self-recursion",
        },
        PatternRule {
            regex: &GO_ALLOC,
            vulnerability: VulnerabilityType::HotAllocation,
            base_score: 12.0,
            description_tmpl: "make/new/slice-literal allocation in hot path",
        },
        PatternRule {
            regex: &GO_BLOCKING_IO,
            vulnerability: VulnerabilityType::BlockingIO,
            base_score: 18.0,
            description_tmpl: "Blocking stdlib I/O call - goroutine contention risk",
        },
        PatternRule {
            regex: &GO_BRANCH,
            vulnerability: VulnerabilityType::CognitiveBranch,
            base_score: 5.0,
            description_tmpl: "Branch/boolean expression increases cyclomatic complexity",
        },
        PatternRule {
            regex: &GO_MUTEX,
            vulnerability: VulnerabilityType::SyncContention,
            base_score: 22.0,
            description_tmpl: "sync.Mutex/atomic - potential throughput bottleneck",
        },
    ]
}

fn js_ts_rules() -> Vec<PatternRule> {
    vec![
        PatternRule {
            regex: &JS_FOR_WHILE,
            vulnerability: VulnerabilityType::DeepNesting,
            base_score: 15.0,
            description_tmpl: "JavaScript/TypeScript loop or iterator in hot path",
        },
        PatternRule {
            regex: &JS_RECURSIVE,
            vulnerability: VulnerabilityType::RecursiveCall,
            base_score: 20.0,
            description_tmpl: "Recursive call detected - call stack exhaustion risk",
        },
        PatternRule {
            regex: &JS_ALLOC,
            vulnerability: VulnerabilityType::HotAllocation,
            base_score: 12.0,
            description_tmpl: "Heap allocation inside loop or frequent execution path",
        },
        PatternRule {
            regex: &JS_BLOCKING_IO,
            vulnerability: VulnerabilityType::BlockingIO,
            base_score: 22.0,
            description_tmpl: "Synchronous blocking I/O freezes the Node event loop",
        },
        PatternRule {
            regex: &JS_BRANCH,
            vulnerability: VulnerabilityType::CognitiveBranch,
            base_score: 5.0,
            description_tmpl: "Branching statement increases cyclomatic complexity",
        },
        PatternRule {
            regex: &JS_MUTEX,
            vulnerability: VulnerabilityType::SyncContention,
            base_score: 18.0,
            description_tmpl: "Synchronization primitive / shared array contention",
        },
    ]
}

fn rust_rules() -> Vec<PatternRule> {
    vec![
        PatternRule {
            regex: &RUST_FOR_WHILE,
            vulnerability: VulnerabilityType::DeepNesting,
            base_score: 15.0,
            description_tmpl: "Rust loop construct - nesting depth multiplier applied",
        },
        PatternRule {
            regex: &RUST_RECURSIVE,
            vulnerability: VulnerabilityType::RecursiveCall,
            base_score: 20.0,
            description_tmpl: "Self-recursion detected - stack depth overhead risk",
        },
        PatternRule {
            regex: &RUST_ALLOC,
            vulnerability: VulnerabilityType::HotAllocation,
            base_score: 12.0,
            description_tmpl: "Heap vector or Box allocation on critical path",
        },
        PatternRule {
            regex: &RUST_BLOCKING_IO,
            vulnerability: VulnerabilityType::BlockingIO,
            base_score: 18.0,
            description_tmpl: "Blocking I/O or sleep in async/worker context",
        },
        PatternRule {
            regex: &RUST_BRANCH,
            vulnerability: VulnerabilityType::CognitiveBranch,
            base_score: 5.0,
            description_tmpl: "Pattern match / branching adds cyclomatic branches",
        },
        PatternRule {
            regex: &RUST_MUTEX,
            vulnerability: VulnerabilityType::SyncContention,
            base_score: 22.0,
            description_tmpl: "Mutex / RwLock lock acquisition contention hotspot",
        },
    ]
}

fn c_cpp_rules() -> Vec<PatternRule> {
    vec![
        PatternRule {
            regex: &C_FOR_WHILE,
            vulnerability: VulnerabilityType::DeepNesting,
            base_score: 15.0,
            description_tmpl: "C/C++ loop construct - nesting multiplier applied",
        },
        PatternRule {
            regex: &C_RECURSIVE,
            vulnerability: VulnerabilityType::RecursiveCall,
            base_score: 20.0,
            description_tmpl: "Function call checked for recursion",
        },
        PatternRule {
            regex: &C_ALLOC,
            vulnerability: VulnerabilityType::HotAllocation,
            base_score: 14.0,
            description_tmpl: "Dynamic memory allocation (malloc/new) inside hot path",
        },
        PatternRule {
            regex: &C_BLOCKING_IO,
            vulnerability: VulnerabilityType::BlockingIO,
            base_score: 18.0,
            description_tmpl: "Blocking POSIX/C-runtime I/O operation",
        },
        PatternRule {
            regex: &C_BRANCH,
            vulnerability: VulnerabilityType::CognitiveBranch,
            base_score: 5.0,
            description_tmpl: "Conditional branch / switch increases cyclomatic entropy",
        },
        PatternRule {
            regex: &C_MUTEX,
            vulnerability: VulnerabilityType::SyncContention,
            base_score: 22.0,
            description_tmpl: "POSIX/std::mutex synchronization primitive contention",
        },
    ]
}

fn java_csharp_rules() -> Vec<PatternRule> {
    vec![
        PatternRule {
            regex: &JAVA_FOR_WHILE,
            vulnerability: VulnerabilityType::DeepNesting,
            base_score: 15.0,
            description_tmpl: "Loop construct - nesting depth multiplier applied",
        },
        PatternRule {
            regex: &JAVA_RECURSIVE,
            vulnerability: VulnerabilityType::RecursiveCall,
            base_score: 20.0,
            description_tmpl: "Method call checked for recursive depth",
        },
        PatternRule {
            regex: &JAVA_ALLOC,
            vulnerability: VulnerabilityType::HotAllocation,
            base_score: 12.0,
            description_tmpl: "Object / collection instantiation in loop path",
        },
        PatternRule {
            regex: &JAVA_BLOCKING_IO,
            vulnerability: VulnerabilityType::BlockingIO,
            base_score: 18.0,
            description_tmpl: "Blocking stream or thread sleep operation",
        },
        PatternRule {
            regex: &JAVA_BRANCH,
            vulnerability: VulnerabilityType::CognitiveBranch,
            base_score: 5.0,
            description_tmpl: "Branch / exception block increases cognitive complexity",
        },
        PatternRule {
            regex: &JAVA_MUTEX,
            vulnerability: VulnerabilityType::SyncContention,
            base_score: 22.0,
            description_tmpl: "Synchronized block or lock primitive contention",
        },
    ]
}

fn shell_rules() -> Vec<PatternRule> {
    vec![
        PatternRule {
            regex: &SH_LOOP,
            vulnerability: VulnerabilityType::DeepNesting,
            base_score: 15.0,
            description_tmpl: "Shell loop detected",
        },
        PatternRule {
            regex: &SH_BLOCKING_IO,
            vulnerability: VulnerabilityType::BlockingIO,
            base_score: 20.0,
            description_tmpl: "Subprocess network or sleep call in shell script",
        },
        PatternRule {
            regex: &SH_BRANCH,
            vulnerability: VulnerabilityType::CognitiveBranch,
            base_score: 5.0,
            description_tmpl: "Shell condition / branch logic",
        },
    ]
}

fn config_rules() -> Vec<PatternRule> {
    vec![
        PatternRule {
            regex: &CONFIG_SECRET,
            vulnerability: VulnerabilityType::SyncContention,
            base_score: 25.0,
            description_tmpl: "Sensitive credential or unrotated secret in config file",
        },
        PatternRule {
            regex: &CONFIG_BLOCKING,
            vulnerability: VulnerabilityType::BlockingIO,
            base_score: 15.0,
            description_tmpl: "Unbounded timeout or excessive connection limit in config",
        },
        PatternRule {
            regex: &CONFIG_BRANCH,
            vulnerability: VulnerabilityType::CognitiveBranch,
            base_score: 5.0,
            description_tmpl: "Complex conditional rule or assertion in config file",
        },
    ]
}

fn generic_rules() -> Vec<PatternRule> {
    vec![
        PatternRule {
            regex: &GENERIC_LOOP,
            vulnerability: VulnerabilityType::DeepNesting,
            base_score: 15.0,
            description_tmpl: "Loop construct detected in text file",
        },
        PatternRule {
            regex: &GENERIC_RECURSIVE,
            vulnerability: VulnerabilityType::RecursiveCall,
            base_score: 15.0,
            description_tmpl: "Identifier invocation checked for recursive call",
        },
        PatternRule {
            regex: &GENERIC_ALLOC,
            vulnerability: VulnerabilityType::HotAllocation,
            base_score: 10.0,
            description_tmpl: "Memory allocation operation",
        },
        PatternRule {
            regex: &GENERIC_BLOCKING,
            vulnerability: VulnerabilityType::BlockingIO,
            base_score: 15.0,
            description_tmpl: "Blocking sleep / wait / network primitive",
        },
        PatternRule {
            regex: &GENERIC_BRANCH,
            vulnerability: VulnerabilityType::CognitiveBranch,
            base_score: 5.0,
            description_tmpl: "Branch / condition construct",
        },
        PatternRule {
            regex: &GENERIC_MUTEX,
            vulnerability: VulnerabilityType::SyncContention,
            base_score: 18.0,
            description_tmpl: "Synchronization / concurrency primitive",
        },
    ]
}

// =============================================================================
// § 5  File language detection & Security Filters
// =============================================================================

/// Maximum file size scanned by the engine (512 KB) to prevent DoS / memory exhaustion.
pub const MAX_FILE_SIZE_BYTES: u64 = 512 * 1024;

/// Check if a file extension represents a compiled binary, asset, or archive.
pub fn is_excluded_extension(ext: &str) -> bool {
    matches!(
        ext.to_ascii_lowercase().as_str(),
        // Compiled & byte-code
        "exe" | "dll" | "so" | "dylib" | "bin" | "o" | "a" | "lib" | "class" | "jar" | "war"
        | "pyc" | "pyo" | "pyd" | "wasm"
        // Images & graphics
        | "png" | "jpg" | "jpeg" | "gif" | "bmp" | "ico" | "webp" | "tiff" | "psd" | "raw" | "svg"
        // Audio & video
        | "mp3" | "mp4" | "wav" | "ogg" | "flac" | "mkv" | "avi" | "mov" | "webm"
        // Archives & compression
        | "zip" | "tar" | "gz" | "bz2" | "xz" | "7z" | "rar" | "iso" | "dmg" | "pkg"
        // Documents & presentations
        | "pdf" | "doc" | "docx" | "ppt" | "pptx" | "xls" | "xlsx"
        // Fonts
        | "woff" | "woff2" | "ttf" | "eot" | "otf"
        // Databases & data dumps
        | "db" | "sqlite" | "sqlite3" | "parquet" | "arrow" | "avro"
        // Generated lockfiles & minified assets
        | "lock" | "sum" | "map"
    )
}

/// Security integrity check: inspect the first 512 bytes for null byte (0x00)
/// or read errors to guarantee non-binary text processing.
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

/// Returns the language string and rule table for a given file.
/// Safely skips excluded extensions, minified assets, lockfiles, and binary files.
pub fn detect_language(path: &Path) -> Option<(&'static str, Vec<PatternRule>)> {
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
        "py" | "pyw" => Some(("Python", python_rules())),
        "go" => Some(("Go", go_rules())),
        "js" | "jsx" | "mjs" | "cjs" => Some(("JavaScript", js_ts_rules())),
        "ts" | "tsx" => Some(("TypeScript", js_ts_rules())),
        "rs" => Some(("Rust", rust_rules())),
        "c" | "h" => Some(("C", c_cpp_rules())),
        "cpp" | "cc" | "cxx" | "hpp" | "hxx" => Some(("C++", c_cpp_rules())),
        "java" => Some(("Java", java_csharp_rules())),
        "cs" => Some(("C#", java_csharp_rules())),
        "kt" | "kts" => Some(("Kotlin", java_csharp_rules())),
        "scala" => Some(("Scala", java_csharp_rules())),
        "rb" => Some(("Ruby", generic_rules())),
        "php" => Some(("PHP", generic_rules())),
        "sh" | "bash" | "zsh" => Some(("Shell", shell_rules())),
        "json" | "yaml" | "yml" | "toml" | "xml" | "sql" => Some(("Config", config_rules())),
        _ => {
            if file_name.eq_ignore_ascii_case("dockerfile")
                || file_name.starts_with("Dockerfile.")
            {
                Some(("Config", config_rules()))
            } else if !file_name.starts_with('.') || !ext.is_empty() {
                Some(("Generic", generic_rules()))
            } else {
                None
            }
        }
    }
}

// Returns the function-name regex for a language.
fn func_regex_for(language: &str) -> &'static Lazy<Regex> {
    match language {
        "Python" => &PY_FUNC,
        "Go" => &GO_FUNC,
        "JavaScript" | "TypeScript" => &JS_FUNC,
        "Rust" => &RUST_FUNC,
        "C" | "C++" => &C_FUNC,
        "Java" | "C#" | "Kotlin" | "Scala" => &JAVA_FUNC,
        "Shell" => &SH_FUNC,
        _ => &GENERIC_FUNC,
    }
}

// =============================================================================
// § 6  Line-level analyzer
// =============================================================================

/// State threaded through the line-by-line scan.
struct ScanState {
    current_function: String,
    nesting_depth: usize, // tracks block nesting depth
    loop_depth: usize,    // specifically loop nesting (for / while)
    loop_levels: Vec<usize>, // tracks block depth or indent level of active loops
}

impl ScanState {
    fn new() -> Self {
        Self {
            current_function: "<module>".to_owned(),
            nesting_depth: 0,
            loop_depth: 0,
            loop_levels: Vec::new(),
        }
    }
}

/// Analyze a single trimmed source line.
///
/// Returns zero or more `Hotspot` values discovered on that line.
///
/// The nesting multiplier exponentially increases the entropy contribution
/// of any pattern found inside deeply-nested loops — this models the
/// actual O(n^k) impact on runtime complexity.
fn analyze_line(
    raw_line: &str,
    line_no: usize,
    state: &mut ScanState,
    rules: &[PatternRule],
    func_re: &Regex,
    language: &str,
) -> Vec<Hotspot> {
    let mut hotspots = Vec::new();
    let is_func_decl = func_re.is_match(raw_line);

    // ── Track current function context ────────────────────────────────────────
    if let Some(cap) = func_re.captures(raw_line) {
        let func_name = cap.get(1).or_else(|| cap.get(2)).map(|m| m.as_str().to_owned());
        if let Some(name) = func_name {
            state.current_function = name;
            // Reset per-function loop depth when entering a new function
            state.loop_depth = 0;
            state.loop_levels.clear();
        }
    }

    // ── Track nesting depth ───────────────────────────────────────────────────
    match language {
        "Python" => {
            let indent = raw_line.len() - raw_line.trim_start().len();
            state.nesting_depth = indent / 4;

            let trimmed = raw_line.trim();
            // A loop ends when an indentation level drops to or below the loop's indent level
            if !trimmed.is_empty() && !trimmed.starts_with('#') {
                while let Some(&loop_indent) = state.loop_levels.last() {
                    if indent <= loop_indent && !PY_FOR_WHILE.is_match(raw_line) {
                        state.loop_levels.pop();
                        state.loop_depth = state.loop_depth.saturating_sub(1);
                    } else {
                        break;
                    }
                }
            }

            if PY_FOR_WHILE.is_match(raw_line) {
                state.loop_depth = state.loop_depth.saturating_add(1);
                state.loop_levels.push(indent);
            }
        }
        "Shell" => {
            let trimmed = raw_line.trim();
            if SH_LOOP.is_match(raw_line) {
                state.loop_depth = state.loop_depth.saturating_add(1);
                state.loop_levels.push(state.nesting_depth);
            }
            if trimmed == "done" || trimmed.starts_with("done ") || trimmed.ends_with("; done") {
                state.loop_levels.pop();
                state.loop_depth = state.loop_depth.saturating_sub(1);
            }
        }
        _ => {
            // Brace-based languages: Go, Rust, JS/TS, C/C++, Java/C#, Generic
            let opens: usize = raw_line.chars().filter(|&c| c == '{').count();
            let closes: usize = raw_line.chars().filter(|&c| c == '}').count();

            // When closing braces appear, close any loops that were nested at or above this block level
            if closes > 0 {
                let depth_after_closes = state.nesting_depth.saturating_sub(closes);
                while let Some(&loop_level) = state.loop_levels.last() {
                    if depth_after_closes < loop_level {
                        state.loop_levels.pop();
                        state.loop_depth = state.loop_depth.saturating_sub(1);
                    } else {
                        break;
                    }
                }
            }

            state.nesting_depth = state.nesting_depth.saturating_sub(closes).saturating_add(opens);

            let is_loop = match language {
                "Go" => GO_FOR.is_match(raw_line),
                "JavaScript" | "TypeScript" => JS_FOR_WHILE.is_match(raw_line),
                "Rust" => RUST_FOR_WHILE.is_match(raw_line),
                "C" | "C++" => C_FOR_WHILE.is_match(raw_line),
                "Java" | "C#" | "Kotlin" | "Scala" => JAVA_FOR_WHILE.is_match(raw_line),
                _ => GENERIC_LOOP.is_match(raw_line),
            };

            if is_loop {
                state.loop_depth = state.loop_depth.saturating_add(1);
                state.loop_levels.push(state.nesting_depth);
            }
        }
    }

    // ── Loop-nesting entropy multiplier ──────────────────────────────────────
    // Score × (1 + 0.5 × loop_depth²) — exponential cost model:
    //   depth 0 → ×1.0, depth 1 → ×1.5, depth 2 → ×3.0, depth 3 → ×5.5
    let nesting_multiplier = 1.0 + 0.5 * (state.loop_depth as f64).powi(2);

    // ── Apply each rule ───────────────────────────────────────────────────────
    for rule in rules {
        // Skip flagging recursive call on the function declaration itself
        if is_func_decl && rule.vulnerability == VulnerabilityType::RecursiveCall {
            continue;
        }

        // For recursive calls, check that the call target actually matches the current function
        if rule.vulnerability == VulnerabilityType::RecursiveCall {
            if state.current_function == "<module>" {
                continue;
            }
            let self_call_pat = format!("{}(", state.current_function);
            let py_self_call = format!("self.{}(", state.current_function);
            let js_self_call = format!("this.{}(", state.current_function);
            let has_recursion = raw_line.contains(&self_call_pat)
                || raw_line.contains(&py_self_call)
                || raw_line.contains(&js_self_call);
            if !has_recursion {
                continue;
            }
        }
        if rule.regex.is_match(raw_line) {
            let raw_score = rule.base_score * nesting_multiplier;
            // Clamp to [0, 100]
            let entropy = raw_score.min(100.0).max(0.0);

            hotspots.push(Hotspot {
                function_name: state.current_function.clone(),
                line_number: line_no,
                source_snippet: raw_line.trim().chars().take(120).collect(),
                entropy_score: (entropy * 100.0).round() / 100.0, // 2 d.p.
                vulnerability_type: rule.vulnerability.clone(),
                description: rule.description_tmpl.to_owned(),
            });
        }
    }

    hotspots
}

// =============================================================================
// § 7  File-level scanner
// =============================================================================

/// Read and analyze a single source file.
///
/// Guards against non-text / binary files and enforces a 512 KB size limit
/// to maintain memory and security integrity.
pub fn scan_file(path: &Path, min_score: f64) -> io::Result<Option<FileReport>> {
    // ── Security Check: File Size Limit (512 KB) ──────────────────────────────
    if let Ok(meta) = fs::metadata(path) {
        if meta.len() > MAX_FILE_SIZE_BYTES {
            return Ok(None);
        }
    }

    // ── Security Check: Binary File Safety ────────────────────────────────────
    if is_binary_file(path) {
        return Ok(None);
    }

    // ── Language detection ────────────────────────────────────────────────────
    let (language, rules) = match detect_language(path) {
        Some(lr) => lr,
        None => return Ok(None), // unsupported file type
    };

    let func_re = func_regex_for(language);
    let file = fs::File::open(path)?;
    let reader = io::BufReader::new(file);

    let mut state = ScanState::new();
    let mut all_hotspots: Vec<Hotspot> = Vec::new();
    let mut lines_scanned: usize = 0;

    for (idx, line_result) in reader.lines().enumerate() {
        let line = match line_result {
            Ok(l) => l,
            Err(_) => return Ok(None), // Non-UTF-8 character stream fallback
        };
        let trimmed = line.trim();

        // Skip blank lines and pure comment lines to keep the scancount meaningful
        if trimmed.is_empty()
            || trimmed.starts_with('#')   // Python / shell comment
            || trimmed.starts_with("//")  // Go / C-style comment
            || trimmed.starts_with("/*")
        // block comment
        {
            continue;
        }

        lines_scanned += 1;

        let mut hotspots = analyze_line(
            &line,
            idx + 1, // 1-indexed line number
            &mut state,
            &rules,
            func_re,
            language,
        );

        // Apply min-score filter
        hotspots.retain(|h| h.entropy_score >= min_score);
        all_hotspots.extend(hotspots);
    }

    // ── Sort hotspots by entropy descending ───────────────────────────────────
    all_hotspots.sort_by(|a, b| b.entropy_score.partial_cmp(&a.entropy_score).unwrap());

    // ── Aggregate metrics ─────────────────────────────────────────────────────
    let total_entropy = all_hotspots.iter().map(|h| h.entropy_score).sum::<f64>();
    let mean_entropy = if all_hotspots.is_empty() {
        0.0
    } else {
        (total_entropy / all_hotspots.len() as f64 * 100.0).round() / 100.0
    };

    Ok(Some(FileReport {
        file_path: path.to_string_lossy().into_owned(),
        language: language.to_owned(),
        lines_scanned,
        total_entropy: (total_entropy * 100.0).round() / 100.0,
        mean_hotspot_entropy: mean_entropy,
        hotspots: all_hotspots,
    }))
}

// =============================================================================
// § 8  Directory walker
// =============================================================================

/// Walk `root` recursively, scan every supported source file, and collect
/// `FileReport` values. Uses Rayon for parallel execution when the feature
/// is enabled (default).
pub fn scan_directory(root: &Path, min_score: f64, verbose: bool) -> Vec<FileReport> {
    // Collect candidate paths first so we can parallelize the heavy I/O phase.
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

    // ── Parallel scan (Rayon) ─────────────────────────────────────────────────
    #[cfg(feature = "parallel")]
    let reports: Vec<FileReport> = {
        use rayon::iter::IntoParallelIterator;
        candidates
            .into_par_iter()
            .filter_map(|path| {
                if verbose {
                    println!("  {} {}", "⚡".yellow(), path.display());
                }
                match scan_file(&path, min_score) {
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

    // ── Sequential fallback (no Rayon feature) ────────────────────────────────
    #[cfg(not(feature = "parallel"))]
    let reports: Vec<FileReport> = candidates
        .into_iter()
        .filter_map(|path| {
            if verbose {
                println!("  {} {}", "→", path.display());
            }
            match scan_file(&path, min_score) {
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
// § 9  Report builder & JSON serializer
// =============================================================================

/// Assemble the root `ThermodynamicReport`, sort by entropy, and write JSON.
pub fn build_and_write_report(
    mut file_reports: Vec<FileReport>,
    scanned_dir: &Path,
    output_path: &Path,
) -> io::Result<ThermodynamicReport> {
    // Sort files by total_entropy descending — highest-entropy files first
    file_reports.sort_by(|a, b| b.total_entropy.partial_cmp(&a.total_entropy).unwrap());

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

    // Pretty-print JSON with 2-space indentation for human readability
    let json = serde_json::to_string_pretty(&report)
        .map_err(|e| io::Error::new(io::ErrorKind::Other, e))?;

    fs::write(output_path, json)?;

    Ok(report)
}

// =============================================================================
// § 10  Terminal summary banner
// =============================================================================

fn print_summary(report: &ThermodynamicReport) {
    println!();
    println!("{}", "╔══════════════════════════════════════════╗".cyan());
    println!("{}", "║   Thermodynamic AST Engine — Summary     ║".cyan());
    println!("{}", "╚══════════════════════════════════════════╝".cyan());
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
    println!("{}", "  Top 5 Entropy Hotspots:".bold());
    println!(
        "  {:<6}  {:<30}  {:<8}  {}",
        "Score", "File", "Line", "Vulnerability"
    );
    println!("  {}", "─".repeat(72));

    // Flatten and collect top 5 across all files
    let mut all_hotspots: Vec<(f64, &str, usize, &VulnerabilityType)> = report
        .file_reports
        .iter()
        .flat_map(|r| {
            r.hotspots.iter().map(move |h| {
                (
                    h.entropy_score,
                    r.file_path.as_str(),
                    h.line_number,
                    &h.vulnerability_type,
                )
            })
        })
        .collect();

    all_hotspots.sort_by(|a, b| b.0.partial_cmp(&a.0).unwrap());

    for (score, file, line, vuln) in all_hotspots.iter().take(5) {
        let short_file = Path::new(file)
            .file_name()
            .map(|n| n.to_string_lossy())
            .unwrap_or_default();
        println!(
            "  {:<6.2}  {:<30}  {:<8}  {}",
            score,
            short_file,
            line,
            vuln.to_string().red()
        );
    }

    println!();
}

// =============================================================================
// § 11  Entry point
// =============================================================================

fn main() {
    let cli = Cli::parse();

    // ── Validate input directory ──────────────────────────────────────────────
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

    // ── Scan ──────────────────────────────────────────────────────────────────
    let file_reports = scan_directory(&cli.directory, cli.min_score, cli.verbose);

    if file_reports.is_empty() {
        println!(
            "{} No supported source files found in '{}'.",
            "WARN:".yellow().bold(),
            cli.directory.display()
        );
        let _ = build_and_write_report(Vec::new(), &cli.directory, &cli.output);
        std::process::exit(0);
    }

    // ── Build & persist report ────────────────────────────────────────────────
    let report = match build_and_write_report(file_reports, &cli.directory, &cli.output) {
        Ok(r) => r,
        Err(e) => {
            eprintln!("{} Failed to write report: {}", "ERROR:".red().bold(), e);
            std::process::exit(1);
        }
    };

    // ── Print summary banner ──────────────────────────────────────────────────
    print_summary(&report);

    println!(
        "{} Report written to: {}\n",
        "✓".green().bold(),
        cli.output.display().to_string().green()
    );
}

// =============================================================================
// § 12  Unit tests
// =============================================================================

#[cfg(test)]
mod tests {
    use super::*;

    // ── Hotspot detection tests ───────────────────────────────────────────────

    #[test]
    fn test_python_blocking_io_detected() {
        let line = "    result = requests.get(url, timeout=30)";
        let mut state = ScanState::new();
        let rules = python_rules();
        let hotspots = analyze_line(line, 42, &mut state, &rules, &PY_FUNC, "Python");

        assert!(
            hotspots
                .iter()
                .any(|h| h.vulnerability_type == VulnerabilityType::BlockingIO),
            "Expected BlockingIO hotspot for requests.get"
        );
    }

    #[test]
    fn test_go_allocation_detected() {
        let line = "    data := make([]byte, 1024)";
        let mut state = ScanState::new();
        let rules = go_rules();
        let hotspots = analyze_line(line, 7, &mut state, &rules, &GO_FUNC, "Go");

        assert!(
            hotspots
                .iter()
                .any(|h| h.vulnerability_type == VulnerabilityType::HotAllocation),
            "Expected HotAllocation hotspot for make()"
        );
    }

    #[test]
    fn test_nesting_multiplier_increases_score() {
        let mut state = ScanState::new();
        let rules = python_rules();

        // Simulate two levels of loop nesting
        state.loop_depth = 2;

        let line = "    result = requests.get(url)";
        let hotspots = analyze_line(line, 10, &mut state, &rules, &PY_FUNC, "Python");

        let blocking_io = hotspots
            .iter()
            .find(|h| h.vulnerability_type == VulnerabilityType::BlockingIO)
            .expect("Expected BlockingIO hotspot");

        // base_score=18, nesting multiplier at depth 2 = 1 + 0.5*4 = 3.0 → 54.0
        assert!(
            blocking_io.entropy_score > 18.0,
            "Nesting multiplier should increase entropy beyond base score"
        );
    }

    #[test]
    fn test_go_mutex_detected() {
        let line = "    mu.Lock()  // sync.Mutex";
        let mut state = ScanState::new();
        let rules = go_rules();
        let _hotspots = analyze_line(line, 99, &mut state, &rules, &GO_FUNC, "Go");

        // The mutex pattern matches on `sync.` prefix — verify SyncContention fires
        // on a line that explicitly contains the package
        let line2 = "    var mu sync.Mutex";
        let hotspots2 = analyze_line(line2, 100, &mut state, &rules, &GO_FUNC, "Go");

        assert!(
            hotspots2
                .iter()
                .any(|h| h.vulnerability_type == VulnerabilityType::SyncContention),
            "Expected SyncContention for sync.Mutex declaration"
        );
    }

    #[test]
    fn test_language_detection() {
        assert!(detect_language(Path::new("foo.py")).is_some());
        assert!(detect_language(Path::new("bar.go")).is_some());
        assert!(detect_language(Path::new("baz.rs")).is_some());
        assert!(detect_language(Path::new("qux.js")).is_some());
        assert!(detect_language(Path::new("App.tsx")).is_some());
        assert!(detect_language(Path::new("main.cpp")).is_some());
        assert!(detect_language(Path::new("Server.java")).is_some());
        assert!(detect_language(Path::new("deploy.sh")).is_some());
        assert!(detect_language(Path::new("config.yaml")).is_some());
        assert!(detect_language(Path::new("Dockerfile")).is_some());
        assert!(detect_language(Path::new("notes.txt")).is_some());

        // Binary and excluded files must return None
        assert!(detect_language(Path::new("image.png")).is_none());
        assert!(detect_language(Path::new("program.exe")).is_none());
        assert!(detect_language(Path::new("bundle.min.js")).is_none());
        assert!(detect_language(Path::new("package-lock.json")).is_none());
        assert!(detect_language(Path::new("Cargo.lock")).is_none());
        assert!(detect_language(Path::new("module.wasm")).is_none());
    }

    #[test]
    fn test_polyglot_hotspots() {
        // Test JS blocking I/O
        let mut js_state = ScanState::new();
        let js_rules = js_ts_rules();
        let js_hotspots = analyze_line(
            "const data = fs.readFileSync('foo.txt');",
            1,
            &mut js_state,
            &js_rules,
            &JS_FUNC,
            "JavaScript",
        );
        assert!(
            js_hotspots.iter().any(|h| h.vulnerability_type == VulnerabilityType::BlockingIO),
            "Expected BlockingIO for fs.readFileSync"
        );

        // Test Rust allocation and mutex
        let mut rs_state = ScanState::new();
        let rs_rules = rust_rules();
        let rs_hotspots = analyze_line(
            "let items = Vec::with_capacity(1000);",
            1,
            &mut rs_state,
            &rs_rules,
            &RUST_FUNC,
            "Rust",
        );
        assert!(
            rs_hotspots.iter().any(|h| h.vulnerability_type == VulnerabilityType::HotAllocation),
            "Expected HotAllocation for Vec::with_capacity"
        );
    }

    #[test]
    fn test_is_binary_file_detection() {
        let temp_dir = std::env::temp_dir();
        let text_file = temp_dir.join("test_plain_text.txt");
        let bin_file = temp_dir.join("test_binary_blob.bin");

        let _ = fs::write(&text_file, "Hello, this is a plain text file!\nNo null bytes here.");
        let _ = fs::write(&bin_file, b"Hello\x00Binary\x00Data");

        assert!(!is_binary_file(&text_file), "Text file should not be marked binary");
        assert!(is_binary_file(&bin_file), "File with null bytes must be marked binary");

        let _ = fs::remove_file(&text_file);
        let _ = fs::remove_file(&bin_file);
    }

    #[test]
    fn test_entropy_score_clamped() {
        let mut state = ScanState::new();
        // Set absurdly high nesting depth to trigger clamping
        state.loop_depth = 100;
        let rules = python_rules();
        let line = "    mu.acquire()  # threading.Lock()";
        let hotspots = analyze_line(line, 1, &mut state, &rules, &PY_FUNC, "Python");
        for h in &hotspots {
            assert!(h.entropy_score <= 100.0, "Entropy must be clamped to 100.0");
        }
    }

    #[test]
    fn test_sequential_loops_do_not_inflate_nesting_in_go() {
        let mut state = ScanState::new();
        let rules = go_rules();

        // Start func
        analyze_line("func ProcessData() {", 1, &mut state, &rules, &GO_FUNC, "Go");
        assert_eq!(state.loop_depth, 0);

        // First loop
        analyze_line("    for i := 0; i < 10; i++ {", 2, &mut state, &rules, &GO_FUNC, "Go");
        assert_eq!(state.loop_depth, 1);

        // Close first loop
        analyze_line("    }", 3, &mut state, &rules, &GO_FUNC, "Go");
        assert_eq!(state.loop_depth, 0, "Loop depth must decrement back to 0 when loop block closes");

        // Second sequential loop
        analyze_line("    for j := 0; j < 10; j++ {", 4, &mut state, &rules, &GO_FUNC, "Go");
        assert_eq!(state.loop_depth, 1, "Second sequential loop should have loop_depth 1, NOT 2");

        // Close second loop
        analyze_line("    }", 5, &mut state, &rules, &GO_FUNC, "Go");
        assert_eq!(state.loop_depth, 0);
    }

    #[test]
    fn test_nested_loops_scale_properly_in_go() {
        let mut state = ScanState::new();
        let rules = go_rules();

        analyze_line("func ProcessData() {", 1, &mut state, &rules, &GO_FUNC, "Go");
        analyze_line("    for i := 0; i < 10; i++ {", 2, &mut state, &rules, &GO_FUNC, "Go");
        assert_eq!(state.loop_depth, 1);

        analyze_line("        for j := 0; j < 10; j++ {", 3, &mut state, &rules, &GO_FUNC, "Go");
        assert_eq!(state.loop_depth, 2, "Nested loop should have loop_depth 2");

        analyze_line("        }", 4, &mut state, &rules, &GO_FUNC, "Go");
        assert_eq!(state.loop_depth, 1, "Exiting inner loop should drop depth to 1");

        analyze_line("    }", 5, &mut state, &rules, &GO_FUNC, "Go");
        assert_eq!(state.loop_depth, 0, "Exiting outer loop should drop depth to 0");
    }

    #[test]
    fn test_func_decl_not_flagged_as_recursive() {
        let mut state = ScanState::new();
        let rules = go_rules();

        // Function declaration should NOT trigger RecursiveCall
        let hotspots = analyze_line("func NewCrawler(timeout time.Duration) *Crawler {", 1, &mut state, &rules, &GO_FUNC, "Go");
        assert!(
            !hotspots.iter().any(|h| h.vulnerability_type == VulnerabilityType::RecursiveCall),
            "Function declaration must not be flagged as a recursive call"
        );

        // Self-recursion should trigger RecursiveCall
        let recursive_hotspots = analyze_line("    return NewCrawler(timeout)", 2, &mut state, &rules, &GO_FUNC, "Go");
        assert!(
            recursive_hotspots.iter().any(|h| h.vulnerability_type == VulnerabilityType::RecursiveCall),
            "Direct self-recursive call must be flagged as RecursiveCall"
        );

        // Calling another function should NOT trigger RecursiveCall
        let external_hotspots = analyze_line("    return OtherFunction()", 3, &mut state, &rules, &GO_FUNC, "Go");
        assert!(
            !external_hotspots.iter().any(|h| h.vulnerability_type == VulnerabilityType::RecursiveCall),
            "Calling a different function must not be flagged as RecursiveCall"
        );
    }
}

