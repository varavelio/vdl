from pygments.lexer import RegexLexer, bygroups
from pygments.token import *

class VdlLexer(RegexLexer):
    name = "VDL"
    aliases = ["vdl"]
    filenames = ["*.vdl"]

    tokens = {
        "root": [
            # Line comments
            (r"//.*$", Comment.Single),
            # Block comments
            (r"/\*", Comment.Multiline, "block-comment"),
            # Docstrings (before strings)
            (r'"""', String.Doc, "docstring"),
            # Strings
            (r'"', String.Double, "string"),
            # Annotations
            (r"@[a-zA-Z_][a-zA-Z0-9_]*", Name.Decorator),
            # Type/enum declarations: type Foo, enum Bar
            (
                r"\b(type|enum)\s+([a-zA-Z_][a-zA-Z0-9_]*)",
                bygroups(Keyword.Declaration, Name.Class),
            ),
            # Const declarations: const NAME
            (
                r"\b(const)\s+([a-zA-Z_][a-zA-Z0-9_]*)",
                bygroups(Keyword.Declaration, Name.Constant),
            ),
            # Keywords
            (r"\binclude\b", Keyword.Declaration),
            # Boolean constants
            (r"\btrue\b|\bfalse\b", Keyword.Constant),
            # Numeric constants
            (r"\b\d+\.\d+\b", Number.Float),
            (r"\b\d+\b", Number.Integer),
            # Spreads: ...TypeName
            (
                r"(\.\.\.)([a-zA-Z_][a-zA-Z0-9_]*)",
                bygroups(Operator, Name.Class),
            ),
            # Primitive types
            (r"\b(string|int|float|bool|datetime)\b", Keyword.Type),
            # Map type
            (r"\bmap\b", Keyword.Type),
            # Enum member reference: EnumName.MEMBER
            (
                r"\b([A-Z][a-zA-Z0-9]*)\.([A-Z][a-zA-Z0-9_]*)",
                bygroups(Name.Class, Name.Constant),
            ),
            # PascalCase identifiers (type/enum references)
            (r"\b[A-Z][a-zA-Z0-9]*\b", Name.Class),
            # Line breaks (before field pattern to reset line start position)
            (r"\n", Text),
            # Field names: after indent at line start
            (
                r'^[ \t]+([a-z_][a-zA-Z0-9_]*)(\?)?(?=\s+[a-zA-Z\[{"]|\s*$)',
                bygroups(Name.Variable, Operator),
            ),
            # Inline whitespace (non-newline)
            (r"[ \t]+", Text),
            # References: lowercase identifiers used as values
            (
                r"(?:^|(?<=[=\s\[]))([a-z_][a-zA-Z0-9_]*)\b",
                Name.Class,
            ),
            # Operators and punctuation
            (r"[?=]", Operator),
            (r"[{}()\[\]]", Punctuation),
            # Fallback whitespace
            (r"\s+", Text),
            # Generic fallback
            (r".", Text),
        ],
        "block-comment": [
            (r"\*/", Comment.Multiline, "#pop"),
            (r"[^*]+", Comment.Multiline),
            (r"\*", Comment.Multiline),
        ],
        "docstring": [
            (r'"""', String.Doc, "#pop"),
            (r"\./[a-zA-Z0-9_\-./]+\.md", Comment.Special),
            (r'[^"]+', String.Doc),
            (r'"', String.Doc),
        ],
        "string": [
            (r'\\[\\abfnrtv"]', String.Escape),
            (r'"', String.Double, "#pop"),
            (r'[^"\\]+', String.Double),
        ],
    }
