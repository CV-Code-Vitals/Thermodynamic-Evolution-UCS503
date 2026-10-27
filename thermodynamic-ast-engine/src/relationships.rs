use crate::{FileReport, Language};
use serde::{Deserialize, Serialize};
use std::collections::{BTreeSet, HashMap};
use std::path::Path;
use tree_sitter::Node;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ImportReference {
    pub path: String,
    pub kind: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Relationship {
    pub kind: String,
    pub source_file: String,
    pub target_file: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub source_start_byte: Option<usize>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub target_start_byte: Option<usize>,
}

fn text<'a>(node: Node, source: &'a [u8]) -> &'a str {
    node.utf8_text(source).unwrap_or("")
}
fn literal(node: Node, source: &[u8]) -> Option<String> {
    let raw = text(node, source).trim();
    let quote = raw.chars().next()?;
    if !matches!(quote, '\'' | '"' | '`') || !raw.ends_with(quote) || raw.len() < 2 {
        return None;
    }
    let value = &raw[1..raw.len() - 1];
    if value.contains('$') || value.contains('\n') {
        return None;
    }
    Some(value.to_owned())
}

pub fn imports(root: Node, source: &[u8], lang: Language) -> Vec<ImportReference> {
    fn walk(node: Node, source: &[u8], lang: Language, found: &mut BTreeSet<(String, String)>) {
        let kind = node.kind();
        let mut add = |path: String, kind: &str| {
            if !path.is_empty() {
                found.insert((path, kind.to_owned()));
            }
        };
        match lang {
            Language::JavaScript | Language::TypeScript | Language::Tsx => {
                if matches!(kind, "import_statement" | "export_statement") {
                    if let Some(path) = node
                        .child_by_field_name("source")
                        .and_then(|n| literal(n, source))
                    {
                        add(path, "relative");
                    }
                } else if kind == "call_expression" {
                    if let Some(function) = node.child_by_field_name("function") {
                        if matches!(text(function, source), "require" | "import") {
                            if let Some(value) = node
                                .child_by_field_name("arguments")
                                .and_then(|n| n.named_child(0))
                                .and_then(|n| literal(n, source))
                            {
                                add(value, "relative");
                            }
                        }
                    }
                }
            }
            Language::Python => {
                if kind == "import_from_statement" {
                    if let Some(name) = node.child_by_field_name("module_name") {
                        add(text(name, source).to_owned(), "python");
                    }
                } else if kind == "import_statement" {
                    let mut cursor = node.walk();
                    for child in node.named_children(&mut cursor) {
                        let name = child.child_by_field_name("name").unwrap_or(child);
                        add(text(name, source).to_owned(), "python");
                    }
                }
            }
            Language::Go if kind == "import_spec" => {
                if let Some(path) = node
                    .child_by_field_name("path")
                    .and_then(|n| literal(n, source))
                {
                    add(path, "go");
                }
            }
            Language::C | Language::Cpp if kind == "preproc_include" => {
                if let Some(path) = node
                    .child_by_field_name("path")
                    .and_then(|n| literal(n, source))
                {
                    add(path, "relative");
                }
            }
            Language::Java if kind == "import_declaration" => {
                let name = text(node, source)
                    .trim_start_matches("import ")
                    .trim_start_matches("static ")
                    .trim_end_matches(';')
                    .trim();
                if !name.ends_with(".*") {
                    add(name.to_owned(), "java");
                }
            }
            Language::Swift if kind == "import_declaration" => {
                if let Some(name) = node.named_child(0) {
                    add(text(name, source).to_owned(), "swift");
                }
            }
            Language::Rust if kind == "mod_item" && node.child_by_field_name("body").is_none() => {
                if let Some(name) = node.child_by_field_name("name") {
                    add(text(name, source).to_owned(), "rust_mod");
                }
            }
            Language::Rust if kind == "use_declaration" => {
                let name = text(node, source)
                    .trim_start_matches("use ")
                    .trim_end_matches(';')
                    .trim();
                if !name.contains(['{', '*']) {
                    add(
                        name.split(" as ").next().unwrap_or(name).to_owned(),
                        "rust_use",
                    );
                }
            }
            Language::Ruby if kind == "call" => {
                let method = node
                    .child_by_field_name("method")
                    .map(|n| text(n, source))
                    .unwrap_or("");
                if matches!(method, "require_relative" | "require") {
                    if let Some(value) = node
                        .child_by_field_name("arguments")
                        .and_then(|n| n.named_child(0))
                        .and_then(|n| literal(n, source))
                    {
                        add(
                            value,
                            if method == "require_relative" {
                                "relative"
                            } else {
                                "ruby"
                            },
                        );
                    }
                }
            }
            Language::Php
                if matches!(
                    kind,
                    "include_expression"
                        | "require_expression"
                        | "include_once_expression"
                        | "require_once_expression"
                ) =>
            {
                if let Some(value) = node.named_child(0).and_then(|n| literal(n, source)) {
                    add(value, "relative");
                }
            }
            Language::Bash if kind == "command" => {
                let raw = text(node, source);
                let mut words = raw.split_whitespace();
                if matches!(words.next(), Some("source" | ".")) {
                    if let Some(path) = words.next() {
                        if !path.contains('$') {
                            add(path.trim_matches(['\'', '"']).to_owned(), "relative");
                        }
                    }
                }
            }
            _ => {}
        }
        let mut cursor = node.walk();
        for child in node.named_children(&mut cursor) {
            walk(child, source, lang, found);
        }
    }
    let mut found = BTreeSet::new();
    walk(root, source, lang, &mut found);
    found
        .into_iter()
        .map(|(path, kind)| ImportReference { path, kind })
        .collect()
}

