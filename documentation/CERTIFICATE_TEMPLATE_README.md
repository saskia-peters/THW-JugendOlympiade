# Certificate Template Usage

The application generates two types of PDF certificates. Both support optional custom background images.

---

## 1. Participant Certificates (`Urkunden_Teilnehmende.pdf`)

One certificate per participant. Generated via **📊 Ausgabe → „Urkunden Teilnehmende"**.

Two styles are available, controlled by `urkunden_stil` in `config.toml`:

| Setting | Effect |
|---------|--------|
| `urkunden_stil = "text"` | Members table (default) — lists all group members below the rank |
| `urkunden_stil = "picture"` | Group photo — embeds a photo of the group instead of the members table |

### Style `picture` — Group Photos

Set `urkunden_stil = "picture"` in `config.toml`. Place JPEG photos in the configured picture directory (default: `pictures/`, set via `bilder_ordner`):

```
pictures/
├── group_picture_001.jpg   ← Group 1
├── group_picture_002.jpg   ← Group 2
└── ...
```

The `pictures/` directory is created automatically on first launch. If a photo file is missing, a grey placeholder rectangle is shown instead.

### Optional Background Image

Place a file named `certificate_template.png` in the application directory:

```
certificate_template.png
```

If the file is absent, a built-in programmatic layout is used instead.

### Template Specifications

- **Format**: PNG
- **Size**: A4 portrait — 210 mm × 297 mm (2480 × 3508 px at 300 DPI)

### Design Guidelines – leave space for dynamic content:

| Content | Approx. Y position |
|---------|-------------------|
| Event year | ~35 mm from top |
| Participant name | ~85 mm from top |
| Ortsverband | ~105 mm from top |
| Group number ("Gruppe X") | ~125 mm from top |
| Rank ("1. Platz" etc.) | ~140 mm from top |
| Group members table | starts ~165 mm from top |

### Adjusting Text Positions

Edit Y-coordinates in `backend/io/pdf_cert_teilnehmende.go`:

```go
pdf.SetXY(0, 85)  // second number = Y position in mm
```

### Example Layout

```
┌─────────────────────────────────────┐
│       [Decorative Header/Logo]      │
│           2026  (35mm)              │
│                                     │
│      [Participant Name] (85mm)      │
│      [Ortsverband]      (105mm)     │
│      Gruppe X           (125mm)     │
│      1. Platz           (140mm)     │
│                                     │
│      Gruppenmitglieder: (165mm)     │
│      ┌──────────────────────┐       │
│      │ [Members Table]      │       │
│      └──────────────────────┘       │
│       [Decorative Footer]           │
└─────────────────────────────────────┘
```

---

## 2. Ortsverband Certificates (`Urkunden_Ortsverbaende.pdf`)

One certificate page per Ortsverband. Generated via **📊 Ausgabe → „Urkunden Ortsverbände"**.

- The **best Ortsverband(e)** (highest average score; all tied Ortsverbände) receive a **Siegerurkunde** with trophy image and "Bester Ortsverband" heading.
- All other Ortsverbände receive a standard participation certificate.
- The event name used on each certificate is taken from `config.toml` (`veranstaltung.name`).

### Optional Files

| File | Purpose |
|------|---------|
| `cert_background_ov.png` | Background image for OV certificates (A4 PNG, same specs as above). Gracefully skipped if missing or invalid. |
| `ov_winner_image.png` | Trophy/winner image displayed on the Siegerurkunde. Skipped if missing. |

### Dynamic Content on OV Certificates

- Veranstaltungsname (from `config.toml`)
- "Bester Ortsverband" heading (winner only)
- Trophy image at 140 mm width (winner only)
- "Teilnehmende" list (plain, no table) for each OV
- OV name

---

## Tips

- Use semi-transparent or white areas where text will be overlaid.
- Test with a small dataset (2–3 groups) first to check alignment.
- Text color is hardcoded (black / purple accents for participant certificates).
- Convert PDF or JPG templates to PNG using an online tool (e.g. pdf2png.com) or Adobe Acrobat.


## How to Use a Custom Template

### 1. Create Your Template Design

Create your certificate template using any design tool (Canva, Photoshop, PowerPoint, etc.) with the following specifications:

- **Size**: A4 (210mm x 297mm) or 2480 x 3508 pixels at 300 DPI
- **Format**: PNG
- **Orientation**: Portrait

### 2. Design Guidelines

Leave space for the dynamic content that will be overlaid:

- **Year** (Top center, ~35mm from top): the event year from `config.toml` (`veranstaltung.jahr`), in large text
- **Participant Name** (Center, ~85mm from top): Large, prominent
- **Ortsverband** (~105mm from top): Smaller text below name
- **Group Number** (~125mm from top): "Gruppe X"
- **Rank** (~140mm from top): "1. Platz", "2. Platz", etc.
- **Group Members Table** (Starting ~165mm from top): Table with participant details

### 3. Save the Template

Save your template as the following file in the application directory:

```
certificate_template.png
```

**Note:** If you have a PDF or JPG template, convert it to PNG first using:
- An online converter (e.g., pdf2png.com)
- Adobe Acrobat (Export as PNG)
- Any PDF reader’s “Save as Image” feature

### 4. Generate Certificates

1. Run the application
2. Load your Excel file  
3. Click "Teilnehmer-Zertifikate"
4. The certificates will be generated with your template as the background

## Adjusting Text Positions

If you need to adjust where the text appears on your template, you can modify the Y-coordinates in the code:

File: `backend/io/pdf_cert_teilnehmende.go`  
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
