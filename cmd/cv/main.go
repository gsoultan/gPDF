// cv demonstrates recreating "CV CONTOH.pdf" — a black-and-white modern CV
// produced with Canva — using gPDF.
//
// Layout:
//   - Left sidebar  (dark background, ~195 pt wide): name, personal data, contact, social media.
//   - Right content (white background, ~371 pt wide): about, education, work experience, skills.
//
// Usage:
//
//	go run ./cmd/cv <output.pdf>
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"

	"gpdf/doc"
	"gpdf/doc/style"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./cmd/cv <output.pdf>")
		fmt.Println("Recreates CV CONTOH.pdf — a black-and-white modern CV — using gPDF.")
		os.Exit(1)
	}
	path := os.Args[1]

	// ── Palette ──────────────────────────────────────────────────────────────
	black := doc.Color{R: 0, G: 0, B: 0}
	white := doc.Color{R: 1, G: 1, B: 1}
	lightGray := doc.Color{R: 0.75, G: 0.75, B: 0.75}
	darkText := doc.Color{R: 0.1, G: 0.1, B: 0.1}

	// ── Layout constants ─────────────────────────────────────────────────────
	const (
		pageW       = 596.0
		pageH       = 842.0
		sidebarW    = 195.0
		sidebarPadX = 15.0
		contentX    = 210.0
		contentW    = 366.0
		contentPadX = contentX
	)

	b := doc.New().
		Title("hitam dan putih modern cv resume riwayat hidup").
		Author("Pedro Fernandes").
		Subject("CV / Resume").
		Creator("gPDF").
		SetLanguage("id-ID").
		PageSize(pageW, pageH).
		AddPage()

	// ═══════════════════════════════════════════════════════════════════════
	// SIDEBAR — black background
	// ═══════════════════════════════════════════════════════════════════════
	b.FillRect(0, 0, 0, sidebarW, pageH, black)

	sidebarFlow := b.Flow(doc.FlowOptions{
		Left:  sidebarPadX,
		Right: pageW - (sidebarW - sidebarPadX),
		Top:   pageH - 800,
	})

	// ── Profile Picture ──────────────────────────────────────────────────────
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 200; x++ {
			img.Set(x, y, color.RGBA{100, 100, 100, 255})
		}
	}
	var buf bytes.Buffer
	jpeg.Encode(&buf, img, nil)
	sidebarFlow.ImageWithLayout(buf.Bytes(), style.ImageLayout{
		Width: 60, Height: 60, Align: style.ImageAlignCenter, Margin: 10,
	}, style.ImageStyle{ClipCircle: true, ClipCX: 30, ClipCY: 30, ClipR: 30})

	// ── Name ────────────────────────────────────────────────────────────────
	sidebarFlow.
		Color(white).Font("Helvetica-Bold").Size(20).
		Paragraph("PEDRO").
		Paragraph("FERNANDES").
		Size(10).Font("Helvetica-Oblique").Color(lightGray).
		Paragraph("Lulusan Baru").
		Space(5).
		Line(0.5, lightGray).
		Space(10)

	// ── DATA DIRI section ────────────────────────────────────────────────────
	sidebarFlow.
		Size(9).Font("Helvetica-Bold").Color(white).Paragraph("DATA DIRI").
		Space(5).
		Size(8).Font("Helvetica").Color(lightGray).
		Paragraph("Tempat / Tanggal Lahir").
		Size(9).Color(white).Paragraph("St., Any City, 12 Juli 1996").
		Space(3).
		Size(8).Color(lightGray).Paragraph("Jenis Kelamin").
		Size(9).Color(white).Paragraph("Laki-Laki").
		Space(3).
		Size(8).Color(lightGray).Paragraph("Status").
		Size(9).Color(white).Paragraph("Belum menikah").
		Space(3).
		Size(8).Color(lightGray).Paragraph("Kewarganegaraan").
		Size(9).Color(white).Paragraph("Indonesia").
		Space(10).
		Line(0.5, lightGray).
		Space(10)

	// ── KONTAK section ───────────────────────────────────────────────────────
	sidebarFlow.
		Size(9).Font("Helvetica-Bold").Color(white).Paragraph("KONTAK").
		Space(5).
		Size(9).Font("Helvetica").Paragraph("+123-456-7890").
		Size(8).Paragraph("hello@reallygreatsite.com").
		Size(9).Paragraph("123 Anywhere St., Any City").
		Space(10).
		Line(0.5, lightGray).
		Space(10)

	// ── SOSIAL MEDIA section ─────────────────────────────────────────────────
	sidebarFlow.
		Size(9).Font("Helvetica-Bold").Paragraph("SOSIAL MEDIA").
		Space(5).
		Size(9).Font("Helvetica").Paragraph("@reallygreatsite")

	// ═══════════════════════════════════════════════════════════════════════
	// RIGHT CONTENT AREA
	// ═══════════════════════════════════════════════════════════════════════

	contentFlow := b.Flow(doc.FlowOptions{
		Left:  contentX,
		Right: pageW - (contentX + contentW),
		Top:   pageH - 800,
	})

	// ── TENTANG SAYA ─────────────────────────────────────────────────────────
	contentFlow.
		Color(darkText).Font("Helvetica-Bold").Size(13).
		Paragraph("TENTANG SAYA").
		Line(1.0, black).
		Space(5).
		Font("Helvetica").Size(10).
		Paragraph("Saya merupakan seorang profesional dalam bidang arsitektur, dan " +
			"telah berpengalaman lebih dari 2 tahun dalam pembangunan infrastruktur. " +
			"Memiliki jiwa pemimpin serta mampu bekerja dalam tim.").
		Space(20)

	// ── PENDIDIKAN ───────────────────────────────────────────────────────────
	contentFlow.
		Font("Helvetica-Bold").Size(13).
		Paragraph("PENDIDIKAN").
		Line(1.0, black).
		Space(10)

	education := []struct {
		school string
		period string
	}{
		{"S1 Arsitektur Universitas Fauget", "2019-2022"},
		{"SMA Negeri 1 Fauget", "2016-2019"},
		{"SMP Negeri 1 Fauget", "2013-2016"},
		{"SD Negeri 1 Fauget", "2008-2013"},
	}
	for _, e := range education {
		contentFlow.
			Font("Helvetica-Bold").Size(10).Paragraph(e.school).
			Font("Helvetica").Size(9).Color(doc.ColorGray).Paragraph(e.period).
			Color(darkText).Space(5)
	}
	contentFlow.Space(15)

	// ── PENGALAMAN KERJA ─────────────────────────────────────────────────────
	contentFlow.
		Font("Helvetica-Bold").Size(13).
		Paragraph("PENGALAMAN KERJA").
		Line(1.0, black).
		Space(10)

	// PT Borcelle
	contentFlow.
		Font("Helvetica-Bold").Size(10).Paragraph("PT Borcelle").
		Space(5).
		Font("Helvetica").Size(9).
		List([]string{
			"Menganalisis data dan memantau kinerja di lapangan",
			"Merancang pembuatan rencana anggaran belanja",
			"Memimpin pelaksanaan pembangunan",
		}, false).
		Space(10)

	// PT Fauget
	contentFlow.
		Font("Helvetica-Bold").Size(10).Paragraph("PT Fauget").
		Space(5).
		Font("Helvetica").Size(9).
		List([]string{
			"Membangun dan merancang gambar gudang penyimpanan",
			"Merancang pembuatan rencana anggaran belanja",
		}, false).
		Space(20)

	// ── KEMAMPUAN ─────────────────────────────────────────────────────────────
	contentFlow.
		Font("Helvetica-Bold").Size(13).
		Paragraph("KEMAMPUAN").
		Line(1.0, black).
		Space(10).
		Font("Helvetica").Size(9).
		List([]string{
			"Aktif berbahasa Indonesia dan Inggris",
			"Mampu menggunakan software desain",
		}, false)

	// ── Profile photo (circular placeholder) ────────────────────────────────
	// The original CV has a circular photo in the upper sidebar.
	// Position derived from the original PDF content stream (scaled coords / 3.126178):
	//   image lower-left: x≈51, y≈590; display size: ≈184×184 pt
	//   clip circle center: cx≈143, cy≈682, radius≈92
	const (
		photoSize = 184.0
		photoX    = 51.0
		photoY    = 590.0
		photoCX   = photoX + photoSize/2
		photoCY   = photoY + photoSize/2
		photoR    = photoSize / 2
	)
	// Build a small gray JPEG placeholder (will be replaced with a real photo).
	const photoRes = 64
	img = image.NewRGBA(image.Rect(0, 0, photoRes, photoRes))
	grayVal := color.RGBA{R: 80, G: 80, B: 80, A: 255}
	for py := range photoRes {
		for px := range photoRes {
			img.Set(px, py, grayVal)
		}
	}
	var jpegBuf bytes.Buffer
	_ = jpeg.Encode(&jpegBuf, img, &jpeg.Options{Quality: 85})
	b = b.DrawCircularJPEG(photoX, photoY, photoSize, photoSize,
		jpegBuf.Bytes(), photoRes, photoRes, "DeviceRGB",
		photoCX, photoCY, photoR)

	// ── Build & save ─────────────────────────────────────────────────────────
	document, err := b.Build()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer document.Close()

	f, err := os.Create(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := document.Save(f); err != nil {
		f.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := f.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Saved CV example: %s\n", path)
}
