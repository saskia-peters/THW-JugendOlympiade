package io

// cert_layout.go — TOML-driven certificate layout engine (Finding 9, Option A).
//
// The layout is stored in certificate_layout.toml alongside config.toml.
// When the file is absent, the built-in default layout is written and returned.
//
// See defaultCertLayoutTOML below for a fully-commented example of the file format.

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"THW-JugendOlympiade/backend/models"

	"github.com/BurntSushi/toml"
	"github.com/go-pdf/fpdf"
)

const certLayoutFile = "certificate_layout.toml"

// ---- Data types -------------------------------------------------------

// ContentArea defines the usable rectangle on the A4 page (mm from page edges).
// All element positions and widths default to this area when the element's
// own x/width are set to the sentinel values -1 / 0.
type ContentArea struct {
	Left   float64 `toml:"left"   json:"left"`
	Top    float64 `toml:"top"    json:"top"`
	Right  float64 `toml:"right"  json:"right"`
	Bottom float64 `toml:"bottom" json:"bottom"`
}

// Width returns the horizontal extent of the content area.
func (a ContentArea) Width() float64 { return a.Right - a.Left }

// CertLayoutElement describes one element on a certificate page.
type CertLayoutElement struct {
	Type        string  `toml:"type"         json:"type"`         // text | dynamic | members_table | group_picture | ov_image
	Content     string  `toml:"content"      json:"content"`      // static text (type=text)
	Field       string  `toml:"field"        json:"field"`        // dynamic field name (type=dynamic)
	X           float64 `toml:"x"            json:"x"`            // mm from left; -1 = use content_area.left
	Y           float64 `toml:"y"            json:"y"`            // mm from top;  -1 = use current cursor Y
	Width       float64 `toml:"width"        json:"width"`        // cell / image width (mm); 0 = use content_area width
	Height      float64 `toml:"height"       json:"height"`       // cell height (mm); 0 = auto from font size
	ImgWidth    float64 `toml:"img_width"    json:"img_width"`    // image render width (mm)
	FontFamily  string  `toml:"font_family"  json:"font_family"`  // "" = theme default | "Arial" | "Helvetica" | "Times" | "Courier"
	FontStyle   string  `toml:"font_style"   json:"font_style"`   // "" | "B" | "I" | "BI"
	FontSize    float64 `toml:"font_size"    json:"font_size"`    // pt
	Align       string  `toml:"align"        json:"align"`        // "C" | "L" | "R"
	Color       [3]int  `toml:"color"        json:"color"`        // [R, G, B]
	SpaceBefore float64 `toml:"space_before" json:"space_before"` // Ln() before element (mm); only used in flow mode
}

// CertPageLayout holds a content area and the list of elements for one certificate variant.
type CertPageLayout struct {
	BackgroundImage string              `toml:"background_image" json:"background_image"` // path relative to working dir; "" = none
	Area            ContentArea         `toml:"content_area"     json:"content_area"`
	Elements        []CertLayoutElement `toml:"elements"         json:"elements"`
}

// CertLayoutFile is the top‑level TOML document.
type CertLayoutFile struct {
	Participant        CertPageLayout `toml:"participant"         json:"participant"`
	ParticipantPicture CertPageLayout `toml:"participant_picture" json:"participant_picture"`
	OVWinner           CertPageLayout `toml:"ov_winner"           json:"ov_winner"`
	OVParticipant      CertPageLayout `toml:"ov_participant"      json:"ov_participant"`
}

// ---- Default layout ---------------------------------------------------

// defaultCertLayoutTOML is the built-in certificate layout written to disk the
// first time the application runs.  It is a fully commented TOML file so that
// users can edit it directly without needing to look up field names.
//
// Quick reference
// ---------------
// Coordinates and sizes are all in millimetres (A4 = 210 × 297 mm).
//
// content_area defines the rectangle that the background image leaves free.
// Element positions that equal -1 (x) or use width = 0 automatically adopt
// the content_area edge / full width – so moving the whole block only requires
// changing content_area.
//
// type values:
//   text          – static string given by  content = "…"
//   dynamic       – value filled in at runtime; choose with  field = "…"
//                   valid fields: event_name | year | name | ortsverband |
//                                 group | rank | winner_label
//   members_table – renders the list of group members (Teilnehmende)
//   group_picture – embeds the group photo, width given by img_width
//   ov_image      – embeds ov_winner_image.png, width given by img_width
//
// font_style: ""=normal  "B"=bold  "I"=italic  "BI"=bold-italic
// align:      "C"=centre  "L"=left  "R"=right
// color:      [R, G, B] – three integers 0-255
//
// x = -1   → use content_area.left
// width = 0 → use full content_area width (right − left)

