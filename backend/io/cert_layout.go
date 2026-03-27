package io

// cert_layout.go — JSON-driven certificate layout engine (Finding 9, Option A).
//
// The layout is stored in certificate_layout.json alongside config.toml.
// When the file is absent, DefaultCertLayout() is written and returned.
//
// Layout file structure:
//
//	{
//	  "participant": { … },          // Urkunden Teilnehmende (text style)
//	  "participant_picture": { … },  // Urkunden Teilnehmende (picture style)
//	  "ov_winner": { … },            // Siegerurkunde Ortsverband
//	  "ov_participant": { … }        // Teilnahme-Urkunde Ortsverband
//	}
//
// Each page layout has two parts:
//
//  1. "content_area" – the rectangle (mm) inside which all elements are placed.
//     This defines the usable area left free by the background image.
//     Fields: left, top, right, bottom  (all in mm from the page edge).
//
//  2. "elements" array.  Every element has mandatory fields:
//     - "type": "text" | "dynamic" | "members_table" | "group_picture" | "ov_image"
//     - "x", "y": absolute position in mm.
//       Special value -1 means "use content_area.left" (x) or "use current cursor" (y).
//     - "width": cell width in mm.
//       Special value 0 means "use full content_area width (right − left)".
//     - "height": cell height in mm; 0 = derived from font size.
//     - "font_style": "" | "B" | "I" | "BI"
//     - "font_size": pt
//     - "align": "C" | "L" | "R"
//     - "color": [R, G, B]
//
// "text" elements also have "content" (static string).
// "dynamic" elements have "field":
//
//	year | name | ortsverband | group | rank | event_name | winner_label
//
// "members_table" is self‑sizing; only x, y, and width matter.
// "group_picture" uses "img_width" (mm) for scaling; centred within the content area.
// "ov_image"      renders ov_winner_image.png; centred within the content area.

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"THW-JugendOlympiade/backend/models"

	"github.com/jung-kurt/gofpdf"
)

const certLayoutFile = "certificate_layout.json"

// ---- Data types -------------------------------------------------------

// ContentArea defines the usable rectangle on the A4 page (mm from page edges).
// All element positions and widths default to this area when the element's
// own x/width are set to the sentinel values -1 / 0.
type ContentArea struct {
	Left   float64 `json:"left"`
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
}

// Width returns the horizontal extent of the content area.
func (a ContentArea) Width() float64 { return a.Right - a.Left }

// CertLayoutElement describes one element on a certificate page.
type CertLayoutElement struct {
	Type        string  `json:"type"`         // text | dynamic | members_table | group_picture | ov_image
	Content     string  `json:"content"`      // static text (type=text)
	Field       string  `json:"field"`        // dynamic field name (type=dynamic)
	X           float64 `json:"x"`            // mm from left; -1 = use content_area.left
	Y           float64 `json:"y"`            // mm from top;  -1 = use current cursor Y
	Width       float64 `json:"width"`        // cell / image width (mm); 0 = use content_area width
	Height      float64 `json:"height"`       // cell height (mm); 0 = auto from font size
	ImgWidth    float64 `json:"img_width"`    // image render width (mm)
	FontStyle   string  `json:"font_style"`   // "" | "B" | "I" | "BI"
	FontSize    float64 `json:"font_size"`    // pt
	Align       string  `json:"align"`        // "C" | "L" | "R"
	Color       [3]int  `json:"color"`        // [R, G, B]
	SpaceBefore float64 `json:"space_before"` // Ln() before element (mm); only used in flow mode
}

// CertPageLayout holds a content area and the list of elements for one certificate variant.
type CertPageLayout struct {
	Area     ContentArea         `json:"content_area"`
	Elements []CertLayoutElement `json:"elements"`
}

// CertLayoutFile is the top‑level JSON document.
type CertLayoutFile struct {
	Participant        CertPageLayout `json:"participant"`
	ParticipantPicture CertPageLayout `json:"participant_picture"`
	OVWinner           CertPageLayout `json:"ov_winner"`
	OVParticipant      CertPageLayout `json:"ov_participant"`
}

