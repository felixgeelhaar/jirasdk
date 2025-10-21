package issue

import (
	"encoding/json"
	"testing"
)

// TestNewADF verifies ADF initialization
func TestNewADF(t *testing.T) {
	adf := NewADF()

	if adf.Type != "doc" {
		t.Errorf("Expected Type 'doc', got '%s'", adf.Type)
	}
	if adf.Version != 1 {
		t.Errorf("Expected Version 1, got %d", adf.Version)
	}
	if adf.Content == nil {
		t.Error("Expected Content to be initialized, got nil")
	}
	if len(adf.Content) != 0 {
		t.Errorf("Expected empty Content, got %d items", len(adf.Content))
	}
}

// TestADFFromText_Empty verifies empty text handling
func TestADFFromText_Empty(t *testing.T) {
	adf := ADFFromText("")

	if adf.Type != "doc" {
		t.Errorf("Expected Type 'doc', got '%s'", adf.Type)
	}
	if len(adf.Content) != 0 {
		t.Errorf("Expected empty Content, got %d items", len(adf.Content))
	}
}

// TestADFFromText_SingleParagraph verifies single paragraph conversion
func TestADFFromText_SingleParagraph(t *testing.T) {
	text := "This is a simple paragraph"
	adf := ADFFromText(text)

	if len(adf.Content) != 1 {
		t.Fatalf("Expected 1 paragraph, got %d", len(adf.Content))
	}

	para := adf.Content[0]
	if para.Type != "paragraph" {
		t.Errorf("Expected paragraph type, got '%s'", para.Type)
	}
	if len(para.Content) != 1 {
		t.Fatalf("Expected 1 text node, got %d", len(para.Content))
	}
	if para.Content[0].Text != text {
		t.Errorf("Expected text '%s', got '%s'", text, para.Content[0].Text)
	}
}

// TestADFFromText_MultipleParagraphs verifies multiple paragraph conversion
func TestADFFromText_MultipleParagraphs(t *testing.T) {
	text := "First paragraph\n\nSecond paragraph\n\nThird paragraph"
	adf := ADFFromText(text)

	if len(adf.Content) != 3 {
		t.Fatalf("Expected 3 paragraphs, got %d", len(adf.Content))
	}

	expected := []string{"First paragraph", "Second paragraph", "Third paragraph"}
	for i, para := range adf.Content {
		if para.Type != "paragraph" {
			t.Errorf("Paragraph %d: Expected paragraph type, got '%s'", i, para.Type)
		}
		if len(para.Content) != 1 {
			t.Fatalf("Paragraph %d: Expected 1 text node, got %d", i, len(para.Content))
		}
		if para.Content[0].Text != expected[i] {
			t.Errorf("Paragraph %d: Expected text '%s', got '%s'", i, expected[i], para.Content[0].Text)
		}
	}
}

// TestADFFromText_SingleNewlines verifies single newlines are converted to spaces
func TestADFFromText_SingleNewlines(t *testing.T) {
	text := "Line one\nLine two\nLine three"
	adf := ADFFromText(text)

	if len(adf.Content) != 1 {
		t.Fatalf("Expected 1 paragraph, got %d", len(adf.Content))
	}

	expected := "Line one Line two Line three"
	if adf.Content[0].Content[0].Text != expected {
		t.Errorf("Expected text '%s', got '%s'", expected, adf.Content[0].Content[0].Text)
	}
}

// TestAddParagraph verifies paragraph addition
func TestAddParagraph(t *testing.T) {
	adf := NewADF()
	adf.AddParagraph("First paragraph").AddParagraph("Second paragraph")

	if len(adf.Content) != 2 {
		t.Fatalf("Expected 2 paragraphs, got %d", len(adf.Content))
	}

	expected := []string{"First paragraph", "Second paragraph"}
	for i, para := range adf.Content {
		if para.Type != "paragraph" {
			t.Errorf("Paragraph %d: Expected paragraph type, got '%s'", i, para.Type)
		}
		if para.Content[0].Text != expected[i] {
			t.Errorf("Paragraph %d: Expected text '%s', got '%s'", i, expected[i], para.Content[0].Text)
		}
	}
}