const defaultCertLayoutTOML = `
# ---------------------------------------------------------------------------
# certificate_layout.toml
# Definiert das visuelle Layout aller vier Urkunden-Varianten.
# Alle Maße in Millimetern (A4 = 210 × 297 mm).
#
# INHALTSBEREICHE (content_area)
#   Definieren den Bereich, den das Hintergrundbild freilässt.
#   x = -1  → linker Rand des Inhaltsbereichs
#   width = 0 → volle Breite des Inhaltsbereichs (right − left)
#
# ELEMENT-TYPEN (type)
#   text          – fester Text in  content = "…"
#   dynamic       – Laufzeitwert;  field = "event_name|year|name|ortsverband|group|rank|winner_label|event_location|event_date|location_date"
#   members_table – Tabelle der Gruppenmitglieder
#   group_picture – Gruppenfoto;  img_width = Breite in mm
#   ov_image      – ov_winner_image.png;  img_width = Breite in mm
#
# SCHRIFT  font_style: ""=normal  "B"=fett  "I"=kursiv  "BI"=fett-kursiv
# AUSRICH  align:  "C"=zentriert  "L"=links  "R"=rechts
# FARBE    color = [R, G, B]  – Werte 0-255
# ---------------------------------------------------------------------------

# ===========================================================================
# Urkunde Teilnehmende – Textstil (kein Foto)
# ===========================================================================

[participant]
background_image = "templates/background_urkunde_teilnehmende.png"  # "" = kein Hintergrundbild

[participant.content_area]
# Die Vorlage lässt die linken ~148 mm für Inhalt frei.
left   = 10
top    = 55
right  = 147.83
bottom = 275

[[participant.elements]]
type       = "dynamic"
field      = "event_name"
y          = 60      # mm von oben
height     = 12
font_style = "B"
font_size  = 28
align      = "C"
color      = [102, 126, 234]
x          = -1      # -1 = content_area.left
width      = 0       # 0  = volle content_area-Breite

[[participant.elements]]
type       = "dynamic"
field      = "year"
y          = 74
height     = 10
font_style = "B"
font_size  = 24
align      = "C"
color      = [102, 126, 234]
x          = -1
width      = 0

[[participant.elements]]
type       = "dynamic"
field      = "name"
y          = 95
height     = 10
font_style = "B"
font_size  = 28
align      = "C"
color      = [0, 0, 0]
x          = -1
width      = 0

[[participant.elements]]
type       = "dynamic"
field      = "ortsverband"
y          = 105
height     = 8
font_style = ""
font_size  = 14
align      = "C"
color      = [100, 100, 100]
x          = -1
width      = 0

[[participant.elements]]
type       = "dynamic"
field      = "group"
y          = 125
height     = 10
font_style = "B"
font_size  = 16
align      = "C"
color      = [0, 0, 0]
x          = -1
width      = 0

[[participant.elements]]
type       = "dynamic"
field      = "rank"
y          = 140
height     = 12
font_style = "B"
font_size  = 22
align      = "C"
color      = [180, 140, 10]
x          = -1
width      = 0

[[participant.elements]]
type       = "text"
content    = "Gruppenmitglieder"
y          = 157
height     = 8
font_style = "B"
font_size  = 12
align      = "C"
color      = [0, 0, 0]
x          = -1
width      = 0

[[participant.elements]]
type  = "members_table"
y     = 167
x     = -1
width = 0

[[participant.elements]]
type       = "dynamic"
field      = "location_date"
y          = 262
height     = 10
font_style = "B"
font_size  = 12
align      = "C"
color      = [102, 126, 234]
x          = -1
width      = 0

# ===========================================================================
# Urkunde Teilnehmende – Bildstil (mit Gruppenfoto)
# ===========================================================================

[participant_picture]
background_image = "templates/background_urkunde_teilnehmende.png"  # "" = kein Hintergrundbild

[participant_picture.content_area]
left   = 10
top    = 55
right  = 147.83
bottom = 275

[[participant_picture.elements]]
type       = "dynamic"
field      = "event_name"
y          = 60
height     = 12
font_style = "B"
font_size  = 28
align      = "C"
color      = [102, 126, 234]
x          = -1
width      = 0

[[participant_picture.elements]]
type       = "dynamic"
field      = "year"
y          = 74
height     = 10
font_style = "B"
font_size  = 24
align      = "C"
color      = [102, 126, 234]
x          = -1
width      = 0

[[participant_picture.elements]]
type       = "dynamic"
field      = "name"
y          = 95
height     = 10
font_style = "B"
font_size  = 28
align      = "C"
color      = [0, 0, 0]
x          = -1
width      = 0

[[participant_picture.elements]]
type       = "dynamic"
field      = "ortsverband"
y          = 105
height     = 8
font_style = ""
font_size  = 14
align      = "C"
color      = [100, 100, 100]
x          = -1
width      = 0

[[participant_picture.elements]]
type       = "dynamic"
field      = "group"
y          = 120
height     = 10
font_style = "B"
font_size  = 16
align      = "C"
color      = [0, 0, 0]
x          = -1
width      = 0

[[participant_picture.elements]]
type       = "dynamic"
field      = "rank"
y          = 132
height     = 12
font_style = "B"
font_size  = 22
align      = "C"
color      = [180, 140, 10]
x          = -1
width      = 0

[[participant_picture.elements]]
# img_width: Breite des Fotos in mm; wird innerhalb des Inhaltsbereichs zentriert
type      = "group_picture"
y         = 148
img_width = 120
x         = -1
width     = 0

[[participant_picture.elements]]
type       = "dynamic"
field      = "location_date"
y          = 262
height     = 10
font_style = "B"
font_size  = 12
align      = "C"
color      = [102, 126, 234]
x          = -1
width      = 0

# ===========================================================================
# Siegerurkunde Ortsverband
# ===========================================================================

[ov_winner]
background_image = "templates/background_urkunde_ovs.png"  # "" = kein Hintergrundbild

[ov_winner.content_area]
# Vollbedruckbare Seite mit 15 mm Rändern
left   = 15
top    = 20
right  = 195
bottom = 277

[[ov_winner.elements]]
type       = "dynamic"
field      = "event_name"
y          = 25
height     = 14
font_style = "B"
font_size  = 28
align      = "C"
color      = [102, 126, 234]
x          = -1
width      = 0

[[ov_winner.elements]]
type       = "dynamic"
field      = "year"
y          = 44
height     = 12
font_style = "B"
font_size  = 24
align      = "C"
color      = [102, 126, 234]
x          = -1
width      = 0

[[ov_winner.elements]]
type       = "text"
content    = "Siegerurkunde"
y          = 62
height     = 12
font_style = "B"
font_size  = 16
align      = "C"
color      = [0, 0, 0]
x          = -1
width      = 0

[[ov_winner.elements]]
type       = "dynamic"
field      = "ortsverband"
y          = 78
height     = 14
font_style = "B"
font_size  = 28
align      = "C"
color      = [0, 0, 0]
x          = -1
width      = 0

[[ov_winner.elements]]
# img_width: Breite des OV-Siegerbilds in mm
type      = "ov_image"
y         = 88
img_width = 140
x         = -1
width     = 0

[[ov_winner.elements]]
type       = "dynamic"
field      = "winner_label"
y          = 187
height     = 14
font_style = "B"
font_size  = 22
align      = "C"
color      = [180, 140, 10]
x          = -1
width      = 0

[[ov_winner.elements]]
type       = "text"
content    = "Teilnehmende"
y          = 201
height     = 10
font_style = "B"
font_size  = 16
align      = "C"
color      = [0, 0, 0]
x          = -1
width      = 0

[[ov_winner.elements]]
type  = "members_table"
y     = 212
x     = -1
width = 0

# ===========================================================================
# Teilnahmeurkunde Ortsverband
# ===========================================================================

[ov_participant]
background_image = "templates/background_urkunde_ovs.png"  # "" = kein Hintergrundbild

[ov_participant.content_area]
left   = 15
top    = 20
right  = 195
bottom = 277

[[ov_participant.elements]]
type       = "dynamic"
field      = "event_name"
y          = 40
height     = 14
font_style = "B"
font_size  = 28
align      = "C"
color      = [102, 126, 234]
x          = -1
width      = 0

[[ov_participant.elements]]
type       = "dynamic"
field      = "year"
y          = 60
height     = 12
font_style = "B"
font_size  = 24
align      = "C"
color      = [102, 126, 234]
x          = -1
width      = 0

[[ov_participant.elements]]
type       = "text"
content    = "Urkunde"
y          = 80
height     = 12
font_style = "B"
font_size  = 16
align      = "C"
color      = [0, 0, 0]
x          = -1
width      = 0

[[ov_participant.elements]]
type       = "dynamic"
field      = "ortsverband"
y          = 100
height     = 14
font_style = "B"
font_size  = 28
align      = "C"
color      = [0, 0, 0]
x          = -1
width      = 0

[[ov_participant.elements]]
type       = "text"
content    = "Teilnehmende"
y          = 125
height     = 10
font_style = "B"
font_size  = 16
align      = "C"
color      = [0, 0, 0]
x          = -1
width      = 0

[[ov_participant.elements]]
type  = "members_table"
y     = 138
x     = -1
width = 0
`

