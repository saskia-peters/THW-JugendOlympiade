# Certificate Template Usage

The participant certificate generator supports custom templates using background images.

## How to Use a Custom Template

### 1. Create Your Template Design

Create your certificate template using any design tool (Canva, Photoshop, PowerPoint, etc.) with the following specifications:

- **Size**: A4 (210mm x 297mm) or 2480 x 3508 pixels at 300 DPI
- **Format**: PNG or JPG
- **Orientation**: Portrait

### 2. Design Guidelines

Leave space for the dynamic content that will be overlaid:

- **Year** (Top center, ~35mm from top): "2026" in large text
- **Participant Name** (Center, ~85mm from top): Large, prominent
- **Ortsverband** (~105mm from top): Smaller text below name
- **Group Number** (~125mm from top): "Gruppe X"
- **Rank** (~140mm from top): "1. Platz", "2. Platz", etc.
- **Group Members Table** (Starting ~165mm from top): Table with participant details

### 3. Save the Template

Save your template as one of these files in the application directory:

```
certificate_template.png    (Recommended)
certificate_template.jpg
```

**Note:** If you have a PDF template, you'll need to convert it to PNG or JPG first using:
- An online converter (e.g., pdf2png.com)
- Adobe Acrobat (Export as PNG)
- Any PDF reader's "Save as Image" feature

### 4. Generate Certificates

1. Run the application
2. Load your Excel file  
3. Click "Teilnehmer-Zertifikate"
4. The certificates will be generated with your template as the background

## Adjusting Text Positions

If you need to adjust where the text appears on your template, you can modify the Y-coordinates in the code:

File: `backend/io/output.go`  
Function: `GenerateParticipantCertificates`

Look for lines like:
```go
pdf.SetXY(0, 85)  // Second number is the Y position in mm
```

## Example Template Layout

```
┌─────────────────────────────────────┐
│                                     │
│          [Your Decorative           │
│            Header/Logo]             │
│                                     │
│            2026  ← Year (35mm)      │
│                                     │
│                                     │
│                                     │
│      [Participant Name] (85mm)      │
│      [Ortsverband] (105mm)          │
│                                     │
│        Gruppe X (125mm)             │
│        Platz Y (140mm)              │
│                                     │
│      Gruppenmitglieder: (165mm)     │
│      ┌──────────────────────┐       │
│      │ [Members Table]      │       │
│      │                      │       │
│      └──────────────────────┘       │
│                                     │
│      [Your Decorative Footer]       │
│                                     │
└─────────────────────────────────────┘
```

## No Template Mode

If no template file is found, the certificates will be generated using the built-in programmatic layout with basic styling.

## Tips

- Use semi-transparent or white areas where text will be placed
- Test with a few certificates first to ensure proper text alignment
- The text color is hardcoded (black for main content, purple accents)
- Consider using light backgrounds for better text readability