// TestAddHeading verifies heading addition with different levels
func TestAddHeading(t *testing.T) {
	tests := []struct {
		name     string
		level    int
		expected int
	}{
		{"Level 1", 1, 1},
		{"Level 3", 3, 3},
		{"Level 6", 6, 6},
		{"Level too low", 0, 1},
		{"Level too high", 10, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adf := NewADF()
			adf.AddHeading("Test Heading", tt.level)

			if len(adf.Content) != 1 {
				t.Fatalf("Expected 1 heading, got %d", len(adf.Content))
			}

			heading := adf.Content[0]
			if heading.Type != "heading" {
				t.Errorf("Expected heading type, got '%s'", heading.Type)
			}

			level, ok := heading.Attrs["level"].(int)
			if !ok {
				t.Fatal("Expected level attribute to be int")
			}
			if level != tt.expected {
				t.Errorf("Expected level %d, got %d", tt.expected, level)
			}
			if heading.Content[0].Text != "Test Heading" {
				t.Errorf("Expected text 'Test Heading', got '%s'", heading.Content[0].Text)
			}
		})
	}
}

// TestAddBulletList verifies bullet list addition
func TestAddBulletList(t *testing.T) {
	items := []string{"Item 1", "Item 2", "Item 3"}
	adf := NewADF()
	adf.AddBulletList(items)

	if len(adf.Content) != 1 {
		t.Fatalf("Expected 1 bulletList, got %d", len(adf.Content))
	}

	bulletList := adf.Content[0]
	if bulletList.Type != "bulletList" {
		t.Errorf("Expected bulletList type, got '%s'", bulletList.Type)
	}
	if len(bulletList.Content) != 3 {
		t.Fatalf("Expected 3 list items, got %d", len(bulletList.Content))
	}

	for i, listItem := range bulletList.Content {
		if listItem.Type != "listItem" {
			t.Errorf("Item %d: Expected listItem type, got '%s'", i, listItem.Type)
		}
		if len(listItem.Content) != 1 {
			t.Fatalf("Item %d: Expected 1 paragraph, got %d", i, len(listItem.Content))
		}
		para := listItem.Content[0]
		if para.Type != "paragraph" {
			t.Errorf("Item %d: Expected paragraph type, got '%s'", i, para.Type)
		}
		if para.Content[0].Text != items[i] {
			t.Errorf("Item %d: Expected text '%s', got '%s'", i, items[i], para.Content[0].Text)
		}
	}
}

// TestAddOrderedList verifies ordered list addition
func TestAddOrderedList(t *testing.T) {
	items := []string{"First", "Second", "Third"}
	adf := NewADF()
	adf.AddOrderedList(items)

	if len(adf.Content) != 1 {
		t.Fatalf("Expected 1 orderedList, got %d", len(adf.Content))
	}

	orderedList := adf.Content[0]
	if orderedList.Type != "orderedList" {
		t.Errorf("Expected orderedList type, got '%s'", orderedList.Type)
	}
	if len(orderedList.Content) != 3 {
		t.Fatalf("Expected 3 list items, got %d", len(orderedList.Content))
	}

	for i, listItem := range orderedList.Content {
		if listItem.Type != "listItem" {
			t.Errorf("Item %d: Expected listItem type, got '%s'", i, listItem.Type)
		}
		if listItem.Content[0].Content[0].Text != items[i] {
			t.Errorf("Item %d: Expected text '%s', got '%s'", i, items[i], listItem.Content[0].Content[0].Text)
		}
	}
}

// TestAddCodeBlock verifies code block addition
func TestAddCodeBlock(t *testing.T) {
	code := "func main() {\n    fmt.Println(\"Hello\")\n}"

	tests := []struct {
		name     string
		language string
	}{
		{"With language", "go"},
		{"Without language", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adf := NewADF()
			adf.AddCodeBlock(code, tt.language)

			if len(adf.Content) != 1 {
				t.Fatalf("Expected 1 codeBlock, got %d", len(adf.Content))
			}

			codeBlock := adf.Content[0]
			if codeBlock.Type != "codeBlock" {
				t.Errorf("Expected codeBlock type, got '%s'", codeBlock.Type)
			}
			if codeBlock.Content[0].Text != code {
				t.Errorf("Expected code '%s', got '%s'", code, codeBlock.Content[0].Text)
			}

			if tt.language != "" {
				lang, ok := codeBlock.Attrs["language"].(string)
				if !ok {
					t.Fatal("Expected language attribute to be string")
				}
				if lang != tt.language {
					t.Errorf("Expected language '%s', got '%s'", tt.language, lang)
				}
			} else {
				if len(codeBlock.Attrs) != 0 {
					t.Errorf("Expected no attributes, got %d", len(codeBlock.Attrs))
				}
			}
		})
	}
}

