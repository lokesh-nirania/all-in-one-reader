package reader

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetUniqueFilePath_NoCollision(t *testing.T) {
	dir := t.TempDir()
	orig := filepath.Join(dir, "file.txt")

	got, err := GetUniqueFilePath(orig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != orig {
		t.Fatalf("expected %q, got %q", orig, got)
	}
}

func TestGetUniqueFilePath_WithCollisionOnce(t *testing.T) {
	dir := t.TempDir()
	orig := filepath.Join(dir, "file.txt")

	if err := os.WriteFile(orig, []byte("x"), 0o644); err != nil {
		t.Fatalf("precreate file: %v", err)
	}

	got, err := GetUniqueFilePath(orig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(got) != "file_1.txt" {
		t.Fatalf("expected basename file_1.txt, got %q", filepath.Base(got))
	}
}

func TestGetUniqueFilePath_WithTwoCollisions(t *testing.T) {
	dir := t.TempDir()
	orig := filepath.Join(dir, "file.txt")
	collide1 := filepath.Join(dir, "file_1.txt")

	if err := os.WriteFile(orig, []byte("x"), 0o644); err != nil {
		t.Fatalf("precreate file: %v", err)
	}
	if err := os.WriteFile(collide1, []byte("y"), 0o644); err != nil {
		t.Fatalf("precreate file_1: %v", err)
	}

	got, err := GetUniqueFilePath(orig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(got) != "file_2.txt" {
		t.Fatalf("expected basename file_2.txt, got %q", filepath.Base(got))
	}
}

func TestGetUniqueFilePath_NoExtension(t *testing.T) {
	dir := t.TempDir()
	orig := filepath.Join(dir, "file")

	if err := os.WriteFile(orig, []byte("x"), 0o644); err != nil {
		t.Fatalf("precreate file: %v", err)
	}

	got, err := GetUniqueFilePath(orig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(got) != "file_1" {
		t.Fatalf("expected basename file_1, got %q", filepath.Base(got))
	}
}

func TestHumanizeReadableSize(t *testing.T) {
	type tc struct {
		bytes    int64
		wantVal  float64
		wantUnit string
	}
	tests := []tc{
		{bytes: 500, wantVal: 500, wantUnit: "bytes"},
		{bytes: 1024, wantVal: 1, wantUnit: "KB"},
		{bytes: 1536, wantVal: 1.5, wantUnit: "KB"},
		{bytes: 1 << 20, wantVal: 1, wantUnit: "MB"},
		{bytes: 5 << 20, wantVal: 5, wantUnit: "MB"},
		{bytes: 1 << 30, wantVal: 1, wantUnit: "GB"},
		{bytes: 1 << 40, wantVal: 1, wantUnit: "TB"},
	}

	for _, tt := range tests {
		gotVal, gotUnit := HumanizeReadableSize(tt.bytes)
		if gotUnit != tt.wantUnit {
			t.Fatalf("bytes=%d: unit: want %q got %q", tt.bytes, tt.wantUnit, gotUnit)
		}
		if gotVal != tt.wantVal {
			t.Fatalf("bytes=%d: val: want %v got %v", tt.bytes, tt.wantVal, gotVal)
		}
	}
}

func captureStdout(f func()) string {
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	_ = w.Close()
	os.Stdout = orig
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()
	return buf.String()
}

func TestNotifyProgress_WithTotalSize(t *testing.T) {
	out := captureStdout(func() {
		NotifyProgress(512, 1024)
	})
	if !strings.Contains(out, "Downloaded") {
		t.Fatalf("expected output to contain 'Downloaded', got %q", out)
	}
	if !strings.Contains(out, "1.00 KB") {
		t.Fatalf("expected output to contain '1.00 KB', got %q", out)
	}
	if !strings.Contains(out, "(50.00%)") {
		t.Fatalf("expected output to contain '(50.00%%)', got %q", out)
	}
}

func TestNotifyProgress_WithoutTotalSize_Bytes(t *testing.T) {
	out := captureStdout(func() {
		NotifyProgress(5, 0)
	})
	if !strings.Contains(out, "Downloaded 5 bytes...") {
		t.Fatalf("expected 'Downloaded 5 bytes...' in output, got %q", out)
	}
}

func TestNotifyProgress_WithoutTotalSize_KB(t *testing.T) {
	out := captureStdout(func() {
		NotifyProgress(2048, 0)
	})
	if !strings.Contains(out, "Downloaded 2.00 KB") {
		t.Fatalf("expected 'Downloaded 2.00 KB' in output, got %q", out)
	}
}