// ---- Default layout (mirrors current hard‑coded layout) ---------------

// DefaultCertLayout returns a CertLayoutFile whose values exactly reproduce
// the hard-coded layout that was previously baked into the renderer functions.
//
// All element x values are set to -1 (= use content_area.left) and widths to 0
// (= use content_area width). Adjust the content_area to reposition everything
// at once. Set explicit x/width on individual elements to override.
func DefaultCertLayout() CertLayoutFile {
	// Content area for Teilnehmende certificates.
	// The background template image leaves the left ~148 mm free for content.
	participantArea := ContentArea{Left: 10, Top: 55, Right: 147.83, Bottom: 275}

	// Participant (text style)
	participant := CertPageLayout{
		Area: participantArea,
		Elements: []CertLayoutElement{
			{Type: "dynamic", Field: "event_name", X: -1, Y: 60, Width: 0, Height: 12, FontStyle: "B", FontSize: 28, Align: "C", Color: [3]int{102, 126, 234}},
			{Type: "dynamic", Field: "year", X: -1, Y: 74, Width: 0, Height: 10, FontStyle: "B", FontSize: 24, Align: "C", Color: [3]int{102, 126, 234}},
			{Type: "dynamic", Field: "name", X: -1, Y: 95, Width: 0, Height: 10, FontStyle: "B", FontSize: 28, Align: "C", Color: [3]int{0, 0, 0}},
			{Type: "dynamic", Field: "ortsverband", X: -1, Y: 105, Width: 0, Height: 8, FontStyle: "", FontSize: 14, Align: "C", Color: [3]int{100, 100, 100}},
			{Type: "dynamic", Field: "group", X: -1, Y: 125, Width: 0, Height: 10, FontStyle: "B", FontSize: 16, Align: "C", Color: [3]int{0, 0, 0}},
			{Type: "dynamic", Field: "rank", X: -1, Y: 140, Width: 0, Height: 12, FontStyle: "B", FontSize: 22, Align: "C", Color: [3]int{180, 140, 10}},
			{Type: "text", Content: "Gruppenmitglieder", X: -1, Y: 157, Width: 0, Height: 8, FontStyle: "B", FontSize: 12, Align: "C", Color: [3]int{0, 0, 0}},
			{Type: "members_table", X: -1, Y: 167, Width: 0},
		},
	}

	// Participant picture style
	participantPicture := CertPageLayout{
		Area: participantArea,
		Elements: []CertLayoutElement{
			{Type: "dynamic", Field: "event_name", X: -1, Y: 60, Width: 0, Height: 12, FontStyle: "B", FontSize: 28, Align: "C", Color: [3]int{102, 126, 234}},
			{Type: "dynamic", Field: "year", X: -1, Y: 74, Width: 0, Height: 10, FontStyle: "B", FontSize: 24, Align: "C", Color: [3]int{102, 126, 234}},
			{Type: "dynamic", Field: "name", X: -1, Y: 95, Width: 0, Height: 10, FontStyle: "B", FontSize: 28, Align: "C", Color: [3]int{0, 0, 0}},
			{Type: "dynamic", Field: "ortsverband", X: -1, Y: 105, Width: 0, Height: 8, FontStyle: "", FontSize: 14, Align: "C", Color: [3]int{100, 100, 100}},
			{Type: "dynamic", Field: "group", X: -1, Y: 120, Width: 0, Height: 10, FontStyle: "B", FontSize: 16, Align: "C", Color: [3]int{0, 0, 0}},
			{Type: "dynamic", Field: "rank", X: -1, Y: 132, Width: 0, Height: 12, FontStyle: "B", FontSize: 22, Align: "C", Color: [3]int{180, 140, 10}},
			{Type: "group_picture", X: -1, Y: 148, Width: 0, ImgWidth: 120},
		},
	}

	// Content area for Ortsverband certificates (full printable page with 15 mm margins).
	ovArea := ContentArea{Left: 15, Top: 20, Right: 195, Bottom: 277}

	// Ortsverband winner
	ovWinner := CertPageLayout{
		Area: ovArea,
		Elements: []CertLayoutElement{
			{Type: "dynamic", Field: "event_name", X: -1, Y: 25, Width: 0, Height: 14, FontStyle: "B", FontSize: 28, Align: "C", Color: [3]int{102, 126, 234}},
			{Type: "dynamic", Field: "year", X: -1, Y: 44, Width: 0, Height: 12, FontStyle: "B", FontSize: 24, Align: "C", Color: [3]int{102, 126, 234}},
			{Type: "text", Content: "Siegerurkunde", X: -1, Y: 62, Width: 0, Height: 12, FontStyle: "B", FontSize: 16, Align: "C", Color: [3]int{0, 0, 0}},
			{Type: "dynamic", Field: "ortsverband", X: -1, Y: 78, Width: 0, Height: 14, FontStyle: "B", FontSize: 28, Align: "C", Color: [3]int{0, 0, 0}},
			{Type: "ov_image", X: -1, Y: 88, Width: 0, ImgWidth: 140},
			{Type: "dynamic", Field: "winner_label", X: -1, Y: 187, Width: 0, Height: 14, FontStyle: "B", FontSize: 22, Align: "C", Color: [3]int{180, 140, 10}},
			{Type: "text", Content: "Teilnehmende", X: -1, Y: 201, Width: 0, Height: 10, FontStyle: "B", FontSize: 16, Align: "C", Color: [3]int{0, 0, 0}},
			{Type: "members_table", X: -1, Y: 212, Width: 0},
		},
	}

	// Ortsverband participant
	ovParticipant := CertPageLayout{
		Area: ovArea,
		Elements: []CertLayoutElement{
			{Type: "dynamic", Field: "event_name", X: -1, Y: 40, Width: 0, Height: 14, FontStyle: "B", FontSize: 28, Align: "C", Color: [3]int{102, 126, 234}},
			{Type: "dynamic", Field: "year", X: -1, Y: 60, Width: 0, Height: 12, FontStyle: "B", FontSize: 24, Align: "C", Color: [3]int{102, 126, 234}},
			{Type: "text", Content: "Urkunde", X: -1, Y: 80, Width: 0, Height: 12, FontStyle: "B", FontSize: 16, Align: "C", Color: [3]int{0, 0, 0}},
			{Type: "dynamic", Field: "ortsverband", X: -1, Y: 100, Width: 0, Height: 14, FontStyle: "B", FontSize: 28, Align: "C", Color: [3]int{0, 0, 0}},
			{Type: "text", Content: "Teilnehmende", X: -1, Y: 125, Width: 0, Height: 10, FontStyle: "B", FontSize: 16, Align: "C", Color: [3]int{0, 0, 0}},
			{Type: "members_table", X: -1, Y: 138, Width: 0},
		},
	}

	return CertLayoutFile{
		Participant:        participant,
		ParticipantPicture: participantPicture,
		OVWinner:           ovWinner,
		OVParticipant:      ovParticipant,
	}
}

