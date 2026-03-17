# Jugendolympiade Verwaltung

A cross-platform desktop application for managing youth Olympics events. Built with [Wails v2](https://wails.io/) (Go backend + Web frontend), this application handles participant registration, group distribution, station scoring, evaluations, and certificate generation.

![Version](https://img.shields.io/badge/version-1.0.0-blue)
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey)

## Features

### ��� Participant Management
- **Excel Import**: Import participant data from XLSX files with automatic validation
- **Smart Groups**: Automatically creates balanced groups with at most 8 participants
- **Database Storage**: All data stored securely in SQLite database

### ��� Station Scoring
- **Track Performance**: Record scores for each group at different stations
- **Import Stations**: Load station data from Excel files
- **Real-time Updates**: Scores are saved automatically

### ��� Evaluations
- **Group Rankings**: View rankings by group with total scores
- **Ortsverband Rankings**: Compare locations/districts by average scores
- **Statistics**: Participant counts, score distributions, and averages

### ��� PDF Generation
All PDFs are automatically saved to the \`pdfdocs/\` directory:
- **Groups Report**: One page per group with participant lists and statistics
- **Group Evaluations**: Rankings by group with scores
- **Ortsverband Evaluations**: Rankings by location
- **Participant Certificates**: Individual certificates for all participants
  - Supports custom certificate templates
  - Shows participant details, group assignment, and ranking
  - Lists all group members

### ���️ Desktop Application
- **Cross-Platform**: Runs on Windows, macOS, and Linux
- **Modern GUI**: Clean, intuitive interface
- **Fast Performance**: Native Go backend
- **Native File Dialogs**: OS-integrated file picker

## Installation

### Download
Download the latest release for your platform:
- **Windows**: \`THW-JugendOlympiade.exe\`
- **macOS**: \`THW-JugendOlympiade.app\`
- **Linux**: \`THW-JugendOlympiade\`

### First Launch
1. Double-click the executable to launch
2. On Windows, ensure [WebView2](https://developer.microsoft.com/microsoft-edge/webview2/) is installed
3. On macOS, you may need to allow the app in System Preferences → Security & Privacy

## Usage

### 1. Load Participant Data

**Prepare Your Excel File:**
- Create an XLSX file with a sheet named "Teilnehmer"
- Required columns (in order):
  1. **Name**: Participant name
  2. **Ortsverband**: Location/district
  3. **Alter**: Age (must be between 1-100)
  4. **Geschlecht**: Gender
- First row is treated as header and skipped

**Import:**
1. Click "Load Excel File"
2. Select your XLSX file
3. Wait for confirmation message
4. Groups are automatically created

### 2. View Groups

- Click "Gruppen" to view all created groups
- Groups are automatically balanced by:
  - Location (Ortsverband)
  - Age (Alter)
  - Gender (Geschlecht)
- Maximum 8 participants per group

### 3. Add Stations

Station names are loaded automatically from the **second sheet** (`Stationen`) of your Excel file during import. If the sheet is present, all station names are stored in the database and will be available for score entry.

### 4. Enter Scores

1. Click "Ergebniseingabe" to open the results entry view
2. Select a group from the dropdown
3. Enter scores (100–1200) for each station
4. Click "Speichern" per row or "Alle Ergebnisse speichern" to save all at once
5. Switch to the next group and repeat

### 5. View Evaluations

**Group Rankings:**
- Click "Auswertung nach Gruppen"
- View total scores and rankings by group

**Ortsverband Rankings:**
- Click "Auswertung nach Ortsverband"
- View average scores per location/district

### 6. Generate PDFs

**Group Reports:**
- Click "Gruppen-PDF erstellen"
- Generates detailed report in `pdfdocs/Gruppeneinteilung.pdf`

**Certificates:**
- Click "Teilnehmer-Zertifikate"
- Generates certificates in `pdfdocs/Urkunden_Teilnehmende.pdf`

All PDFs are saved to the \`pdfdocs/\` directory.

## Certificate Templates

### Using Custom Templates

Create professional-looking certificates with custom designs:

1. **Create Template File:**
   - Design your certificate in A4 size (210mm × 297mm)
   - Save as `certificate_template.png`
   - Place in the same directory as the application

2. **Template Specifications:**
   - **Size**: A4 (210mm × 297mm)
   - **Format**: PNG
   - **Resolution**: 2480×3508 pixels at 300 DPI
   - **Important**: Leave space for dynamic content between x-coordinates 23px (5mm) and 680px (147.83mm)

3. **Dynamic Content:**
   The following information is automatically overlaid on your template:
   - Participant name
   - Ortsverband (location)
   - Group number
   - Group ranking (1st, 2nd, 3rd place, etc.)
   - List of all group members

See [CERTIFICATE_TEMPLATE_README.md](CERTIFICATE_TEMPLATE_README.md) for detailed template guidelines.

## Output Files

After using the application, you'll find:

### Database
- **data.db**: SQLite database with all data
  - Participant information
  - Group assignments
  - Station scores
  - Evaluations

### PDFs (in pdfdocs/ directory)
- **Gruppeneinteilung.pdf**: Complete group listings with statistics
- **Auswertung_nach_Gruppe.pdf**: Group rankings by total score
- **Auswertung_nach_Ortsverband.pdf**: Location rankings by average score
- **Urkunden_Teilnehmende.pdf**: Individual certificates for all participants

## Troubleshooting

### Common Issues

**"failed to initialize database"**
- Ensure you have write permissions in the application directory
- Close any other programs that might be accessing \`data.db\`

**"invalid file format"**
- Check that your Excel file has a sheet named "Teilnehmer"
- Verify column headers: Name, Ortsverband, Alter, Geschlecht
- Ensure file is \`.xlsx\` format (not \`.xls\` or \`.csv\`)

**"age must be between 1 and 100"**
- Check for invalid age values in your Excel file
- Ensure the Alter column contains only numbers
- Remove any empty rows or non-numeric values

**PDFs not generating**
- Ensure \`pdfdocs/\` directory can be created
- Close any PDF files that might be open
- Check available disk space

**Application won't start (Windows)**
- Install [Microsoft Edge WebView2](https://developer.microsoft.com/microsoft-edge/webview2/)
- If still failing, try running as administrator

**Application won't start (macOS)**
- Right-click the app → Open (first time only)
- Go to System Preferences → Security & Privacy → Allow the app

**Application won't start (Linux)**
- Make the file executable: `chmod +x THW-JugendOlympiade`
- Install required libraries: \`sudo apt-get install libgtk-3-0 libwebkit2gtk-4.0-37\`

### Getting Help

1. Check this README for solutions
2. Review error messages carefully
3. Verify your input data format
4. Try with a smaller test dataset first

## System Requirements

- **Windows**: Windows 10/11 with WebView2
- **macOS**: macOS 10.13 or later
- **Linux**: Modern distribution with GTK3 and WebKit2GTK

**Disk Space**: ~50MB for application, plus space for database and PDFs

**Memory**: 256MB minimum, 512MB recommended

## License

[Add your license here]

## Credits

Built with:
- [Wails](https://wails.io/) - Desktop application framework
- [Go](https://golang.org/) - Backend language
- [excelize](https://github.com/qax-os/excelize) - Excel processing
- [gofpdf](https://github.com/jung-kurt/gofpdf) - PDF generation

---

**For Developers**: See [README_DEVELOPER.md](README_DEVELOPER.md) for technical documentation, architecture details, and contribution guidelines.