pub fn calls(function: Node, source: &[u8], lang: Language) -> Vec<String> {
    fn walk(node: Node, source: &[u8], lang: Language, names: &mut BTreeSet<String>) {
        if crate::is_function_node(node.kind(), lang) {
            return;
        }
        if matches!(
            node.kind(),
            "call_expression"
                | "call"
                | "method_invocation"
                | "invocation_expression"
                | "function_call_expression"
                | "command"
        ) && node.child_by_field_name("object").is_none()
            && node.child_by_field_name("receiver").is_none()
        {
            if let Some(target) = node
                .child_by_field_name("function")
                .or_else(|| node.child_by_field_name("name"))
                .or_else(|| node.child_by_field_name("method"))
                .or_else(|| {
                    if lang == Language::Swift {
                        node.named_child(0)
                    } else {
                        None
                    }
                })
            {
                if matches!(
                    target.kind(),
                    "identifier" | "simple_identifier" | "name" | "command_name"
                ) {
                    names.insert(text(target, source).to_owned());
                }
            }
        }
        let mut cursor = node.walk();
        for child in node.named_children(&mut cursor) {
            walk(child, source, lang, names);
        }
    }
    let mut names = BTreeSet::new();
    let mut cursor = function.walk();
    for child in function.named_children(&mut cursor) {
        walk(child, source, lang, &mut names);
    }
    names.into_iter().collect()
}

fn normalize(path: &str) -> String {
    let mut parts = Vec::new();
    for part in path.split('/') {
        match part {
            "" | "." => {}
            ".." => {
                if parts.last().is_some_and(|part| *part != "..") {
                    parts.pop();
                } else {
                    parts.push("..");
                }
            }
            _ => parts.push(part),
        }
    }
    parts.join("/")
}