// ---- File I/O ---------------------------------------------------------

// LoadCertLayout reads certificate_layout.toml.
// If the file does not exist the default layout is written and returned.
func LoadCertLayout() (CertLayoutFile, error) {
	if _, err := os.Stat(certLayoutFile); os.IsNotExist(err) {
		if writeErr := os.WriteFile(certLayoutFile, []byte(defaultCertLayoutTOML), 0644); writeErr != nil {
			// Non-fatal: parse the in-memory default and return it
			var layout CertLayoutFile
			_, _ = toml.Decode(defaultCertLayoutTOML, &layout)
			return layout, nil
		}
	}

	data, err := os.ReadFile(certLayoutFile)
	if err != nil {
		var fallback CertLayoutFile
		_, _ = toml.Decode(defaultCertLayoutTOML, &fallback)
		return fallback, fmt.Errorf("certificate_layout.toml konnte nicht gelesen werden: %w", err)
	}

	var layout CertLayoutFile
	if _, err := toml.Decode(string(data), &layout); err != nil {
		var fallback CertLayoutFile
		_, _ = toml.Decode(defaultCertLayoutTOML, &fallback)
		return fallback, fmt.Errorf("certificate_layout.toml: ungültiges TOML: %w", err)
	}
	applyLayoutDefaults(&layout)
	return layout, nil
}