// ---- File I/O ---------------------------------------------------------

// LoadCertLayout reads certificate_layout.json.
// If the file does not exist the default layout is written and returned.
func LoadCertLayout() (CertLayoutFile, error) {
	if _, err := os.Stat(certLayoutFile); os.IsNotExist(err) {
		layout := DefaultCertLayout()
		if writeErr := SaveCertLayout(layout); writeErr != nil {
			return layout, nil // non-fatal: return default even if write fails
		}
		return layout, nil
	}

	data, err := os.ReadFile(certLayoutFile)
	if err != nil {
		return DefaultCertLayout(), fmt.Errorf("certificate_layout.json konnte nicht gelesen werden: %w", err)
	}

	var layout CertLayoutFile
	if err := json.Unmarshal(data, &layout); err != nil {
		return DefaultCertLayout(), fmt.Errorf("certificate_layout.json: ungültiges JSON: %w", err)
	}
	return layout, nil
}

// SaveCertLayout writes the layout to certificate_layout.json (indented for readability).
func SaveCertLayout(layout CertLayoutFile) error {
	data, err := json.MarshalIndent(layout, "", "  ")
	if err != nil {
		return fmt.Errorf("Layout konnte nicht serialisiert werden: %w", err)
	}
	return os.WriteFile(certLayoutFile, data, 0644)
}

