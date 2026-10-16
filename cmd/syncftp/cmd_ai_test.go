package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolveCmdShim(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows only")
	}
	dir := t.TempDir()
	exe := filepath.Join(dir, "node_modules", "tool", "bin", "tool.exe")
	os.MkdirAll(filepath.Dir(exe), 0755)
	os.WriteFile(exe, []byte("x"), 0755)
	shim := filepath.Join(dir, "tool.cmd")
	os.WriteFile(shim, []byte("@ECHO off\r\n"+`"%dp0%\node_modules\tool\bin\tool.exe"   %*`+"\r\n"), 0644)
	if bin, pre := resolveCmdShim(shim); bin != exe || len(pre) != 0 {
		t.Fatalf("exe shim: %s %v", bin, pre)
	}

	js := filepath.Join(dir, "node_modules", "cli", "dist", "index.js")
	os.MkdirAll(filepath.Dir(js), 0755)
	os.WriteFile(js, []byte("x"), 0644)
	os.WriteFile(filepath.Join(dir, "node.exe"), []byte("x"), 0755)
	shim2 := filepath.Join(dir, "cli.cmd")
	os.WriteFile(shim2, []byte(`"%_prog%"  "%dp0%\node_modules\cli\dist\index.js" %*`+"\r\n"), 0644)
	if bin, pre := resolveCmdShim(shim2); bin != filepath.Join(dir, "node.exe") || len(pre) != 1 || pre[0] != js {
		t.Fatalf("js shim: %s %v", bin, pre)
	}
	if bin, _ := resolveCmdShim(exe); bin != exe {
		t.Fatal("non-shim must pass through")
	}
}
