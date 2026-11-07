package scanner

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Finding struct {
	ID          string `json:"id"`
	Rule        string `json:"rule"`
	Severity    string `json:"severity"`
	Path        string `json:"path"`
	Line        int    `json:"line"`
	Message     string `json:"message"`
	Fingerprint string `json:"fingerprint"`
}

type secretRule struct {
	name     string
	severity string
	pattern  *regexp.Regexp
}

var secretRules = []secretRule{
	{name: "aws-access-key", severity: "high", pattern: regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`)},
	{name: "github-token", severity: "high", pattern: regexp.MustCompile(`\bgh[pousr]_[A-Za-z0-9_]{20,}\b`)},
	{name: "private-key", severity: "critical", pattern: regexp.MustCompile(`-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----`)},
	{name: "generic-secret-assignment", severity: "medium", pattern: regexp.MustCompile(`(?i)\b(?:password|secret|token|api[_-]?key)\s*[:=]\s*["']?[^\s"']{12,}`)},
}

func ScanSecrets(ctx context.Context, root string) ([]Finding, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve scan root: %w", err)
	}
	var findings []Finding
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "node_modules" || entry.Name() == ".next" {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 || info.Size() > 2<<20 {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 64<<10), 2<<20)
		line := 0
		for scanner.Scan() {
			line++
			text := scanner.Text()
			for _, rule := range secretRules {
				if rule.pattern.MatchString(text) {
					fingerprint := sha256.Sum256([]byte(rule.name + "\x00" + path + "\x00" + fmt.Sprint(line)))
					findings = append(findings, Finding{ID: "finding_" + hex.EncodeToString(fingerprint[:8]), Rule: rule.name, Severity: rule.severity, Path: path, Line: line, Message: "Potential secret detected; value redacted", Fingerprint: hex.EncodeToString(fingerprint[:])})
				}
			}
		}
		return scanner.Err()
	})
	if err != nil {
		return nil, err
	}
	return findings, nil
}

func IsSecretRule(rule string) bool {
	for _, candidate := range secretRules {
		if candidate.name == strings.TrimSpace(rule) {
			return true
		}
	}
	return false
}
