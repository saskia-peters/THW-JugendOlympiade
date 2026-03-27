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
// Layout positions are loaded from certificate_layout.json (created with defaults on first run).
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

	certLayout, err := LoadCertLayout()
	if err != nil {
		return fmt.Errorf("certificate_layout.json: %w", err)
	}

	theme := DefaultTheme
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)

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
			if useTemplate {
				pdf.Image(templateFile, 0, 0, 210, 297, false, imageTypeFromFile(templateFile), 0, "")
			}
			ctx := CertContext{
				EventName:   "Jugendolympiade",
				Year:        currentYear,
				Name:        participant.Name,
				Ortsverband: participant.Ortsverband,
				GroupID:     group.GroupID,
				RankText:    rankText,
				PicturePath: picturePath,
				Members:     group.Teilnehmende,
			}
			if certStyle == "picture" {
				RenderCertPage(pdf, theme, certLayout.ParticipantPicture, ctx)
			} else {
				RenderCertPage(pdf, theme, certLayout.Participant, ctx)
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