// TestADFToText_Simple verifies simple text extraction
func TestADFToText_Simple(t *testing.T) {
	adf := NewADF()
	adf.AddParagraph("First paragraph")
	adf.AddParagraph("Second paragraph")

	text := adf.ToText()
	expected := "First paragraph\nSecond paragraph"
	if text != expected {
		t.Errorf("Expected text '%s', got '%s'", expected, text)
	}
}

// TestADFToText_Complex verifies complex text extraction
func TestADFToText_Complex(t *testing.T) {
	adf := NewADF()
	adf.AddHeading("Title", 1)
	adf.AddParagraph("Introduction paragraph")
	adf.AddBulletList([]string{"Item 1", "Item 2"})
	adf.AddParagraph("Conclusion")

	text := adf.ToText()
	expected := "Title\nIntroduction paragraph\nItem 1 Item 2\nConclusion"
	if text != expected {
		t.Errorf("Expected text '%s', got '%s'", expected, text)
	}
}

// TestADFToText_Empty verifies empty ADF handling
func TestADFToText_Empty(t *testing.T) {
	tests := []struct {
		name string
		adf  *ADF
	}{
		{"Nil ADF", nil},
		{"Empty ADF", NewADF()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := tt.adf.ToText()
			if text != "" {
				t.Errorf("Expected empty string, got '%s'", text)
			}
		})
	}
}

// TestADFIsEmpty verifies isEmpty check
func TestADFIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		adf      *ADF
		expected bool
	}{
		{"Nil ADF", nil, true},
		{"Empty ADF", NewADF(), true},
		{"ADF with content", NewADF().AddParagraph("text"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEmpty := tt.adf.IsEmpty()
			if isEmpty != tt.expected {
				t.Errorf("Expected IsEmpty=%v, got %v", tt.expected, isEmpty)
			}
		})
	}
}

// TestADFMarshalJSON verifies JSON marshaling
func TestADFMarshalJSON(t *testing.T) {
	adf := NewADF()
	adf.AddParagraph("Test paragraph")

	data, err := json.Marshal(adf)
	if err != nil {
		t.Fatalf("Failed to marshal ADF: %v", err)
	}

	// Unmarshal to verify structure
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result["type"] != "doc" {
		t.Errorf("Expected type 'doc', got '%v'", result["type"])
	}

	version, ok := result["version"].(float64)
	if !ok || int(version) != 1 {
		t.Errorf("Expected version 1, got %v", result["version"])
	}

	content, ok := result["content"].([]interface{})
	if !ok || len(content) != 1 {
		t.Fatalf("Expected content array with 1 item, got %v", result["content"])
	}
}

// TestADFUnmarshalJSON verifies JSON unmarshaling
func TestADFUnmarshalJSON(t *testing.T) {
	jsonData := `{
		"type": "doc",
		"version": 1,
		"content": [
			{
				"type": "paragraph",
				"content": [
					{
						"type": "text",
						"text": "Test paragraph"
					}
				]
			}
		]
	}`

	var adf ADF
	if err := json.Unmarshal([]byte(jsonData), &adf); err != nil {
		t.Fatalf("Failed to unmarshal ADF: %v", err)
	}

	if adf.Type != "doc" {
		t.Errorf("Expected Type 'doc', got '%s'", adf.Type)
	}
	if adf.Version != 1 {
		t.Errorf("Expected Version 1, got %d", adf.Version)
	}
	if len(adf.Content) != 1 {
		t.Fatalf("Expected 1 content node, got %d", len(adf.Content))
	}

	text := adf.ToText()
	if text != "Test paragraph" {
		t.Errorf("Expected text 'Test paragraph', got '%s'", text)
	}
}

// TestIssueFieldsSetDescriptionText verifies SetDescriptionText convenience method
func TestIssueFieldsSetDescriptionText(t *testing.T) {
	fields := &IssueFields{
		Summary: "Test Issue",
	}

	fields.SetDescriptionText("This is a test description")

	if fields.Description == nil {
		t.Fatal("Expected Description to be set, got nil")
	}
	if fields.Description.Type != "doc" {
		t.Errorf("Expected ADF type 'doc', got '%s'", fields.Description.Type)
	}

	text := fields.Description.ToText()
	if text != "This is a test description" {
		t.Errorf("Expected text 'This is a test description', got '%s'", text)
	}
}

// TestIssueFieldsSetDescription verifies SetDescription method
func TestIssueFieldsSetDescription(t *testing.T) {
	fields := &IssueFields{
		Summary: "Test Issue",
	}

	adf := NewADF().
		AddHeading("Problem", 2).
		AddParagraph("The application crashes")

	fields.SetDescription(adf)

	if fields.Description == nil {
		t.Fatal("Expected Description to be set, got nil")
	}
	if len(fields.Description.Content) != 2 {
		t.Errorf("Expected 2 content nodes, got %d", len(fields.Description.Content))
	}
}

