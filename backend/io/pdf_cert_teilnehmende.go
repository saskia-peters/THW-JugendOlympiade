package io

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"THW-JugendOlympiade/backend/database"
	"THW-JugendOlympiade/backend/models"

	"github.com/jung-kurt/gofpdf"
)

// GenerateParticipantCertificates creates a PDF with one certificate per participant.
// certStyle: "text" (default) shows a group members table; "picture" embeds a group photo.
// pictureDir: directory containing group photos named group_picture_XXX.jpg.
// If certificate_template.png exists in the working directory it is used as background.
func GenerateParticipantCertificates(db *sql.DB, eventYear int, certStyle string, pictureDir string) error {
	if err := ensurePDFDirectory(); err != nil {
		return err
	}

	groups, err := database.GetGroupsForReport(db)
	if err != nil {
		return fmt.Errorf("failed to get groups: %w", err)
	}
	if len(groups) == 0 {
		return fmt.Errorf("no groups found to generate certificates")
	}

	evaluations, err := database.GetGroupEvaluations(db)
	if err != nil {
		return fmt.Errorf("failed to get group evaluations: %w", err)
	}

	groupRanks := make(map[int]int)
	for i, eval := range evaluations {
		groupRanks[eval.GroupID] = i + 1
	}

	_, err = os.Stat("certificate_template.png")
	useTemplate := err == nil
	const templateFile = "certificate_template.png"

	theme := DefaultTheme
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)

	// Content area x-boundaries driven by the certificate template image.
	const contentLeft = 10.0
	const contentRight = 147.83
	const contentWidth = contentRight - contentLeft

	currentYear := eventYear
	if currentYear == 0 {
		currentYear = time.Now().Year()
	}

	for _, group := range groups {
		rank := groupRanks[group.GroupID]
		rankText := certRankLabel(rank)

		picturePath := groupPicturePath(pictureDir, group.GroupID)

		for _, participant := range group.Teilnehmende {
			pdf.AddPage()
			if certStyle == "picture" {
				if useTemplate {
					pdf.Image(templateFile, 0, 0, 210, 297, false, imageTypeFromFile(templateFile), 0, "")
					certRenderTemplatePicture(pdf, theme, participant, group.GroupID, rankText, picturePath, contentLeft, contentWidth, currentYear)
				} else {
					certRenderProgrammaticPicture(pdf, theme, participant, group.GroupID, rankText, picturePath, contentLeft, contentWidth, currentYear)
				}
			} else {
				if useTemplate {
					pdf.Image(templateFile, 0, 0, 210, 297, false, imageTypeFromFile(templateFile), 0, "")
					certRenderTemplate(pdf, theme, participant, group.GroupID, rankText, group.Teilnehmende, contentLeft, contentWidth, currentYear)
				} else {
					certRenderProgrammatic(pdf, theme, participant, group.GroupID, rankText, group.Teilnehmende, contentLeft, contentWidth, currentYear)
				}
			}
		}
	}

	if err = pdf.OutputFileAndClose(filepath.Join(pdfOutputDir, "Urkunden_Teilnehmende.pdf")); err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}
	return nil
}

// certRankLabel returns the formatted rank string.
func certRankLabel(rank int) string {
	if rank >= 1 && rank <= 3 {
		return fmt.Sprintf("%d. Platz", rank)
	}
	return fmt.Sprintf("Platz %d", rank)
}

