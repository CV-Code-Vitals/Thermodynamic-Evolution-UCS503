use std::path::Path;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Language {
    Go,
    Python,
    JavaScript,
    TypeScript,
    Tsx,
    Rust,
    Java,
    C,
    Cpp,
    CSharp,
    Ruby,
    Php,
    Bash,
    Swift,
    Text,
}

impl Language {
    pub fn grammar(self) -> Option<tree_sitter::Language> {
        Some(match self {
            Self::Go => tree_sitter_go::LANGUAGE.into(),
            Self::Python => tree_sitter_python::LANGUAGE.into(),
            Self::JavaScript => tree_sitter_javascript::LANGUAGE.into(),
            Self::TypeScript => tree_sitter_typescript::LANGUAGE_TYPESCRIPT.into(),
            Self::Tsx => tree_sitter_typescript::LANGUAGE_TSX.into(),
            Self::Rust => tree_sitter_rust::LANGUAGE.into(),
            Self::Java => tree_sitter_java::LANGUAGE.into(),
            Self::C => tree_sitter_c::LANGUAGE.into(),
            Self::Cpp => tree_sitter_cpp::LANGUAGE.into(),
            Self::CSharp => tree_sitter_c_sharp::LANGUAGE.into(),
            Self::Ruby => tree_sitter_ruby::LANGUAGE.into(),
            Self::Php => tree_sitter_php::LANGUAGE_PHP.into(),
            Self::Bash => tree_sitter_bash::LANGUAGE.into(),
            Self::Swift => tree_sitter_swift::LANGUAGE.into(),
            Self::Text => return None,
        })
    }

    pub fn name(self, path: &Path) -> String {
        let name = match self {
            Self::Go => "Go",
            Self::Python => "Python",
            Self::JavaScript => "JavaScript / JSX",
            Self::TypeScript => "TypeScript",
            Self::Tsx => "TSX",
            Self::Rust => "Rust",
            Self::Java => "Java",
            Self::C => "C",
            Self::Cpp => "C++",
            Self::CSharp => "C#",
            Self::Ruby => "Ruby",
            Self::Php => "PHP",
            Self::Bash => "Shell",
            Self::Swift => "Swift",
            Self::Text => match path
                .extension()
                .and_then(|s| s.to_str())
                .unwrap_or("")
                .to_ascii_lowercase()
                .as_str()
            {
                "html" | "htm" => "HTML",
                "css" | "scss" | "sass" | "less" => "Stylesheet",
                "json" | "jsonc" => "JSON",
                "yaml" | "yml" => "YAML",
                "toml" => "TOML",
                "md" | "mdx" | "rst" => "Documentation",
                "kt" | "kts" => "Kotlin",
                "swift" => "Swift",
                "dart" => "Dart",
                "lua" => "Lua",
                "r" | "rmd" => "R",
                "scala" | "sc" => "Scala",
                "sql" => "SQL",
                "vue" => "Vue",
                "svelte" => "Svelte",
                "ex" | "exs" => "Elixir",
                "erl" | "hrl" => "Erlang",
                "hs" => "Haskell",
                "pl" | "pm" => "Perl",
                "jl" => "Julia",
                "ps1" | "psm1" => "PowerShell",
                "m" | "mm" => "Objective-C / MATLAB",
                "fs" | "fsx" => "F#",
                "clj" | "cljs" | "cljc" => "Clojure",
                "tf" | "tfvars" => "Terraform",
                "proto" => "Protocol Buffers",
                "xml" => "XML",
                "ini" | "cfg" | "conf" => "Configuration",
                "" => {
                    return path
                        .file_name()
                        .unwrap_or_default()
                        .to_string_lossy()
                        .into_owned()
                }
                extension => return format!("{} (text)", extension),
            },
        };
        name.to_owned()
    }
}
