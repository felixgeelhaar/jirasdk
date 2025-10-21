package issue

import "encoding/json"

// ADF represents an Atlassian Document Format document.
// This is used for rich text fields in Jira Cloud API v3, including issue descriptions.
//
// Example ADF structure:
//
//	{
//	  "type": "doc",
//	  "version": 1,
//	  "content": [
//	    {
//	      "type": "paragraph",
//	      "content": [
//	        {
//	          "type": "text",
//	          "text": "This is the description text"
//	        }
//	      ]
//	    }
//	  ]
//	}
type ADF struct {
	Type    string      `json:"type"`
	Version int         `json:"version"`
	Content []ADFNode   `json:"content,omitempty"`
}

// ADFNode represents a node in an ADF document.
type ADFNode struct {
	Type    string                 `json:"type"`
	Content []ADFNode              `json:"content,omitempty"`
	Text    string                 `json:"text,omitempty"`
	Attrs   map[string]interface{} `json:"attrs,omitempty"`
	Marks   []ADFMark              `json:"marks,omitempty"`
}

// ADFMark represents text formatting marks (bold, italic, etc.).
type ADFMark struct {
	Type  string                 `json:"type"`
	Attrs map[string]interface{} `json:"attrs,omitempty"`
}

// NewADF creates a new ADF document.
func NewADF() *ADF {
	return &ADF{
		Type:    "doc",
		Version: 1,
		Content: []ADFNode{},
	}
}

// ADFFromText creates an ADF document from plain text.
// This is a convenience function for simple text descriptions.
// The text will be split into paragraphs on newlines.
//
// Example:
//
//	adf := issue.ADFFromText("First paragraph\n\nSecond paragraph")
func ADFFromText(text string) *ADF {
	if text == "" {
		return NewADF()
	}

	adf := NewADF()

	// Split text into paragraphs on double newlines
	paragraphs := splitParagraphs(text)

	for _, para := range paragraphs {
		if para == "" {
			continue
		}
		adf.AddParagraph(para)
	}

	return adf
}

// splitParagraphs splits text into paragraphs.
// Single newlines are preserved within paragraphs, double newlines create new paragraphs.
func splitParagraphs(text string) []string {
	// For simplicity, we'll treat each non-empty line as a paragraph
	// A more sophisticated implementation could handle double newlines differently
	var paragraphs []string
	current := ""

	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			if i+1 < len(text) && text[i+1] == '\n' {
				// Double newline - end paragraph
				if current != "" {
					paragraphs = append(paragraphs, current)
					current = ""
				}
				i++ // Skip the second newline
			} else {
				// Single newline - add space within paragraph
				if current != "" {
					current += " "
				}
			}
		} else {
			current += string(text[i])
		}
	}

	// Add any remaining text
	if current != "" {
		paragraphs = append(paragraphs, current)
	}

	return paragraphs
}

// AddParagraph adds a paragraph with text content to the ADF document.
func (a *ADF) AddParagraph(text string) *ADF {
	paragraph := ADFNode{
		Type: "paragraph",
		Content: []ADFNode{
			{
				Type: "text",
				Text: text,
			},
		},
	}
	a.Content = append(a.Content, paragraph)
	return a
}

// AddHeading adds a heading to the ADF document.
// Level should be 1-6 (h1-h6).
func (a *ADF) AddHeading(text string, level int) *ADF {
	if level < 1 {
		level = 1
	}
	if level > 6 {
		level = 6
	}

	heading := ADFNode{
		Type: "heading",
		Attrs: map[string]interface{}{
			"level": level,
		},
		Content: []ADFNode{
			{
				Type: "text",
				Text: text,
			},
		},
	}
	a.Content = append(a.Content, heading)
	return a
}

// AddBulletList adds a bullet list to the ADF document.
func (a *ADF) AddBulletList(items []string) *ADF {
	listItems := make([]ADFNode, 0, len(items))
	for _, item := range items {
		listItems = append(listItems, ADFNode{
			Type: "listItem",
			Content: []ADFNode{
				{
					Type: "paragraph",
					Content: []ADFNode{
						{
							Type: "text",
							Text: item,
						},
					},
				},
			},
		})
	}

	bulletList := ADFNode{
		Type:    "bulletList",
		Content: listItems,
	}
	a.Content = append(a.Content, bulletList)
	return a
}

// AddOrderedList adds a numbered list to the ADF document.
func (a *ADF) AddOrderedList(items []string) *ADF {
	listItems := make([]ADFNode, 0, len(items))
	for _, item := range items {
		listItems = append(listItems, ADFNode{
			Type: "listItem",
			Content: []ADFNode{
				{
					Type: "paragraph",
					Content: []ADFNode{
						{
							Type: "text",
							Text: item,
						},
					},
				},
			},
		})
	}

	orderedList := ADFNode{
		Type:    "orderedList",
		Content: listItems,
	}
	a.Content = append(a.Content, orderedList)
	return a
}

// AddCodeBlock adds a code block to the ADF document.
func (a *ADF) AddCodeBlock(code string, language string) *ADF {
	attrs := make(map[string]interface{})
	if language != "" {
		attrs["language"] = language
	}

	codeBlock := ADFNode{
		Type:  "codeBlock",
		Attrs: attrs,
		Content: []ADFNode{
			{
				Type: "text",
				Text: code,
			},
		},
	}
	a.Content = append(a.Content, codeBlock)
	return a
}

// String returns the JSON representation of the ADF document.
// This is useful for debugging.
func (a *ADF) String() string {
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return ""
	}
	return string(data)
}

// IsEmpty returns true if the ADF document has no content.
func (a *ADF) IsEmpty() bool {
	return a == nil || len(a.Content) == 0
}

// ToText extracts plain text from an ADF document.
// This is useful for displaying descriptions in a simple text format.
func (a *ADF) ToText() string {
	if a == nil || len(a.Content) == 0 {
		return ""
	}

	var text string
	for i, node := range a.Content {
		if i > 0 {
			text += "\n"
		}
		text += nodeToText(node)
	}
	return text
}

// nodeToText recursively extracts text from an ADF node.
func nodeToText(node ADFNode) string {
	if node.Text != "" {
		return node.Text
	}

	var text string
	for _, child := range node.Content {
		childText := nodeToText(child)
		if childText != "" {
			if text != "" {
				text += " "
			}
			text += childText
		}
	}
	return text
}