// certRenderTemplate overlays all certificate content on top of a background image.
func certRenderTemplate(pdf *gofpdf.Fpdf, theme PDFTheme, p models.Teilnehmende, groupID int, rankText string, members []models.Teilnehmende, left, width float64, year int) {
	// Heading
	pdf.SetXY(left, 60)
	theme.Font(pdf, "B", theme.SizeCertTitle)
	theme.TextColor(pdf, theme.ColorPrimary)
	pdf.CellFormat(width, 12, "Jugendolympiade", "", 0, "C", false, 0, "")

	// Year
	pdf.SetXY(left, 74)
	theme.Font(pdf, "B", theme.SizeCertYear)
	theme.TextColor(pdf, theme.ColorPrimary)
	pdf.CellFormat(width, 10, fmt.Sprintf("%d", year), "", 0, "C", false, 0, "")

	// Participant name
	pdf.SetXY(left, 95)
	theme.Font(pdf, "B", theme.SizeCertName)
	theme.TextColor(pdf, theme.ColorText)
	pdf.CellFormat(width, 10, enc(p.Name), "", 0, "C", false, 0, "")

	// Ortsverband
	pdf.SetXY(left, 105)
	theme.Font(pdf, "", theme.SizeCertOrtsverband)
	theme.TextColor(pdf, theme.ColorSubtext)
	pdf.CellFormat(width, 8, enc(fmt.Sprintf("Ortsverband %s", p.Ortsverband)), "", 0, "C", false, 0, "")

	// Group
	pdf.SetXY(left, 125)
	theme.Font(pdf, "B", theme.SizeCertGroup)
	theme.TextColor(pdf, theme.ColorText)
	pdf.CellFormat(width, 10, fmt.Sprintf("Gruppe %d", groupID), "", 0, "C", false, 0, "")

	// Rank
	pdf.SetXY(left, 140)
	theme.Font(pdf, "B", theme.SizeCertRank)
	theme.TextColor(pdf, theme.ColorAccent)
	pdf.CellFormat(width, 12, rankText, "", 0, "C", false, 0, "")

	// Group members label
	pdf.SetXY(left, 157)
	theme.Font(pdf, "B", theme.SizeCertLabel)
	theme.TextColor(pdf, theme.ColorText)
	pdf.CellFormat(width, 8, "Gruppenmitglieder", "", 1, "C", false, 0, "")

	// Members table
	certMembersTable(pdf, theme, members, left, width, 167)
}

// certRenderProgrammatic renders a certificate without a background image.
func certRenderProgrammatic(pdf *gofpdf.Fpdf, theme PDFTheme, p models.Teilnehmende, groupID int, rankText string, members []models.Teilnehmende, left, width float64, year int) {
	// Heading — 1.5cm top margin
	pdf.Ln(15)
	pdf.SetX(left)
	theme.Font(pdf, "B", theme.SizeCertTitle)
	theme.TextColor(pdf, theme.ColorPrimary)
	pdf.CellFormat(width, 20, "Jugendolympiade", "", 1, "C", false, 0, "")

	// Year
	pdf.Ln(3)
	pdf.SetX(left)
	theme.Font(pdf, "B", theme.SizeCertYear)
	pdf.CellFormat(width, 12, fmt.Sprintf("%d", year), "", 1, "C", false, 0, "")
	pdf.Ln(20)

	// Participant name
	pdf.SetX(left)
	theme.Font(pdf, "B", theme.SizeCertName)
	theme.TextColor(pdf, theme.ColorText)
	pdf.CellFormat(width, 15, enc(p.Name), "", 1, "C", false, 0, "")
	pdf.Ln(5)

	// Ortsverband
	pdf.SetX(left)
	theme.Font(pdf, "", theme.SizeCertOrtsverband)
	theme.TextColor(pdf, theme.ColorSubtext)
	pdf.CellFormat(width, 8, enc(fmt.Sprintf("Ortsverband: %s", p.Ortsverband)), "", 1, "C", false, 0, "")
	pdf.Ln(8)

	// Group
	pdf.SetX(left)
	theme.Font(pdf, "B", theme.SizeCertGroup)
	theme.TextColor(pdf, theme.ColorText)
	pdf.CellFormat(width, 10, fmt.Sprintf("Gruppe %d", groupID), "", 1, "C", false, 0, "")

	// Rank
	pdf.SetX(left)
	theme.Font(pdf, "B", theme.SizeCertRank)
	theme.TextColor(pdf, theme.ColorAccent)
	pdf.CellFormat(width, 12, rankText, "", 1, "C", false, 0, "")
	pdf.Ln(4)

	// Group members label
	pdf.SetX(left)
	theme.Font(pdf, "B", theme.SizeCertLabel)
	theme.TextColor(pdf, theme.ColorText)
	pdf.CellFormat(width, 10, "Gruppenmitglieder", "", 1, "C", false, 0, "")
	pdf.Ln(2)

	// Members table at current cursor position
	certMembersTable(pdf, theme, members, left, width, -1)
}