// ReadCertLayoutRaw returns the raw JSON text of certificate_layout.json.
// If the file does not exist the default layout is created and returned.
func ReadCertLayoutRaw() (string, error) {
	if _, err := os.Stat(certLayoutFile); os.IsNotExist(err) {
		layout := DefaultCertLayout()
		if writeErr := SaveCertLayout(layout); writeErr != nil {
			// Return a serialised default even without writing
			data, _ := json.MarshalIndent(layout, "", "  ")
			return string(data), nil
		}
	}
	data, err := os.ReadFile(certLayoutFile)
	if err != nil {
		return "", fmt.Errorf("certificate_layout.json konnte nicht gelesen werden: %w", err)
	}
	return string(data), nil
}

// ValidateAndSaveCertLayoutRaw parses content as JSON into CertLayoutFile,
// then writes it back.  Returns the parsed layout and any error.
func ValidateAndSaveCertLayoutRaw(content string) (CertLayoutFile, error) {
	var layout CertLayoutFile
	if err := json.Unmarshal([]byte(content), &layout); err != nil {
		return CertLayoutFile{}, fmt.Errorf("Ungültiges JSON: %w", err)
	}
	if err := SaveCertLayout(layout); err != nil {
		return CertLayoutFile{}, err
	}
	return layout, nil
}

// ---- Renderer ---------------------------------------------------------

// CertContext holds the per-certificate dynamic values passed to the renderer.
type CertContext struct {
	EventName   string
	Year        int
	Name        string // participant name (empty for OV certs)
	Ortsverband string
	GroupID     int
	RankText    string
	PicturePath string                // group photo path (picture style)
	Members     []models.Teilnehmende // group members (text style)
	OVNames     []string              // OV participant names (OV certs)
}

// RenderCertPage renders all elements of a CertPageLayout onto the current
// pdf page using the given theme and context.
// The layout's ContentArea is used to resolve element x=-1 and width=0 sentinels.
func RenderCertPage(pdf *gofpdf.Fpdf, theme PDFTheme, layout CertPageLayout, ctx CertContext) {
	area := layout.Area
	// Safety: if area is unset (old JSON without content_area), fall back to
	// standard margins so existing absolute-coordinate elements still work.
	if area.Right <= area.Left {
		area.Left = 15
		area.Right = 195
	}
	for _, el := range layout.Elements {
		renderElement(pdf, theme, el, ctx, area)
	}
}

// resolveElementBounds returns the effective x and width for an element,
// substituting content area values for the sentinel -1 / 0.
func resolveElementBounds(el CertLayoutElement, area ContentArea) (x, width float64) {
	x = el.X
	if x < 0 {
		x = area.Left
	}
	width = el.Width
	if width <= 0 {
		width = area.Width()
	}
	return
}