// applyLayoutDefaults fills in background_image with the historical default filenames
// when the field is absent (empty string). Returns true if any field was changed.
func applyLayoutDefaults(l *CertLayoutFile) bool {
	changed := false
	if l.Participant.BackgroundImage == "" {
		l.Participant.BackgroundImage = "templates/background_urkunde_teilnehmende.png"
		changed = true
	}
	if l.ParticipantPicture.BackgroundImage == "" {
		l.ParticipantPicture.BackgroundImage = "templates/background_urkunde_teilnehmende.png"
		changed = true
	}
	if l.OVWinner.BackgroundImage == "" {
		l.OVWinner.BackgroundImage = "templates/background_urkunde_ovs.png"
		changed = true
	}
	if l.OVParticipant.BackgroundImage == "" {
		l.OVParticipant.BackgroundImage = "templates/background_urkunde_ovs.png"
		changed = true
	}
	return changed
}

// SaveCertLayout encodes layout as TOML and writes it to certificate_layout.toml.
func SaveCertLayout(layout CertLayoutFile) error {
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(layout); err != nil {
		return fmt.Errorf("Layout konnte nicht serialisiert werden: %w", err)
	}
	return os.WriteFile(certLayoutFile, buf.Bytes(), 0644)
}

// ReadCertLayoutRaw returns the raw TOML text of certificate_layout.toml.
// If the file does not exist the default layout is written and the default text returned.
// If fields added in newer versions are missing from the file, they are injected
// into the returned text (but NOT written to disk – the file is only updated when
// the user explicitly saves through the editor). This preserves all existing comments.
func ReadCertLayoutRaw() (string, error) {
	if _, err := os.Stat(certLayoutFile); os.IsNotExist(err) {
		_ = os.WriteFile(certLayoutFile, []byte(defaultCertLayoutTOML), 0644)
		return defaultCertLayoutTOML, nil
	}
	data, err := os.ReadFile(certLayoutFile)
	if err != nil {
		return "", fmt.Errorf("certificate_layout.toml konnte nicht gelesen werden: %w", err)
	}
	// Parse to find any missing fields.
	var layout CertLayoutFile
	_, _ = toml.Decode(string(data), &layout)
	return injectMissingBackgroundImages(string(data), layout), nil
}