// certMembersTable renders the group members table.
// Pass startY >= 0 to position absolutely; pass -1 to use the current cursor.
func certMembersTable(pdf *gofpdf.Fpdf, theme PDFTheme, members []models.Teilnehmende, left, width, startY float64) {
	colWidths := []float64{width / 2, width / 2}

	if startY >= 0 {
		pdf.SetXY(left, startY)
	} else {
		pdf.SetX(left)
	}

	// Header row
	theme.Font(pdf, "B", theme.SizeCertTableHeader)
	theme.FillColor(pdf, theme.ColorCertTableHeader)
	theme.TextColor(pdf, theme.ColorOnHeader)
	for i, header := range []string{"Name", "Ortsverband"} {
		pdf.CellFormat(colWidths[i], 8, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Data rows
	theme.Font(pdf, "", theme.SizeCertTableBody)
	theme.TextColor(pdf, theme.ColorText)
	for i, m := range members {
		fill := i%2 == 0
		if fill {
			theme.FillColor(pdf, theme.ColorCertTableRowAlt)
		} else {
			theme.FillColor(pdf, [3]int{255, 255, 255})
		}
		pdf.SetX(left)
		pdf.CellFormat(colWidths[0], 7, enc(m.Name), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colWidths[1], 7, enc(m.Ortsverband), "1", 0, "C", fill, 0, "")
		pdf.Ln(-1)
	}
}

// groupPicturePath returns the expected path for a group's photo.
// Format: <pictureDir>/group_picture_XXX.jpg (zero-padded to 3 digits).
func groupPicturePath(pictureDir string, groupID int) string {
	return filepath.Join(pictureDir, fmt.Sprintf("group_picture_%03d.jpg", groupID))
}

// certRenderTemplatePicture renders the picture-style certificate over a background template.
//
// Vertical layout (y in mm):
//
//	 60  event title
//	 74  year
//	 95  participant name
//	105  ortsverband
//	125  group + rank
//	145  group photo (120 mm wide, centred; placeholder rectangle if missing)
func certRenderTemplatePicture(pdf *gofpdf.Fpdf, theme PDFTheme, p models.Teilnehmende, groupID int, rankText string, picturePath string, left, width float64, year int) {
	pdf.SetXY(left, 60)
	theme.Font(pdf, "B", theme.SizeCertTitle)
	theme.TextColor(pdf, theme.ColorPrimary)
	pdf.CellFormat(width, 12, "Jugendolympiade", "", 0, "C", false, 0, "")

	pdf.SetXY(left, 74)
	theme.Font(pdf, "B", theme.SizeCertYear)
	theme.TextColor(pdf, theme.ColorPrimary)
	pdf.CellFormat(width, 10, fmt.Sprintf("%d", year), "", 0, "C", false, 0, "")

	pdf.SetXY(left, 95)
	theme.Font(pdf, "B", theme.SizeCertName)
	theme.TextColor(pdf, theme.ColorText)
	pdf.CellFormat(width, 10, enc(p.Name), "", 0, "C", false, 0, "")

	pdf.SetXY(left, 105)
	theme.Font(pdf, "", theme.SizeCertOrtsverband)
	theme.TextColor(pdf, theme.ColorSubtext)
	pdf.CellFormat(width, 8, enc(fmt.Sprintf("Ortsverband %s", p.Ortsverband)), "", 0, "C", false, 0, "")

	pdf.SetXY(left, 120)
	theme.Font(pdf, "B", theme.SizeCertGroup)
	theme.TextColor(pdf, theme.ColorText)
	pdf.CellFormat(width, 10, fmt.Sprintf("Gruppe %d", groupID), "", 0, "C", false, 0, "")

	pdf.SetXY(left, 132)
	theme.Font(pdf, "B", theme.SizeCertRank)
	theme.TextColor(pdf, theme.ColorAccent)
	pdf.CellFormat(width, 12, rankText, "", 0, "C", false, 0, "")

	certDrawGroupPicture(pdf, theme, picturePath, left, width, 148)
}

// certRenderProgrammaticPicture renders the picture-style certificate without a background template.
func certRenderProgrammaticPicture(pdf *gofpdf.Fpdf, theme PDFTheme, p models.Teilnehmende, groupID int, rankText string, picturePath string, left, width float64, year int) {
	pdf.Ln(15)
	pdf.SetX(left)
	theme.Font(pdf, "B", theme.SizeCertTitle)
	theme.TextColor(pdf, theme.ColorPrimary)
	pdf.CellFormat(width, 20, "Jugendolympiade", "", 1, "C", false, 0, "")

	pdf.Ln(3)
	pdf.SetX(left)
	theme.Font(pdf, "B", theme.SizeCertYear)
	pdf.CellFormat(width, 12, fmt.Sprintf("%d", year), "", 1, "C", false, 0, "")
	pdf.Ln(10)

	pdf.SetX(left)
	theme.Font(pdf, "B", theme.SizeCertName)
	theme.TextColor(pdf, theme.ColorText)
	pdf.CellFormat(width, 15, enc(p.Name), "", 1, "C", false, 0, "")
	pdf.Ln(3)

	pdf.SetX(left)
	theme.Font(pdf, "", theme.SizeCertOrtsverband)
	theme.TextColor(pdf, theme.ColorSubtext)
	pdf.CellFormat(width, 8, enc(fmt.Sprintf("Ortsverband: %s", p.Ortsverband)), "", 1, "C", false, 0, "")
	pdf.Ln(5)

	pdf.SetX(left)
	theme.Font(pdf, "B", theme.SizeCertGroup)
	theme.TextColor(pdf, theme.ColorText)
	pdf.CellFormat(width, 10, fmt.Sprintf("Gruppe %d", groupID), "", 1, "C", false, 0, "")

	pdf.SetX(left)
	theme.Font(pdf, "B", theme.SizeCertRank)
	theme.TextColor(pdf, theme.ColorAccent)
	pdf.CellFormat(width, 12, rankText, "", 1, "C", false, 0, "")
	pdf.Ln(6)

	certDrawGroupPicture(pdf, theme, picturePath, left, width, -1)
}

// certDrawGroupPicture embeds the group photo centred on the page.
// If the photo file does not exist a placeholder rectangle with a label is drawn instead.
// Pass startY >= 0 for absolute positioning; -1 uses the current cursor Y.
func certDrawGroupPicture(pdf *gofpdf.Fpdf, theme PDFTheme, picturePath string, left, width, startY float64) {
	const imgW = 120.0
	const imgH = 80.0 // placeholder height; actual image scales by aspect ratio
	imgX := left + (width-imgW)/2

	if startY < 0 {
		startY = pdf.GetY()
	}

	if _, err := os.Stat(picturePath); err == nil {
		pdf.Image(picturePath, imgX, startY, imgW, 0, false, "", 0, "")
	} else {
		// Placeholder: grey rectangle with centred label
		pdf.SetFillColor(220, 220, 220)
		pdf.SetDrawColor(150, 150, 150)
		pdf.Rect(imgX, startY, imgW, imgH, "FD")
		theme.Font(pdf, "", theme.SizeSmall)
		theme.TextColor(pdf, theme.ColorSubtext)
		pdf.SetXY(imgX, startY+imgH/2-3)
		pdf.CellFormat(imgW, 6, enc(fmt.Sprintf("[Gruppenfoto %s]", filepath.Base(picturePath))), "", 0, "C", false, 0, "")
		theme.TextColor(pdf, theme.ColorText)
	}
}
