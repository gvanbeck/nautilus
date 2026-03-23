package rml

// ─── Top-level document ──────────────────────────────────────────────────────

// rmlDoc is the parsed in-memory representation of an RML document.
type rmlDoc struct {
	docinit    docinit
	template   tmpl
	stylesheet stylesheet
	story      []storyNode
}

// ─── Docinit ─────────────────────────────────────────────────────────────────

type docinit struct {
	fonts    []fontReg
	families []fontFamilyReg
}

// fontFamilyReg is a <registerFontFamily name="…" fontName="…" bold="…" …/>.
type fontFamilyReg struct {
	name       string
	fontName   string // regular variant
	bold       string
	italic     string
	boldItalic string
}

// fontReg is a <registerTTFont fontName="…" fontFile="…"/> directive.
type fontReg struct {
	name string
	file string
}

// ─── Template ────────────────────────────────────────────────────────────────

type tmpl struct {
	pageSize          string // "A4", "letter", "(w,h)"
	leftMargin        string
	rightMargin       string
	topMargin         string
	bottomMargin      string
	firstPageTemplate string // id of template to use for page 1
	title             string
	author            string
	subject           string
	creator           string
	templates         []pageTmpl
}


type pageTmpl struct {
	id           string
	frames       []frameDef
	pageGraphics *graphicsBlock // <pageGraphics> decorator
}

// frameDef is a <frame> within a pageTemplate.
// RML uses bottom-left origin; x1/y1 are the bottom-left corner.
type frameDef struct {
	id     string
	x1     string // left edge (same in both coordinate systems)
	y1     string // bottom edge from page bottom (must be converted)
	width  string
	height string
}

// ─── Stylesheet ──────────────────────────────────────────────────────────────

type stylesheet struct {
	paraStyles  []paraStyle
	tableStyles []blockTableStyle
}

// paraStyle is a <paraStyle name="…" …/> definition.
type paraStyle struct {
	name           string
	parent         string
	fontName       string
	fontSize       string
	leading        string
	alignment      string
	spaceBefore    string
	spaceAfter     string
	textColor      string
	backColor      string
	leftIndent     string
	rightIndent    string
	firstLineIndent string
	underline      string // "1" or "true"
	strike         string // "1" or "true"
	keepWithNext   string // "1" or "true"
}

// blockTableStyle is a <blockTableStyle id="…"> definition.
type blockTableStyle struct {
	id       string
	commands []tableCmd
}

// tableCmd represents one styling command inside a blockTableStyle.
// Which fields are relevant depends on kind.
type tableCmd struct {
	kind      string // "lineStyle", "blockBackground", "blockFont",
	//              "blockTextColor", "blockAlignment", "blockValign",
	//              "blockPadding"

	// range (both inclusive, may be negative = from end)
	startCol, startRow int
	stopCol, stopRow   int

	// lineStyle fields
	lineKind  string  // GRID, OUTLINE, INNERGRID, BOX
	thickness float64 // 0 → 0.5 default

	// shared
	colorName string // named or hex color
	fontName  string
	fontSize  string
	alignment string // left, center, right
	valign    string // top, middle, bottom
	padding   string // uniform padding value
}

// ─── Story nodes ─────────────────────────────────────────────────────────────

// storyNode is any element that can appear in <story>.
type storyNode interface{ isStoryNode() }

// paraNode is <para style="…">text</para>.
type paraNode struct {
	style string
	text  string // plain text (XML markup stripped)
}

// spacerNode is <spacer length="…" width="…"/>.
type spacerNode struct {
	length string
	width  string
}

// pageBreakNode is <pageBreak/>.
type pageBreakNode struct{}

// frameBreakNode is <frameBreak/>.
type frameBreakNode struct{}

// hrNode is <hr/> or <hRule/>.
type hrNode struct {
	width     string
	thickness string
	colorName string
}

// blockTableNode is <blockTable colWidths="…" rowHeights="…" style="…">.
type blockTableNode struct {
	colWidths   string // "150,200,100"
	rowHeights  string // optional; comma-separated fixed heights
	style       string // blockTableStyle id reference
	repeatRows  string // number of leading rows to repeat after page overflow
	align       string // left, center, right — horizontal alignment of table
	spaceBefore string
	spaceAfter  string
	rows        []trNode
}

// trNode is a <tr> within a blockTable.
type trNode struct {
	height string // optional fixed height override for this row
	bg     string // optional background colorName
	cells  []tdNode
}

// tdNode is a <td> within a <tr>.
type tdNode struct {
	colSpan      string
	rowSpan      string
	style        string // inline style reference
	fontName     string
	fontSize     string
	bold         string // "1", "true" → append "Bold" to fontName
	bg           string
	textColor    string
	halign       string
	valign       string
	topPadding   string
	rightPadding string
	bottomPadding string
	leftPadding  string
	text         string // plain text content
}

// imageNode is <image file="…" width="…" height="…" align="…"/>.
type imageNode struct {
	file        string
	width       string
	height      string
	align       string
	spaceBefore string
	spaceAfter  string
}

// keepTogetherNode is <keepTogether maxHeight="…">.
type keepTogetherNode struct {
	maxHeight string
	children  []storyNode
}

// condPageBreakNode is <condPageBreak height="…"/>.
type condPageBreakNode struct {
	height string
}

// nextPageTemplateNode is <nextPageTemplate id="…"/>.
type nextPageTemplateNode struct {
	id string
}

// indentNode is <indent left="…" right="…">.
type indentNode struct {
	left     string
	right    string
	children []storyNode
}

// listNode is <ul> or <ol>.
type listNode struct {
	ordered      bool
	start        string // first number for ordered list (default "1")
	bulletIndent string
	style        string // listStyle reference
	items        []listItemNode
}

// listItemNode is <li>.
type listItemNode struct {
	style string
	text  string
}

// ─── Marker methods ──────────────────────────────────────────────────────────

func (*paraNode) isStoryNode()            {}
func (*spacerNode) isStoryNode()          {}
func (*pageBreakNode) isStoryNode()       {}
func (*frameBreakNode) isStoryNode()      {}
func (*hrNode) isStoryNode()              {}
func (*blockTableNode) isStoryNode()      {}
func (*imageNode) isStoryNode()           {}
func (*keepTogetherNode) isStoryNode()    {}
func (*condPageBreakNode) isStoryNode()   {}
func (*nextPageTemplateNode) isStoryNode() {}
func (*indentNode) isStoryNode()          {}
func (*listNode) isStoryNode()            {}
