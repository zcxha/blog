package folio

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
)

// ValidateStaticSite rejects malformed output and broken local links before a
// GitHub Pages artifact can be deployed.
func ValidateStaticSite(outDir, basePath string) error {
	base := NormalizeBasePath(basePath)
	var validationErr error
	err := filepath.WalkDir(outDir, func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(filePath), ".html") {
			return nil
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		if !utf8.Valid(data) {
			validationErr = errors.Join(validationErr, fmt.Errorf("%s is not valid UTF-8", filePath))
			return nil
		}
		doc, err := html.Parse(strings.NewReader(string(data)))
		if err != nil {
			validationErr = errors.Join(validationErr, fmt.Errorf("parse %s: %w", filePath, err))
			return nil
		}
		if strings.Contains(string(data), "{{.") || strings.Contains(string(data), "{{template ") {
			validationErr = errors.Join(validationErr, fmt.Errorf("%s contains an unrendered template action", filePath))
		}

		rel, err := filepath.Rel(outDir, filePath)
		if err != nil {
			return err
		}
		pageURL := "/" + filepath.ToSlash(rel)
		if strings.HasSuffix(pageURL, "/index.html") {
			pageURL = strings.TrimSuffix(pageURL, "index.html")
		}
		pageURL = WithBase(base, pageURL)
		validateHTMLNode(doc, func(raw string) {
			if err := validateLocalReference(outDir, base, pageURL, raw); err != nil {
				validationErr = errors.Join(validationErr, fmt.Errorf("%s: %w", filePath, err))
			}
		})
		return nil
	})
	if err != nil {
		return err
	}
	return validationErr
}

func validateHTMLNode(node *html.Node, validate func(string)) {
	if node.Type == html.ElementNode {
		for _, attr := range node.Attr {
			if attr.Key == "href" || attr.Key == "src" || attr.Key == "poster" {
				validate(strings.TrimSpace(attr.Val))
			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		validateHTMLNode(child, validate)
	}
}

func validateLocalReference(outDir, basePath, pageURL, raw string) error {
	if raw == "" || strings.HasPrefix(raw, "#") || strings.HasPrefix(raw, "//") {
		return nil
	}
	ref, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URL %q: %w", raw, err)
	}
	if ref.Scheme != "" || ref.Host != "" {
		return nil
	}
	page, _ := url.Parse(pageURL)
	resolved := page.ResolveReference(ref)
	targetURLPath, err := url.PathUnescape(resolved.Path)
	if err != nil {
		return fmt.Errorf("invalid escaped path %q: %w", raw, err)
	}
	if basePath != "" {
		if targetURLPath != basePath && !strings.HasPrefix(targetURLPath, basePath+"/") {
			return fmt.Errorf("local URL %q escapes configured base path %q", raw, basePath)
		}
		targetURLPath = strings.TrimPrefix(targetURLPath, basePath)
	}
	targetURLPath = strings.TrimPrefix(targetURLPath, "/")
	target := filepath.Join(outDir, filepath.FromSlash(targetURLPath))
	if info, statErr := os.Stat(target); statErr == nil {
		if info.IsDir() {
			target = filepath.Join(target, "index.html")
		}
	} else if filepath.Ext(target) == "" {
		target = filepath.Join(target, "index.html")
	}
	if _, err := os.Stat(target); err != nil {
		return fmt.Errorf("broken local URL %q (expected %s)", raw, target)
	}
	return nil
}
