package test

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	core "folio/internal/folio"
)

func waitHTTP(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return context.DeadlineExceeded
}

func freeTCPPort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("allocate free port failed: %v", err)
	}
	defer func() { _ = l.Close() }()

	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("unexpected addr type: %T", l.Addr())
	}
	return strconv.Itoa(addr.Port)
}

func firstImageRequestPath(t *testing.T, root string) string {
	t.Helper()
	imagesRoot := filepath.Join(root, "posts", "images")
	var rel string
	err := filepath.WalkDir(imagesRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		found, relErr := filepath.Rel(imagesRoot, path)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(found)
		return io.EOF
	})
	if err != nil && err != io.EOF {
		t.Fatalf("walk images failed: %v", err)
	}
	if rel == "" {
		t.Fatal("expected at least one post image")
	}

	segments := strings.Split(rel, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	return "/images/" + strings.Join(segments, "/")
}

func TestServerMainRoutesSmoke(t *testing.T) {
	root := repoRoot(t)
	port := freeTCPPort(t)
	posts, err := core.LoadPosts(filepath.Join(root, "posts"), "test")
	if err != nil {
		t.Fatalf("load posts failed: %v", err)
	}
	if len(posts) == 0 {
		t.Fatal("expected at least one visible post")
	}
	samplePostPath := "/post/" + posts[0].Slug
	sampleImagePath := firstImageRequestPath(t, root)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", ".")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "PORT="+port)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server failed: %v", err)
	}
	defer func() {
		cancel()
		_ = cmd.Wait()
	}()

	base := "http://127.0.0.1:" + port
	if err := waitHTTP(base+"/", 12*time.Second); err != nil {
		t.Fatalf("server not ready in time: %v", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	cases := []struct {
		path       string
		statusCode int
	}{
		{"/", http.StatusOK},
		{"/tags", http.StatusOK},
		{"/archives", http.StatusOK},
		{"/search", http.StatusOK},
		{"/search-index.json", http.StatusOK},
		{samplePostPath, http.StatusOK},
		{"/post/hello_folio", http.StatusNotFound},
		{"/post/not-found", http.StatusNotFound},
		{"/tags?tag=%3Cscript%3E", http.StatusNotFound},
		{"/static/not-found.css", http.StatusNotFound},
		{"/static/../config.json", http.StatusNotFound},
		{"/static/style.css", http.StatusOK},
		{"/static/vendor/mathjax/tex-svg.js", http.StatusOK},
		{sampleImagePath, http.StatusOK},
		{"/not-found-page", http.StatusNotFound},
	}
	for _, tc := range cases {
		resp, err := client.Get(base + tc.path)
		if err != nil {
			t.Fatalf("request %s failed: %v", tc.path, err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode != tc.statusCode {
			t.Fatalf("unexpected status for %s: got %d want %d", tc.path, resp.StatusCode, tc.statusCode)
		}
	}
}

func TestSearchIndexReturnsJSON(t *testing.T) {
	root := repoRoot(t)
	port := freeTCPPort(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", ".")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "PORT="+port)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server failed: %v", err)
	}
	defer func() {
		cancel()
		_ = cmd.Wait()
	}()

	base := "http://127.0.0.1:" + port
	if err := waitHTTP(base+"/search-index.json", 12*time.Second); err != nil {
		t.Fatalf("server not ready in time: %v", err)
	}

	resp, err := http.Get(base + "/search-index.json")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body failed: %v", err)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("unexpected content-type: %s", ct)
	}
	if !strings.Contains(string(b), `"slug"`) {
		t.Fatalf("unexpected payload: %s", string(b))
	}
}
