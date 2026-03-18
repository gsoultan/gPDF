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
		PageSize(pageW, pageH)
	b.AddPage()

	// ═══════════════════════════════════════════════════════════════════════
	// SIDEBAR — black background
	// ═══════════════════════════════════════════════════════════════════════
	b = b.FillRect(0, 0, 0, sidebarW, pageH, black)

	// ── Name ────────────────────────────────────────────────────────────────
	b = b.
		DrawTextColored("PEDRO", sidebarPadX, 800, "Helvetica-Bold", 20, white).
		DrawTextColored("FERNANDES", sidebarPadX, 776, "Helvetica-Bold", 20, white).
		DrawTextColored("Lulusan Baru", sidebarPadX, 754, "Helvetica-Oblique", 10, lightGray)

	// ── Divider ──────────────────────────────────────────────────────────────
	b = b.DrawLine(0, sidebarPadX, 741, sidebarW-sidebarPadX, 741,
		doc.LineStyle{Width: 0.5, Color: lightGray})

	// ── DATA DIRI section ────────────────────────────────────────────────────
	b = b.
		DrawTextColored("DATA DIRI", sidebarPadX, 726, "Helvetica-Bold", 9, white).
		DrawTextColored("Tempat / Tanggal Lahir", sidebarPadX, 708, "Helvetica", 8, lightGray).
		DrawTextColored("St., Any City, 12 Juli 1996", sidebarPadX, 695, "Helvetica", 9, white).
		DrawTextColored("Jenis Kelamin", sidebarPadX, 677, "Helvetica", 8, lightGray).
		DrawTextColored("Laki-Laki", sidebarPadX, 664, "Helvetica", 9, white).
		DrawTextColored("Status", sidebarPadX, 646, "Helvetica", 8, lightGray).
		DrawTextColored("Belum menikah", sidebarPadX, 633, "Helvetica", 9, white).
		DrawTextColored("Kewarganegaraan", sidebarPadX, 615, "Helvetica", 8, lightGray).
		DrawTextColored("Indonesia", sidebarPadX, 602, "Helvetica", 9, white)

	// ── Divider ──────────────────────────────────────────────────────────────
	b = b.DrawLine(0, sidebarPadX, 589, sidebarW-sidebarPadX, 589,
		doc.LineStyle{Width: 0.5, Color: lightGray})

	// ── KONTAK section ───────────────────────────────────────────────────────
	b = b.
		DrawTextColored("KONTAK", sidebarPadX, 574, "Helvetica-Bold", 9, white).
		DrawTextColored("+123-456-7890", sidebarPadX, 558, "Helvetica", 9, white).
		DrawTextColored("hello@reallygreatsite.com", sidebarPadX, 543, "Helvetica", 8, white).
		DrawTextColored("123 Anywhere St., Any City", sidebarPadX, 528, "Helvetica", 9, white)

	// ── Divider ──────────────────────────────────────────────────────────────
	b = b.DrawLine(0, sidebarPadX, 515, sidebarW-sidebarPadX, 515,
		doc.LineStyle{Width: 0.5, Color: lightGray})

	// ── SOSIAL MEDIA section ─────────────────────────────────────────────────
	b = b.
		DrawTextColored("SOSIAL MEDIA", sidebarPadX, 500, "Helvetica-Bold", 9, white).
		DrawTextColored("@reallygreatsite", sidebarPadX, 484, "Helvetica", 9, white)

	// ═══════════════════════════════════════════════════════════════════════
	// RIGHT CONTENT AREA
	// ═══════════════════════════════════════════════════════════════════════

	// ── TENTANG SAYA ─────────────────────────────────────────────────────────
	b = b.
		DrawTextColored("TENTANG SAYA", contentPadX, 800, "Helvetica-Bold", 13, darkText)
	b = b.DrawLine(0, contentPadX, 784, contentPadX+contentW, 784,
		doc.LineStyle{Width: 1.0, Color: black})
	b = b.DrawTextBoxColored(0,
		"Saya merupakan seorang profesional dalam bidang arsitektur, dan "+
			"telah berpengalaman lebih dari 2 tahun dalam pembangunan infrastruktur. "+
			"Memiliki jiwa pemimpin serta mampu bekerja dalam tim.",
		contentPadX, 769, "Helvetica", 10,
		doc.TextLayoutOptions{Width: contentW, LineHeight: 14},
		darkText,
	)

	// ── PENDIDIKAN ───────────────────────────────────────────────────────────
	b = b.
		DrawTextColored("PENDIDIKAN", contentPadX, 710, "Helvetica-Bold", 13, darkText)
	b = b.DrawLine(0, contentPadX, 694, contentPadX+contentW, 694,
		doc.LineStyle{Width: 1.0, Color: black})

	type eduEntry struct {
		school string
		period string
		y      float64
	}
	education := []eduEntry{
		{"S1 Arsitektur Universitas Fauget", "2019-2022", 679},
		{"SMA Negeri 1 Fauget", "2016-2019", 659},
		{"SMP Negeri 1 Fauget", "2013-2016", 639},
		{"SD Negeri 1 Fauget", "2008-2013", 619},
	}
	for _, e := range education {
		b = b.
			DrawTextColored(e.school, contentPadX, e.y, "Helvetica-Bold", 10, darkText).
			DrawTextColored(e.period, contentPadX, e.y-13, "Helvetica", 9, doc.ColorGray)
	}

	// ── PENGALAMAN KERJA ─────────────────────────────────────────────────────
	b = b.
		DrawTextColored("PENGALAMAN KERJA", contentPadX, 578, "Helvetica-Bold", 13, darkText)
	b = b.DrawLine(0, contentPadX, 562, contentPadX+contentW, 562,
		doc.LineStyle{Width: 1.0, Color: black})

	// PT Borcelle
	b = b.
		DrawTextColored("PT Borcelle", contentPadX, 547, "Helvetica-Bold", 10, darkText)
	ptBorcelleItems := []string{
		"Menganalisis data dan memantau kinerja di lapangan",
		"Merancang pembuatan rencana anggaran belanja",
		"Memimpin pelaksanaan pembangunan",
	}
	y := 530.0
	for _, item := range ptBorcelleItems {
		b = b.
			DrawTextColored("-", contentPadX, y, "Helvetica", 9, darkText).
			DrawTextColored(item, contentPadX+10, y, "Helvetica", 9, darkText)
		y -= 14
	}

	// PT Fauget
	b = b.
		DrawTextColored("PT Fauget", contentPadX, y-4, "Helvetica-Bold", 10, darkText)
	y -= 20
	ptFaugetItems := []string{
		"Membangun dan merancang gambar gudang penyimpanan",
		"Merancang pembuatan rencana anggaran belanja",
	}
	for _, item := range ptFaugetItems {
		b = b.
			DrawTextColored("-", contentPadX, y, "Helvetica", 9, darkText).
			DrawTextColored(item, contentPadX+10, y, "Helvetica", 9, darkText)
		y -= 14
	}

	// ── KEMAMPUAN ─────────────────────────────────────────────────────────────
	kemampuanY := y - 20
	b = b.
		DrawTextColored("KEMAMPUAN", contentPadX, kemampuanY, "Helvetica-Bold", 13, darkText)
	b = b.DrawLine(0, contentPadX, kemampuanY-16, contentPadX+contentW, kemampuanY-16,
		doc.LineStyle{Width: 1.0, Color: black})

	skills := []string{
		"Aktif berbahasa Indonesia dan Inggris",
		"Mampu menggunakan software desain",
	}
	skillY := kemampuanY - 31
	for _, skill := range skills {
		b = b.
			DrawTextColored("-", contentPadX, skillY, "Helvetica", 9, darkText).
			DrawTextColored(skill, contentPadX+10, skillY, "Helvetica", 9, darkText)
		skillY -= 14
	}

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
	img := image.NewRGBA(image.Rect(0, 0, photoRes, photoRes))
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
