package docx

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const stampLayout = "20060102-150405"

// Unpack 将 .docx（ZIP）解压到同级目录；目录已存在则使用「原名-时间戳」。
func Unpack(docxPath string) (string, error) {
	docxPath = filepath.Clean(docxPath)
	if !strings.EqualFold(filepath.Ext(docxPath), ".docx") {
		return "", fmt.Errorf("not a .docx file")
	}
	info, err := os.Stat(docxPath)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("not a .docx file")
	}
	base := strings.TrimSuffix(filepath.Base(docxPath), filepath.Ext(docxPath))
	dest := uniquePath(filepath.Join(filepath.Dir(docxPath), base), true)
	if err := unzipTo(docxPath, dest); err != nil {
		_ = os.RemoveAll(dest)
		return "", err
	}
	return dest, nil
}

// Pack 将目录打成同级 .docx（ZIP）；文件已存在则使用「原名-时间戳.docx」。
func Pack(dirPath string) (string, error) {
	dirPath = filepath.Clean(dirPath)
	info, err := os.Stat(dirPath)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("not a directory")
	}
	dest := uniquePath(filepath.Join(filepath.Dir(dirPath), filepath.Base(dirPath)+".docx"), false)
	if err := zipDir(dirPath, dest); err != nil {
		_ = os.Remove(dest)
		return "", err
	}
	return dest, nil
}

func uniquePath(path string, isDir bool) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}
	stamp := time.Now().Format(stampLayout)
	var candidate string
	if isDir {
		candidate = path + "-" + stamp
	} else {
		ext := filepath.Ext(path)
		candidate = strings.TrimSuffix(path, ext) + "-" + stamp + ext
	}
	if _, err := os.Stat(candidate); os.IsNotExist(err) {
		return candidate
	}
	for i := 2; i < 1000; i++ {
		next := fmt.Sprintf("%s-%d", candidate, i)
		if _, err := os.Stat(next); os.IsNotExist(err) {
			return next
		}
	}
	return candidate
}

func unzipTo(zipPath, dest string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	destAbs, err := filepath.Abs(dest)
	if err != nil {
		return err
	}
	for _, f := range r.File {
		if err := extractZipFile(f, destAbs); err != nil {
			return err
		}
	}
	return nil
}

func extractZipFile(f *zip.File, destAbs string) error {
	name := filepath.Clean(f.Name)
	if name == "." || strings.HasPrefix(name, "..") {
		return fmt.Errorf("invalid zip entry: %s", f.Name)
	}
	target := filepath.Join(destAbs, name)
	rel, err := filepath.Rel(destAbs, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("zip slip: %s", f.Name)
	}
	if f.FileInfo().IsDir() {
		return os.MkdirAll(target, 0o755)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, rc)
	return err
}

func zipDir(srcDir, destZip string) error {
	out, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer out.Close()
	zw := zip.NewWriter(out)
	defer zw.Close()
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		name := filepath.ToSlash(rel)
		if info.IsDir() {
			_, err := zw.Create(name + "/")
			return err
		}
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		return err
	})
}
