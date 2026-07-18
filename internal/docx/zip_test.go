package docx

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestUnpackAndPackRoundTrip(t *testing.T) {
	root := t.TempDir()
	docxPath := filepath.Join(root, "sample.docx")
	writeTestDocx(t, docxPath, map[string]string{
		"[Content_Types].xml": "<Types/>",
		"word/document.xml":   "<w:document/>",
	})

	outDir, err := Unpack(docxPath)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(outDir) != "sample" {
		t.Fatalf("outDir base = %q", filepath.Base(outDir))
	}
	if _, err := os.Stat(filepath.Join(outDir, "word", "document.xml")); err != nil {
		t.Fatal(err)
	}

	packed, err := Pack(outDir)
	if err != nil {
		t.Fatal(err)
	}
	// sample.docx 已存在 → 带时间戳
	if packed == docxPath {
		t.Fatalf("expected unique name, got same path %q", packed)
	}
	if !strings.HasSuffix(packed, ".docx") || !strings.Contains(filepath.Base(packed), "sample-") {
		t.Fatalf("packed name = %q", filepath.Base(packed))
	}
	assertZipHas(t, packed, "[Content_Types].xml", "word/document.xml")
}

func TestUnpackUniqueDirWhenExists(t *testing.T) {
	root := t.TempDir()
	docxPath := filepath.Join(root, "report.docx")
	writeTestDocx(t, docxPath, map[string]string{"a.txt": "1"})
	if err := os.MkdirAll(filepath.Join(root, "report"), 0o755); err != nil {
		t.Fatal(err)
	}
	out, err := Unpack(docxPath)
	if err != nil {
		t.Fatal(err)
	}
	base := filepath.Base(out)
	if !strings.HasPrefix(base, "report-") {
		t.Fatalf("want timestamped dir, got %q", base)
	}
}

func TestUniquePathStamp(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x.docx")
	if err := os.WriteFile(p, []byte("z"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := uniquePath(p, false)
	if got == p || !strings.Contains(got, time.Now().Format("20060102")) {
		t.Fatalf("uniquePath = %q", got)
	}
}

func TestUnpackRejectsZipSlip(t *testing.T) {
	root := t.TempDir()
	docxPath := filepath.Join(root, "evil.docx")
	f, err := os.Create(docxPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("../escape.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.Write([]byte("x"))
	_ = zw.Close()
	_ = f.Close()

	if _, err := Unpack(docxPath); err == nil {
		t.Fatal("expected zip slip error")
	}
}

func writeTestDocx(t *testing.T, path string, files map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func assertZipHas(t *testing.T, path string, names ...string) {
	t.Helper()
	r, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	have := map[string]bool{}
	for _, f := range r.File {
		have[f.Name] = true
	}
	for _, n := range names {
		if !have[n] {
			t.Fatalf("zip missing %q; have %#v", n, have)
		}
	}
}