func renderElement(pdf *gofpdf.Fpdf, theme PDFTheme, el CertLayoutElement, ctx CertContext, area ContentArea) {
	x, width := resolveElementBounds(el, area)

	switch el.Type {
	case "text":
		renderTextCell(pdf, theme, el, enc(el.Content), x, width)

	case "dynamic":
		text := resolveDynamicField(el.Field, ctx)
		renderTextCell(pdf, theme, el, text, x, width)

	case "members_table":
		if len(ctx.Members) > 0 {
			certMembersTable(pdf, theme, ctx.Members, x, width, el.Y)
		} else {
			renderOVMembersList(pdf, theme, el, ctx.OVNames, x, width)
		}

	case "group_picture":
		imgW := el.ImgWidth
		if imgW <= 0 {
			imgW = 120
		}
		// Centre image within the content area
		imgX := x + (width-imgW)/2
		startY := el.Y
		if startY < 0 {
			startY = pdf.GetY()
		}
		certDrawGroupPictureAt(pdf, theme, ctx.PicturePath, imgX, startY, imgW)

	case "ov_image":
		imgW := el.ImgWidth
		if imgW <= 0 {
			imgW = 140
		}
		// Centre image within the content area
		imgX := x + (width-imgW)/2
		startY := el.Y
		if startY < 0 {
			startY = pdf.GetY()
		}
		if _, statErr := os.Stat("ov_winner_image.png"); statErr == nil {
			pdf.Image("ov_winner_image.png", imgX, startY, imgW, 0, false, "", 0, "")
		}
	}
}

// renderTextCell positions and draws a single text cell.
// x and width are the already-resolved values (sentinel substitution already applied).
func renderTextCell(pdf *gofpdf.Fpdf, theme PDFTheme, el CertLayoutElement, text string, x, width float64) {
	if el.SpaceBefore > 0 {
		pdf.Ln(el.SpaceBefore)
	}

	pdf.SetFont(theme.FontFamily, el.FontStyle, el.FontSize)
	pdf.SetTextColor(el.Color[0], el.Color[1], el.Color[2])

	h := el.Height
	if h <= 0 {
		h = el.FontSize * 0.352778 * 1.5 // ~1.5× font height in mm
	}

	if el.Y >= 0 {
		pdf.SetXY(x, el.Y)
	} else {
		pdf.SetX(x)
	}

	pdf.CellFormat(width, h, text, "", 0, el.Align, false, 0, "")
}

// resolveDynamicField returns the rendered string for a dynamic field name.
func resolveDynamicField(field string, ctx CertContext) string {
	switch strings.ToLower(field) {
	case "event_name":
		return enc(ctx.EventName)
	case "year":
		return fmt.Sprintf("%d", ctx.Year)
	case "name":
		return enc(ctx.Name)
	case "ortsverband":
		return enc(fmt.Sprintf("Ortsverband %s", ctx.Ortsverband))
	case "group":
		return fmt.Sprintf("Gruppe %d", ctx.GroupID)
	case "rank":
		return enc(ctx.RankText)
	case "winner_label":
		return "Bester Ortsverband"
	default:
		return field
	}
}

// certDrawGroupPictureAt draws a group photo at an explicit position.
// If the file is missing a placeholder rectangle is drawn.
func certDrawGroupPictureAt(pdf *gofpdf.Fpdf, theme PDFTheme, picturePath string, imgX, startY, imgW float64) {
	const imgH = 80.0
	if _, err := os.Stat(picturePath); err == nil {
		pdf.Image(picturePath, imgX, startY, imgW, 0, false, "", 0, "")
	} else {
		pdf.SetFillColor(220, 220, 220)
		pdf.SetDrawColor(150, 150, 150)
		pdf.Rect(imgX, startY, imgW, imgH, "FD")
		theme.Font(pdf, "", theme.SizeSmall)
		pdf.SetTextColor(100, 100, 100)
		pdf.SetXY(imgX, startY+imgH/2-3)
		pdf.CellFormat(imgW, 6, "Gruppenfoto nicht gefunden", "", 0, "C", false, 0, "")
	}
}

// renderOVMembersList renders names for OV certs when using JSON layout
// (the members_table element in an OV layout uses OVNames, not Members).
func renderOVMembersList(pdf *gofpdf.Fpdf, theme PDFTheme, el CertLayoutElement, names []string, x, width float64) {
	if el.Y >= 0 {
		pdf.SetXY(x, el.Y)
	}
	theme.Font(pdf, "", 12)
	pdf.SetTextColor(0, 0, 0)
	for _, name := range names {
		pdf.SetX(x)
		pdf.CellFormat(width, 6, enc(name), "", 1, "C", false, 0, "")
	}
}