// TestIssueGetDescription verifies GetDescription method
func TestIssueGetDescription(t *testing.T) {
	tests := []struct {
		name     string
		issue    *Issue
		expected *ADF
	}{
		{
			name:     "Nil Fields",
			issue:    &Issue{Fields: nil},
			expected: nil,
		},
		{
			name:     "Nil Description",
			issue:    &Issue{Fields: &IssueFields{Description: nil}},
			expected: nil,
		},
		{
			name: "With Description",
			issue: &Issue{Fields: &IssueFields{
				Description: NewADF().AddParagraph("Test"),
			}},
			expected: NewADF().AddParagraph("Test"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := tt.issue.GetDescription()
			if tt.expected == nil {
				if desc != nil {
					t.Errorf("Expected nil, got %v", desc)
				}
			} else {
				if desc == nil {
					t.Fatal("Expected description, got nil")
				}
				if desc.ToText() != tt.expected.ToText() {
					t.Errorf("Expected text '%s', got '%s'", tt.expected.ToText(), desc.ToText())
				}
			}
		})
	}
}

// TestIssueGetDescriptionText verifies GetDescriptionText method
func TestIssueGetDescriptionText(t *testing.T) {
	tests := []struct {
		name     string
		issue    *Issue
		expected string
	}{
		{
			name:     "Nil Fields",
			issue:    &Issue{Fields: nil},
			expected: "",
		},
		{
			name:     "Nil Description",
			issue:    &Issue{Fields: &IssueFields{Description: nil}},
			expected: "",
		},
		{
			name: "With Description",
			issue: &Issue{Fields: &IssueFields{
				Description: NewADF().AddParagraph("Test description"),
			}},
			expected: "Test description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := tt.issue.GetDescriptionText()
			if text != tt.expected {
				t.Errorf("Expected text '%s', got '%s'", tt.expected, text)
			}
		})
	}
}

// TestIssueFieldsDescriptionMarshaling verifies Description field marshaling
func TestIssueFieldsDescriptionMarshaling(t *testing.T) {
	tests := []struct {
		name   string
		fields *IssueFields
	}{
		{
			name: "With plain text description",
			fields: &IssueFields{
				Summary: "Test Issue",
				Project: &Project{Key: "PROJ"},
			},
		},
		{
			name: "With complex ADF description",
			fields: &IssueFields{
				Summary: "Test Issue",
				Project: &Project{Key: "PROJ"},
			},
		},
	}

	// Set description with plain text
	tests[0].fields.SetDescriptionText("Simple description")

	// Set description with complex ADF
	tests[1].fields.SetDescription(NewADF().
		AddHeading("Overview", 2).
		AddParagraph("This is a test issue").
		AddBulletList([]string{"Point 1", "Point 2"}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.fields)
			if err != nil {
				t.Fatalf("Failed to marshal fields: %v", err)
			}

			// Unmarshal to verify structure
			var result map[string]interface{}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("Failed to unmarshal result: %v", err)
			}

			// Verify description is in ADF format
			desc, ok := result["description"].(map[string]interface{})
			if !ok {
				t.Fatal("Expected description to be an object")
			}
			if desc["type"] != "doc" {
				t.Errorf("Expected description type 'doc', got '%v'", desc["type"])
			}
			if desc["version"] != float64(1) {
				t.Errorf("Expected description version 1, got %v", desc["version"])
			}

			content, ok := desc["content"].([]interface{})
			if !ok {
				t.Fatal("Expected description content to be an array")
			}
			if len(content) == 0 {
				t.Error("Expected description content to have items")
			}
		})
	}
}

// TestADFRoundTrip verifies marshaling and unmarshaling preserves data
func TestADFRoundTrip(t *testing.T) {
	original := NewADF().
		AddHeading("Title", 1).
		AddParagraph("First paragraph").
		AddBulletList([]string{"Item 1", "Item 2", "Item 3"}).
		AddParagraph("Second paragraph").
		AddCodeBlock("fmt.Println(\"Hello\")", "go")

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal
	var decoded ADF
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify text content is preserved
	originalText := original.ToText()
	decodedText := decoded.ToText()
	if originalText != decodedText {
		t.Errorf("Text content not preserved.\nOriginal: %s\nDecoded: %s", originalText, decodedText)
	}
}
