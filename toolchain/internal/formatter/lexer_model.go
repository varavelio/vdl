package formatter

type fmtToken struct {
	Type    string
	Value   string
	Line    int
	Column  int
	Inline  bool
	EndLine int
}

type node interface {
	startLine() int
	endLine() int
	nodeKind() string
}

type baseNode struct {
	start int
	end   int
}

func (n baseNode) startLine() int { return n.start }
func (n baseNode) endLine() int   { return n.end }

type commentNode struct {
	baseNode
	Text   string
	Inline bool
}

func (n *commentNode) nodeKind() string { return "comment" }

type docstringNode struct {
	baseNode
	Raw string
}

func (n *docstringNode) nodeKind() string { return "docstring" }

type annotationNode struct {
	baseNode
	Name string
	Arg  *literalNode
}

type includeNode struct {
	baseNode
	Doc      *docstringNode
	Ann      []*annotationNode
	Path     string
	Trailing *commentNode
}

func (n *includeNode) nodeKind() string { return "include" }

type typeNode struct {
	baseNode
	Doc      *docstringNode
	Ann      []*annotationNode
	Name     string
	Type     fieldTypeNode
	Trailing *commentNode
}

func (n *typeNode) nodeKind() string { return "type" }

type constNode struct {
	baseNode
	Doc      *docstringNode
	Ann      []*annotationNode
	Name     string
	TypeName *string
	Value    literalNode
	Trailing *commentNode
}

func (n *constNode) nodeKind() string { return "const" }

type enumNode struct {
	baseNode
	Doc      *docstringNode
	Ann      []*annotationNode
	Name     string
	Members  []*enumMemberNode
	Trailing *commentNode
}

func (n *enumNode) nodeKind() string { return "enum" }

type enumMemberNode struct {
	baseNode
	Comment  *commentNode
	Spread   *referenceNode
	Doc      *docstringNode
	Ann      []*annotationNode
	Name     string
	Value    *enumValueNode
	Trailing *commentNode
}

type enumValueNode struct {
	Str *string
	Int *string
}

type fieldTypeNode struct {
	Named *string
	Map   *fieldTypeNode
	Obj   *objectTypeNode
	Dims  int
}

type objectTypeNode struct {
	Members []*typeMemberNode
}

type typeMemberNode struct {
	baseNode
	Comment    *commentNode
	Standalone *docstringNode
	Spread     *referenceNode
	Field      *fieldNode
	Trailing   *commentNode
}

type fieldNode struct {
	baseNode
	Doc      *docstringNode
	Ann      []*annotationNode
	Name     string
	Optional bool
	Type     fieldTypeNode
	Trailing *commentNode
}

type literalNode struct {
	Obj    *objectLiteralNode
	Array  *arrayLiteralNode
	Scalar *scalarLiteralNode
}

type objectLiteralNode struct {
	Entries []*objectEntryNode
}

type objectEntryNode struct {
	baseNode
	Comment  *commentNode
	Spread   *referenceNode
	Key      string
	Value    *literalNode
	Trailing *commentNode
}

type arrayLiteralNode struct {
	Elements []*arrayElementNode
}

type arrayElementNode struct {
	baseNode
	Comment  *commentNode
	Value    *literalNode
	Trailing *commentNode
}

type scalarLiteralNode struct {
	Str   *string
	Float *string
	Int   *string
	True  bool
	False bool
	Ref   *referenceNode
}

type referenceNode struct {
	Name   string
	Member *string
}

type docNode struct {
	Items []node
}

type tokenParser struct {
	tokens []fmtToken
	idx    int
}

type refCase int

const (
	refTypeDecl refCase = iota
	refEnumDecl
	refConstDecl
	refEnumMember
)

type literalRenderCtx struct {
	spreadRef     refCase
	scalarRef     refCase
	enumMemberRef refCase
}