pub fn build(files: &[FileReport], root: &Path) -> Vec<Relationship> {
    let paths: Vec<String> = files
        .iter()
        .map(|file| {
            let path = Path::new(&file.file_path);
            path.strip_prefix(root)
                .unwrap_or(path)
                .to_string_lossy()
                .replace('\\', "/")
        })
        .collect();
    let lookup: HashMap<&str, usize> = paths
        .iter()
        .enumerate()
        .map(|(i, p)| (p.as_str(), i))
        .collect();
    let mut suffixes: HashMap<String, Vec<usize>> = HashMap::new();
    let mut go_packages: HashMap<String, Vec<usize>> = HashMap::new();
    let mut go_modules = Vec::new();
    for (i, path) in paths.iter().enumerate() {
        let segments: Vec<&str> = path.split('/').collect();
        for start in 0..segments.len() {
            suffixes
                .entry(segments[start..].join("/"))
                .or_default()
                .push(i);
        }
        let parent = path.rsplit_once('/').map(|(p, _)| p).unwrap_or("");
        if files[i].language == "Go" && !path.ends_with("_test.go") {
            go_packages.entry(parent.to_owned()).or_default().push(i);
        }
        if path.ends_with("go.mod") {
            if let Ok(contents) = std::fs::read_to_string(&files[i].file_path) {
                if let Some(module) = contents
                    .lines()
                    .find_map(|line| line.trim().strip_prefix("module "))
                {
                    go_modules.push((
                        module.trim().trim_matches('"').to_owned(),
                        parent.to_owned(),
                    ));
                }
            }
        }
    }
    go_modules.sort_by_key(|(module, _)| std::cmp::Reverse(module.len()));
    let find = |candidate: &str, suffix: bool| -> Option<usize> {
        let key = normalize(candidate);
        let extensions = [
            "",
            ".ts",
            ".tsx",
            ".js",
            ".jsx",
            ".mjs",
            ".cjs",
            ".py",
            ".rs",
            ".rb",
            ".php",
            ".sh",
            ".h",
            ".hpp",
            "/index.ts",
            "/index.tsx",
            "/index.js",
            "/index.jsx",
            "/__init__.py",
            "/mod.rs",
        ];
        for extension in extensions {
            let full = format!("{}{}", key, extension);
            if suffix {
                if let Some(indices) = suffixes.get(&full) {
                    if indices.len() == 1 {
                        return Some(indices[0]);
                    }
                }
            } else if let Some(index) = lookup.get(full.as_str()) {
                return Some(*index);
            }
        }
        None
    };
    let mut output = Vec::new();
    let mut edges = BTreeSet::new();
    for (i, file) in files.iter().enumerate() {
        let source = &paths[i];
        let parent = source.rsplit_once('/').map(|(p, _)| p).unwrap_or("");
        for import in &file.imports {
            let mut targets = Vec::new();
            match import.kind.as_str() {
                "relative" => {
                    if import.path.starts_with('/') {
                        continue;
                    }
                    // Bare JavaScript specifiers refer to packages, not similarly named repository files.
                    if !file.language.contains("Script") && file.language != "TSX"
                        || import.path.starts_with('.')
                    {
                        if let Some(target) = find(&format!("{}/{}", parent, import.path), false) {
                            targets.push(target);
                        }
                    }
                }
                "python" => {
                    let dots = import.path.chars().take_while(|c| *c == '.').count();
                    let module = import.path[dots..].replace('.', "/");
                    let base = if dots > 0 {
                        format!("{}/{}{}", parent, "../".repeat(dots - 1), module)
                    } else {
                        module
                    };
                    if let Some(target) = find(&base, dots == 0) {
                        targets.push(target);
                    }
                }
                "java" => {
                    let mut module = import.path.replace('.', "/");
                    loop {
                        if let Some(target) = find(&format!("{}.java", module), true) {
                            targets.push(target);
                            break;
                        }
                        match module.rsplit_once('/') {
                            Some((base, _)) => module = base.to_owned(),
                            None => break,
                        }
                    }
                }
                "go" => {
                    for (name, base) in &go_modules {
                        let relative = if import.path == *name {
                            Some("")
                        } else {
                            import.path.strip_prefix(&format!("{}/", name))
                        };
                        if let Some(relative) = relative {
                            if let Some(package) =
                                go_packages.get(&normalize(&format!("{}/{}", base, relative)))
                            {
                                targets.extend(package);
                            }
                            break;
                        }
                    }
                }
                "ruby" => {
                    if let Some(target) = find(&import.path, true) {
                        targets.push(target);
                    }
                }
                "rust_mod" | "rust_use" => {
                    if import.kind == "rust_use"
                        && !["crate::", "self::", "super::"]
                            .iter()
                            .any(|prefix| import.path.starts_with(prefix))
                    {
                        continue;
                    }
                    let mut module = import.path.replace("::", "/");
                    let mut base = parent.to_owned();
                    if module.starts_with("crate/") {
                        module = module.trim_start_matches("crate/").to_owned();
                        loop {
                            if lookup.contains_key(format!("{}/lib.rs", base).as_str())
                                || lookup.contains_key(format!("{}/main.rs", base).as_str())
                            {
                                break;
                            }
                            match base.rsplit_once('/') {
                                Some((p, _)) => base = p.to_owned(),
                                None => {
                                    base.clear();
                                    break;
                                }
                            }
                        }
                    } else if module.starts_with("super/") {
                        module = module.replacen("super/", "../", 1);
                    } else {
                        module = module.trim_start_matches("self/").to_owned();
                    }
                    if import.kind == "rust_mod"
                        && !source.ends_with("/mod.rs")
                        && !source.ends_with("/lib.rs")
                        && !source.ends_with("/main.rs")
                        && source.contains('/')
                    {
                        base = source.trim_end_matches(".rs").to_owned();
                    }
                    loop {
                        if let Some(target) = find(&format!("{}/{}", base, module), false) {
                            targets.push(target);
                            break;
                        }
                        if import.kind == "rust_mod" {
                            break;
                        }
                        match module.rsplit_once('/') {
                            Some((p, _)) => module = p.to_owned(),
                            None => break,
                        }
                    }
                }
                _ => {}
            }
            for target in targets {
                if target != i && edges.insert((i, target)) {
                    output.push(Relationship {
                        kind: "imports".to_owned(),
                        source_file: source.clone(),
                        target_file: paths[target].clone(),
                        source_start_byte: None,
                        target_start_byte: None,
                    });
                }
            }
        }
        let mut names: HashMap<&str, Vec<usize>> = HashMap::new();
        for (index, hotspot) in file.hotspots.iter().enumerate() {
            names.entry(&hotspot.function_name).or_default().push(index);
        }
        for hotspot in &file.hotspots {
            for called in &hotspot.calls {
                if let Some(targets) = names.get(called.as_str()) {
                    if targets.len() == 1 {
                        output.push(Relationship {
                            kind: "calls".to_owned(),
                            source_file: source.clone(),
                            target_file: source.clone(),
                            source_start_byte: Some(hotspot.byte_range[0]),
                            target_start_byte: Some(file.hotspots[targets[0]].byte_range[0]),
                        });
                    }
                }
            }
        }
    }
    output
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn resolves_local_imports_and_unambiguous_calls_without_inventing_package_links() {
        let root =
            std::env::temp_dir().join(format!("ucs503-relationships-{}", std::process::id()));
        let sources = [
            ("ui/card.tsx", "import {helper} from './math'; import lodash from 'lodash'; function render() { return local(); } function local() { return 1; }"),
            ("ui/math.ts", "export function helper() { return 1; }"),
            ("ui/lodash.ts", "export function unused() {}"),
            ("app/main.py", "from .helper import pick\ndef render():\n    return local()\ndef local():\n    return 1\n"),
            ("app/helper.py", "def pick():\n    return 1\n"),
            ("goproj/go.mod", "module example.local/project\n\ngo 1.25.0\n"),
            ("goproj/main.go", "package main\nimport \"example.local/project/lib\"\nfunc main() { local() }; func local() {}"),
            ("goproj/lib/helper.go", "package lib\nfunc Helper() {}"),
            ("c/main.c", "#include \"include/helper.h\"\nint main() { return 0; }"),
            ("c/include/helper.h", "int helper(void);"),
            ("src/lib.rs", "mod helper; pub fn render() { local(); } fn local() {}"),
            ("src/helper.rs", "pub fn helper() {}"),
            ("java/Main.java", "import demo.Helper; class Main { int local() { return 1; } int render() { return local(); } }"),
            ("java/demo/Helper.java", "package demo; class Helper {}"),
            ("ambiguous.ts", "class A { run() {} } class B { run() {} } function caller() { run(); }"),
            ("swift/Main.swift", "func render() -> Int { return local() }; func local() -> Int { return 1 }"),
        ];
        for (path, contents) in sources {
            let path = root.join(path);
            std::fs::create_dir_all(path.parent().unwrap()).unwrap();
            std::fs::write(path, contents).unwrap();
        }
        let reports = crate::scan_directory(&root, 0.0, 10.0, false);
        let edges = build(&reports, &root);
        let imports: BTreeSet<_> = edges
            .iter()
            .filter(|edge| edge.kind == "imports")
            .map(|edge| (edge.source_file.as_str(), edge.target_file.as_str()))
            .collect();
        assert_eq!(
            imports,
            BTreeSet::from([
                ("ui/card.tsx", "ui/math.ts"),
                ("app/main.py", "app/helper.py"),
                ("goproj/main.go", "goproj/lib/helper.go"),
                ("c/main.c", "c/include/helper.h"),
                ("src/lib.rs", "src/helper.rs"),
                ("java/Main.java", "java/demo/Helper.java"),
            ])
        );
        for path in [
            "ui/card.tsx",
            "app/main.py",
            "goproj/main.go",
            "src/lib.rs",
            "java/Main.java",
            "swift/Main.swift",
        ] {
            assert!(
                edges.iter().any(|edge| edge.kind == "calls"
                    && edge.source_file == path
                    && edge.target_file == path),
                "missing call in {}",
                path
            );
        }
        assert!(!edges
            .iter()
            .any(|edge| edge.kind == "calls" && edge.source_file == "ambiguous.ts"));
        std::fs::remove_dir_all(root).unwrap();
    }
}
