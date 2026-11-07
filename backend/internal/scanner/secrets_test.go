package scanner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanSecretsFindsMetadataWithoutSecretValue(t *testing.T) {
	root := t.TempDir()
	value := "AKIA1234567890ABCDEF"
	path := filepath.Join(root, "config.txt")
	if err := os.WriteFile(path, []byte("aws_key="+value+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	findings, err := ScanSecrets(context.Background(), root)
	if err != nil || len(findings) != 1 {
		t.Fatalf("unexpected findings: %+v, %v", findings, err)
	}
	if strings.Contains(findings[0].Message, value) || strings.Contains(findings[0].Message, "AKIA") {
		t.Fatalf("secret value leaked into finding: %+v", findings[0])
	}
	if findings[0].Severity != "high" || findings[0].Line != 1 {
		t.Fatalf("unexpected finding metadata: %+v", findings[0])
	}
}

func TestScanSecretsSkipsDevelopmentDirectories(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "node_modules"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "node_modules", "bad.txt"), []byte("token=ghp_123456789012345678901234567890"), 0o600); err != nil {
		t.Fatal(err)
	}
	findings, err := ScanSecrets(context.Background(), root)
	if err != nil || len(findings) != 0 {
		t.Fatalf("unexpected findings: %+v, %v", findings, err)
	}
}
