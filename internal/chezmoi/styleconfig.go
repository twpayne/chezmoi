package chezmoi

import "github.com/charmbracelet/glamour/ansi"

// ANSIStyleConfig is the glamour style config. It is based on glamour's notty
// style but uses only ANSI characters.
var ANSIStyleConfig = ansi.StyleConfig{
	Document: ansi.StyleBlock{
		Indent: uintPtr(2),
		StylePrimitive: ansi.StylePrimitive{
			BlockPrefix: "\n",
			BlockSuffix: "\n",
		},
	},
	Heading: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			BlockSuffix: "\n",
		},
	},
	H1: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "# ",
		},
	},
	H2: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "## ",
		},
	},
	H3: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "### ",
		},
	},
	H4: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "#### ",
		},
	},
	H5: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "##### ",
		},
	},
	H6: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "###### ",
		},
	},
	List: ansi.StyleList{
		LevelIndent: 4,
	},
	Item: ansi.StylePrimitive{
		BlockPrefix: "* ",
	},
	Enumeration: ansi.StylePrimitive{
		BlockPrefix: ". ",
	},
	ImageText: ansi.StylePrimitive{
		Format: "Image: {{.text}} ->",
	},
	Table: ansi.StyleTable{
		CenterSeparator: stringPtr("+"),
		ColumnSeparator: stringPtr("|"),
		RowSeparator:    stringPtr("-"),
	},
	Task: ansi.StyleTask{
		Unticked: "[ ]",
		Ticked:   "[x]",
	},
	Strikethrough: ansi.StylePrimitive{
		BlockPrefix: "~~",
		BlockSuffix: "~~",
	},
	Emph: ansi.StylePrimitive{
		BlockPrefix: "*",
		BlockSuffix: "*",
	},
	Strong: ansi.StylePrimitive{
		BlockPrefix: "**",
		BlockSuffix: "**",
	},
	HorizontalRule: ansi.StylePrimitive{
		Format: "\n--------\n",
	},
	BlockQuote: ansi.StyleBlock{
		Indent:      uintPtr(1),
		IndentToken: stringPtr("| "),
	},
	DefinitionDescription: ansi.StylePrimitive{
		BlockPrefix: "\n* ",
	},
	Code: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			BlockPrefix: "`",
			BlockSuffix: "`",
		},
	},
	CodeBlock: ansi.StyleCodeBlock{
		StyleBlock: ansi.StyleBlock{
			Indent: uintPtr(2),
		},
	},
}

func stringPtr(s string) *string { return &s }
func uintPtr(u uint) *uint       { return &u }