// injectMissingBackgroundImages inserts background_image lines into raw TOML text
// for any variant whose field was absent (empty after parsing). The insertion is
// textual so all existing comments are preserved.
func injectMissingBackgroundImages(raw string, layout CertLayoutFile) string {
	type entry struct {
		section string
		defVal  string
		current string
	}
	for _, e := range []entry{
		{"participant", "templates/background_urkunde_teilnehmende.png", layout.Participant.BackgroundImage},
		{"participant_picture", "templates/background_urkunde_teilnehmende.png", layout.ParticipantPicture.BackgroundImage},
		{"ov_winner", "templates/background_urkunde_ovs.png", layout.OVWinner.BackgroundImage},
		{"ov_participant", "templates/background_urkunde_ovs.png", layout.OVParticipant.BackgroundImage},
	} {
		if e.current != "" {
			continue // already present
		}
		// Find the earliest sub-table or array-table belonging to this section.
		idx := -1
		for _, prefix := range []string{"[" + e.section + ".", "[[" + e.section + "."} {
			i := strings.Index(raw, prefix)
			if i >= 0 && (idx < 0 || i < idx) {
				idx = i
			}
		}
		insert := "[" + e.section + "]\nbackground_image = \"" + e.defVal + "\"  # \"\" = kein Hintergrundbild\n\n"
		if idx < 0 {
			raw += "\n" + insert
		} else {
			raw = raw[:idx] + insert + raw[idx:]
		}
	}
	return raw
}

// ValidateAndSaveCertLayoutRaw parses content as TOML into CertLayoutFile,
// then writes the raw text back.  Returns the parsed layout and any error.
func ValidateAndSaveCertLayoutRaw(content string) (CertLayoutFile, error) {
	var layout CertLayoutFile
	if _, err := toml.Decode(content, &layout); err != nil {
		return CertLayoutFile{}, fmt.Errorf("Ungültiges TOML: %w", err)
	}
	if err := os.WriteFile(certLayoutFile, []byte(content), 0644); err != nil {
		return CertLayoutFile{}, err
	}
	return layout, nil
}

// ---- Renderer ---------------------------------------------------------

// CertContext holds the per-certificate dynamic values passed to the renderer.
type CertContext struct {
	EventName     string
	Year          int
	Name          string // participant name (empty for OV certs)
	Ortsverband   string
	GroupID       int
	RankText      string
	PicturePath   string                // group photo path (picture style)
	Members       []models.Teilnehmende // group members (text style)
	OVNames       []string              // OV participant names (OV certs)
	EventLocation string                // event location from config (printed at bottom)
	EventDate     string                // date of generation, formatted in German
}

// RenderCertPage renders all elements of a CertPageLayout onto the current
// pdf page using the given theme and context.
// The layout's ContentArea is used to resolve element x=-1 and width=0 sentinels.
func RenderCertPage(pdf *fpdf.Fpdf, theme PDFTheme, layout CertPageLayout, ctx CertContext) {
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

func renderElement(pdf *fpdf.Fpdf, theme PDFTheme, el CertLayoutElement, ctx CertContext, area ContentArea) {
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
		imgFile := "ov_winner_image.png"
		if el.Content != "" {
			imgFile = el.Content
		}
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
		if _, statErr := os.Stat(imgFile); statErr == nil {
			pdf.Image(imgFile, imgX, startY, imgW, 0, false, "", 0, "")
		}
	}
}

// renderTextCell positions and draws a single text cell.
// x and width are the already-resolved values (sentinel substitution already applied).
func renderTextCell(pdf *fpdf.Fpdf, theme PDFTheme, el CertLayoutElement, text string, x, width float64) {
	if el.SpaceBefore > 0 {
		pdf.Ln(el.SpaceBefore)
	}

	family := theme.FontFamily
	if el.FontFamily != "" {
		family = el.FontFamily
	}
	pdf.SetFont(family, el.FontStyle, el.FontSize)
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
	case "event_location":
		return enc(ctx.EventLocation)
	case "event_date":
		return enc(ctx.EventDate)
	case "location_date":
		if ctx.EventLocation != "" {
			return enc(ctx.EventLocation + ", " + ctx.EventDate)
		}
		return enc(ctx.EventDate)
	default:
		return field
	}
}

// certDrawGroupPictureAt draws a group photo at an explicit position.
// If the file is missing a placeholder rectangle is drawn.
func certDrawGroupPictureAt(pdf *fpdf.Fpdf, theme PDFTheme, picturePath string, imgX, startY, imgW float64) {
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
func renderOVMembersList(pdf *fpdf.Fpdf, theme PDFTheme, el CertLayoutElement, names []string, x, width float64) {
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
