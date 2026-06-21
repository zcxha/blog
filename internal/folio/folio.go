package folio

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"html/template"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
)

type Post struct {
	Slug        string
	Title       string
	Author      string
	Date        time.Time
	DateDisplay string
	Tags        []string
	Category    string
	Draft       bool
	Format      string
	Content     string
	Markdown    string
	RawHTML     string
	HTML        template.HTML
}

type SearchDoc struct {
	Title    string   `json:"title"`
	Slug     string   `json:"slug"`
	Date     string   `json:"date"`
	Tags     []string `json:"tags"`
	Category string   `json:"category,omitempty"`
	Content  string   `json:"content"`
}

type SEO struct {
	Description   string
	CanonicalURL  string
	OGType        string
	OGURL         string
	OGImage       string
	SiteName      string
	PublishedTime string
}

type AppConfig struct {
	SiteTitle          string `json:"site_title"`
	SiteDescription    string `json:"site_description"`
	SiteURL            string `json:"site_url"`
	AuthorName         string `json:"author_name"`
	AuthorGitHub       string `json:"author_github"`
	Theme              string `json:"theme"`
	DefaultDescription string `json:"default_description"`
	DefaultOGImage     string `json:"default_og_image"`
	DefaultOGType      string `json:"default_og_type"`
	CommentsProvider   string `json:"comments_provider"`
	CommentsRepo       string `json:"comments_repo"`
	CommentsRepoID     string `json:"comments_repo_id"`
	CommentsCategory   string `json:"comments_category"`
	CommentsCategoryID string `json:"comments_category_id"`
	CommentsMapping    string `json:"comments_mapping"`
	CommentsTheme      string `json:"comments_theme"`
	CommentsLang       string `json:"comments_lang"`
	CommentsLabel      string `json:"comments_label"`
	CommentsIssueTerm  string `json:"comments_issue_term"`
	AnalyticsProvider  string `json:"analytics_provider"`
	AnalyticsEndpoint  string `json:"analytics_endpoint"`
	AnalyticsPublicURL string `json:"analytics_public_url"`
}

type TagStat struct {
	Name  string
	Count int
	URL   string
}

type ArchiveGroup struct {
	Label string
	Posts []Post
}

var (
	reHTMLBody      = regexp.MustCompile(`(?is)<body\b[^>]*>(.*)</body>`)
	reHTMLTitle     = regexp.MustCompile(`(?is)<title\b[^>]*>(.*?)</title>`)
	reHTMLH1        = regexp.MustCompile(`(?is)<h1\b[^>]*>(.*?)</h1>`)
	reHTMLScript    = regexp.MustCompile(`(?is)<script\b[^>]*>.*?</script>`)
	reHTMLStyle     = regexp.MustCompile(`(?is)<style\b[^>]*>.*?</style>`)
	reHTMLComment   = regexp.MustCompile(`(?is)<!--.*?-->`)
	reHTMLTag       = regexp.MustCompile(`(?is)<[^>]+>`)
	reHTMLAssetAttr = regexp.MustCompile(`(?i)\b(src|href|poster)=(['"])(?:\./)?images/([^'"]+)(['"])`)
)

var supportedPostExtensions = []string{".md", ".html"}

func DefaultConfig() AppConfig {
	return AppConfig{
		SiteTitle:          "Folio",
		SiteDescription:    "A lightweight blog powered by Go and file storage.",
		SiteURL:            "",
		AuthorName:         "Anonymous",
		AuthorGitHub:       "",
		Theme:              "default",
		DefaultDescription: "A lightweight blog powered by Go and file storage.",
		DefaultOGImage:     "",
		DefaultOGType:      "website",
		CommentsMapping:    "pathname",
		CommentsTheme:      "preferred_color_scheme",
		CommentsLang:       "zh-CN",
		CommentsIssueTerm:  "pathname",
	}
}

func (c *AppConfig) normalize() {
	d := DefaultConfig()
	if strings.TrimSpace(c.SiteTitle) == "" {
		c.SiteTitle = d.SiteTitle
	}
	if strings.TrimSpace(c.SiteDescription) == "" {
		c.SiteDescription = d.SiteDescription
	}
	c.SiteURL = strings.TrimRight(strings.TrimSpace(c.SiteURL), "/")
	if strings.TrimSpace(c.AuthorName) == "" {
		c.AuthorName = d.AuthorName
	}
	c.AuthorGitHub = strings.TrimSpace(c.AuthorGitHub)
	if c.AuthorGitHub != "" &&
		!strings.HasPrefix(c.AuthorGitHub, "http://") &&
		!strings.HasPrefix(c.AuthorGitHub, "https://") {
		c.AuthorGitHub = "https://github.com/" + strings.Trim(c.AuthorGitHub, "/")
	}
	c.Theme = NormalizeThemeName(c.Theme)
	if strings.TrimSpace(c.DefaultDescription) == "" {
		c.DefaultDescription = c.SiteDescription
	}
	if strings.TrimSpace(c.DefaultOGType) == "" {
		c.DefaultOGType = d.DefaultOGType
	}
	c.CommentsProvider = strings.ToLower(strings.TrimSpace(c.CommentsProvider))
	if strings.TrimSpace(c.CommentsMapping) == "" {
		c.CommentsMapping = d.CommentsMapping
	}
	if strings.TrimSpace(c.CommentsTheme) == "" {
		c.CommentsTheme = d.CommentsTheme
	}
	if strings.TrimSpace(c.CommentsLang) == "" {
		c.CommentsLang = d.CommentsLang
	}
	if strings.TrimSpace(c.CommentsIssueTerm) == "" {
		c.CommentsIssueTerm = d.CommentsIssueTerm
	}
	c.AnalyticsProvider = strings.ToLower(strings.TrimSpace(c.AnalyticsProvider))
	c.AnalyticsEndpoint = strings.TrimSpace(c.AnalyticsEndpoint)
	c.AnalyticsPublicURL = strings.TrimSpace(c.AnalyticsPublicURL)
}

func LoadConfig(path string) (AppConfig, error) {
	cfg := DefaultConfig()
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}
	if len(strings.TrimSpace(string(b))) == 0 {
		return cfg, nil
	}
	b = bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF})
	if err := json.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}
	cfg.normalize()
	return cfg, nil
}

func CanonicalURL(siteURL, path string) string {
	site := strings.TrimRight(strings.TrimSpace(siteURL), "/")
	if site == "" {
		return path
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Avoid duplicated base segment, e.g. site=/folio + path=/folio/post/x.
	if u, err := url.Parse(site); err == nil {
		base := strings.TrimRight(u.Path, "/")
		if base != "" {
			if path == base {
				path = "/"
			} else if strings.HasPrefix(path, base+"/") {
				path = strings.TrimPrefix(path, base)
			}
		}
	}
	return site + path
}

func MakeSEO(cfg AppConfig, title, desc, path, ogType, publishedTime string) SEO {
	desc = strings.TrimSpace(desc)
	if desc == "" {
		desc = cfg.DefaultDescription
	}
	ogType = strings.TrimSpace(ogType)
	if ogType == "" {
		ogType = cfg.DefaultOGType
	}
	url := CanonicalURL(cfg.SiteURL, path)
	return SEO{
		Description:   desc,
		CanonicalURL:  url,
		OGType:        ogType,
		OGURL:         url,
		OGImage:       cfg.DefaultOGImage,
		SiteName:      cfg.SiteTitle,
		PublishedTime: publishedTime,
	}
}

func WithBase(basePath, path string) string {
	if basePath == "" {
		return path
	}
	base := "/" + strings.Trim(basePath, "/")
	if strings.HasPrefix(path, "/") {
		return base + path
	}
	return base + "/" + path
}

func NormalizeBasePath(basePath string) string {
	if strings.TrimSpace(basePath) == "" {
		return ""
	}
	return "/" + strings.Trim(basePath, "/")
}

func BuildTagURL(basePath, tag, mode string) string {
	if mode == "static" {
		return WithBase(basePath, "/tags/"+SlugifyTag(tag)+"/")
	}
	return WithBase(basePath, "/tags?tag="+url.QueryEscape(tag))
}

func BuildTagStats(posts []Post, basePath, mode string) []TagStat {
	counts := map[string]int{}
	for _, post := range posts {
		for _, tag := range postTaxonomy(post) {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			counts[tag]++
		}
	}

	stats := make([]TagStat, 0, len(counts))
	for name, count := range counts {
		stats = append(stats, TagStat{Name: name, Count: count, URL: BuildTagURL(basePath, name, mode)})
	}
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Count == stats[j].Count {
			return stats[i].Name < stats[j].Name
		}
		return stats[i].Count > stats[j].Count
	})
	return stats
}

func BuildStaticTagStats(posts []Post, basePath string) ([]TagStat, map[string]string, map[string]string) {
	counts := map[string]int{}
	for _, post := range posts {
		for _, tag := range postTaxonomy(post) {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			counts[tag]++
		}
	}

	names := make([]string, 0, len(counts))
	for name := range counts {
		names = append(names, name)
	}
	sort.Strings(names)

	used := map[string]bool{}
	tagSlugs := map[string]string{}
	tagURLs := map[string]string{}
	stats := make([]TagStat, 0, len(names))

	for _, name := range names {
		baseSlug := SlugifyTag(name)
		slug := baseSlug
		i := 2
		for used[slug] {
			slug = fmt.Sprintf("%s-%d", baseSlug, i)
			i++
		}
		used[slug] = true
		tagSlugs[name] = slug
		tagURLs[name] = WithBase(basePath, "/tags/"+slug+"/")
		stats = append(stats, TagStat{Name: name, Count: counts[name], URL: tagURLs[name]})
	}

	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Count == stats[j].Count {
			return stats[i].Name < stats[j].Name
		}
		return stats[i].Count > stats[j].Count
	})
	return stats, tagSlugs, tagURLs
}

func FilterPostsByTag(posts []Post, target string) []Post {
	target = strings.TrimSpace(target)
	if target == "" {
		return posts
	}
	out := make([]Post, 0)
	for _, post := range posts {
		for _, tag := range postTaxonomy(post) {
			if strings.EqualFold(strings.TrimSpace(tag), target) {
				out = append(out, post)
				break
			}
		}
	}
	return out
}

func postTaxonomy(post Post) []string {
	values := make([]string, 0, len(post.Tags)+1)
	if category := strings.TrimSpace(post.Category); category != "" {
		values = append(values, category)
	}
	for _, tag := range post.Tags {
		tag = strings.TrimSpace(tag)
		if tag == "" || (len(values) > 0 && strings.EqualFold(values[0], tag)) {
			continue
		}
		values = append(values, tag)
	}
	return values
}

func BuildArchiveGroups(posts []Post) []ArchiveGroup {
	keys := make([]string, 0)
	groups := map[string][]Post{}
	for _, post := range posts {
		key := post.Date.Format("2006-01")
		if _, ok := groups[key]; !ok {
			keys = append(keys, key)
		}
		groups[key] = append(groups[key], post)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	out := make([]ArchiveGroup, 0, len(keys))
	for _, key := range keys {
		out = append(out, ArchiveGroup{Label: key, Posts: groups[key]})
	}
	return out
}

func BuildCommentConfig(cfg AppConfig, post Post) CommentConfig {
	provider := strings.ToLower(strings.TrimSpace(cfg.CommentsProvider))
	if provider == "" {
		return CommentConfig{}
	}

	switch provider {
	case "giscus":
		repo := strings.TrimSpace(cfg.CommentsRepo)
		repoID := strings.TrimSpace(cfg.CommentsRepoID)
		category := strings.TrimSpace(cfg.CommentsCategory)
		categoryID := strings.TrimSpace(cfg.CommentsCategoryID)
		if repo == "" || repoID == "" || category == "" || categoryID == "" {
			return CommentConfig{}
		}
		return CommentConfig{
			Enabled:    true,
			Provider:   "giscus",
			Repo:       repo,
			RepoID:     repoID,
			Category:   category,
			CategoryID: categoryID,
			Mapping:    strings.TrimSpace(cfg.CommentsMapping),
			Theme:      strings.TrimSpace(cfg.CommentsTheme),
			Lang:       strings.TrimSpace(cfg.CommentsLang),
		}
	case "utterances":
		repo := strings.TrimSpace(cfg.CommentsRepo)
		if repo == "" {
			return CommentConfig{}
		}
		issueTerm := strings.TrimSpace(cfg.CommentsIssueTerm)
		if issueTerm == "slug" {
			issueTerm = post.Slug
		}
		return CommentConfig{
			Enabled:   true,
			Provider:  "utterances",
			Repo:      repo,
			IssueTerm: issueTerm,
			Label:     strings.TrimSpace(cfg.CommentsLabel),
			Theme:     strings.TrimSpace(cfg.CommentsTheme),
		}
	default:
		return CommentConfig{}
	}
}

func BuildAnalyticsConfig(cfg AppConfig) AnalyticsConfig {
	provider := strings.ToLower(strings.TrimSpace(cfg.AnalyticsProvider))
	endpoint := strings.TrimSpace(cfg.AnalyticsEndpoint)
	if provider != "goatcounter" || endpoint == "" {
		return AnalyticsConfig{}
	}
	if parsed, err := url.Parse(endpoint); err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return AnalyticsConfig{}
	}
	return AnalyticsConfig{
		Enabled:   true,
		Provider:  provider,
		Endpoint:  endpoint,
		PublicURL: strings.TrimSpace(cfg.AnalyticsPublicURL),
	}
}

func PaginatePosts(posts []Post, page, perPage int) ([]Post, int, int) {
	if perPage <= 0 {
		perPage = 10
	}
	totalPages := 1
	if len(posts) > 0 {
		totalPages = (len(posts) + perPage - 1) / perPage
	}
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}
	start := (page - 1) * perPage
	end := start + perPage
	if start >= len(posts) {
		return []Post{}, totalPages, page
	}
	if end > len(posts) {
		end = len(posts)
	}
	return posts[start:end], totalPages, page
}

func MakeSearchDocs(posts []Post) []SearchDoc {
	docs := make([]SearchDoc, 0, len(posts))
	for _, post := range posts {
		content := post.Content
		if strings.TrimSpace(content) == "" {
			content = post.Markdown
		}
		docs = append(docs, SearchDoc{
			Title:    post.Title,
			Slug:     post.Slug,
			Date:     post.DateDisplay,
			Tags:     post.Tags,
			Category: post.Category,
			Content:  NormalizeSearchText(content),
		})
	}
	return docs
}

func LoadPosts(dir, fallbackAuthor string) ([]Post, error) {
	files, err := collectPostPaths(dir)
	if err != nil {
		return nil, err
	}

	posts := make([]Post, 0, len(files))
	seenSlugs := make(map[string]string, len(files))
	var loadErr error
	for _, path := range files {
		slug := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		if previous, exists := seenSlugs[slug]; exists {
			loadErr = errors.Join(loadErr, fmt.Errorf("duplicate post slug %q: %s and %s", slug, previous, path))
			continue
		}
		seenSlugs[slug] = path

		post, err := LoadPost(path, fallbackAuthor)
		if err != nil {
			loadErr = errors.Join(loadErr, fmt.Errorf("%s: %w", path, err))
			continue
		}
		if post.Draft {
			continue
		}
		posts = append(posts, post)
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})
	if loadErr != nil {
		return nil, loadErr
	}
	return posts, nil
}

func LoadPostBySlug(dir, slug, fallbackAuthor string) (Post, error) {
	path, err := resolvePostPathBySlug(dir, slug)
	if err != nil {
		return Post{}, err
	}
	return LoadPost(path, fallbackAuthor)
}

func LoadPost(path, fallbackAuthor string) (Post, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Post{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return Post{}, err
	}

	slug := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	fm, body := splitFrontMatter(string(b))
	body = strings.TrimSpace(body)
	format := detectPostFormat(path)
	title := strings.TrimSpace(fm["title"])
	if title == "" && format == "html" {
		title = detectHTMLTitle(body)
	}

	post := Post{
		Slug:     slug,
		Title:    fallbackTitle(title, slug),
		Author:   fallbackAuthorName(fm["author"], fallbackAuthor),
		Tags:     parseList(fm["tags"]),
		Category: strings.TrimSpace(fm["category"]),
		Draft:    strings.EqualFold(strings.TrimSpace(fm["draft"]), "true"),
		Format:   format,
	}

	post.Date = parseDateOrFallback(fm["date"], info.ModTime())
	post.DateDisplay = post.Date.Format("2006-01-02")

	switch format {
	case "html":
		post.RawHTML = body
		post.Content = htmlToText(body)
		post.HTML = template.HTML(extractHTMLContent(body))
	default:
		post.Markdown = body
		post.Content = body
		post.HTML = template.HTML(renderMarkdownGoldmark(post.Markdown))
	}
	return post, nil
}

func Excerpt(s string, n int) string {
	text := NormalizeSearchText(s)
	runes := []rune(text)
	if len(runes) <= n {
		return text
	}
	return string(runes[:n]) + "..."
}

func NormalizeSearchText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	replacer := strings.NewReplacer(
		"#", " ", "*", " ", "`", " ", ">", " ", "[", " ", "]", " ",
		"(", " ", ")", " ", "-", " ", "_", " ", "\n", " ", "\t", " ",
	)
	s = replacer.Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

func collectPostPaths(dir string) ([]string, error) {
	files := make([]string, 0, len(supportedPostExtensions))
	for _, ext := range supportedPostExtensions {
		matches, err := filepath.Glob(filepath.Join(dir, "*"+ext))
		if err != nil {
			return nil, err
		}
		files = append(files, matches...)
	}
	sort.Strings(files)
	return files, nil
}

func resolvePostPathBySlug(dir, slug string) (string, error) {
	files, err := collectPostPaths(dir)
	if err != nil {
		return "", err
	}
	for _, path := range files {
		if strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)) == slug {
			return path, nil
		}
	}
	return "", os.ErrNotExist
}

func detectPostFormat(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".html":
		return "html"
	default:
		return "markdown"
	}
}

func extractHTMLContent(input string) string {
	input = strings.TrimSpace(strings.ReplaceAll(input, "\r\n", "\n"))
	if input == "" {
		return ""
	}
	if match := reHTMLBody.FindStringSubmatch(input); len(match) == 2 {
		return strings.TrimSpace(match[1])
	}
	return input
}

func detectHTMLTitle(input string) string {
	for _, re := range []*regexp.Regexp{reHTMLTitle, reHTMLH1} {
		if match := re.FindStringSubmatch(input); len(match) == 2 {
			if title := htmlToText(match[1]); title != "" {
				return title
			}
		}
	}
	return ""
}

func htmlToText(input string) string {
	text := extractHTMLContent(input)
	text = reHTMLComment.ReplaceAllString(text, " ")
	text = reHTMLScript.ReplaceAllString(text, " ")
	text = reHTMLStyle.ReplaceAllString(text, " ")
	text = reHTMLTag.ReplaceAllString(text, " ")
	text = html.UnescapeString(text)
	text = strings.ReplaceAll(text, "\u00a0", " ")
	return strings.Join(strings.Fields(text), " ")
}

func PreparePostForRender(post Post, basePath string) Post {
	if post.HTML == "" {
		return post
	}
	post.HTML = template.HTML(rewriteHTMLAssetURLs(string(post.HTML), basePath))
	return post
}

func PreparePostDocumentForRender(post Post, basePath string) string {
	if post.Format != "html" {
		return ""
	}
	return rewriteHTMLAssetURLs(post.RawHTML, basePath)
}

func rewriteHTMLAssetURLs(input, basePath string) string {
	return reHTMLAssetAttr.ReplaceAllStringFunc(input, func(match string) string {
		parts := reHTMLAssetAttr.FindStringSubmatch(match)
		if len(parts) != 5 || parts[2] != parts[4] {
			return match
		}

		assetPath, ok := buildPostAssetURL(basePath, parts[3])
		if !ok {
			return match
		}
		return fmt.Sprintf(`%s=%s%s%s`, parts[1], parts[2], assetPath, parts[2])
	})
}

func buildPostAssetURL(basePath, rel string) (string, bool) {
	if decoded, err := url.PathUnescape(rel); err == nil {
		rel = decoded
	}
	cleaned := path.Clean("/images/" + strings.TrimSpace(rel))
	if !strings.HasPrefix(cleaned, "/images/") {
		return "", false
	}
	return WithBase(basePath, encodeURLPath(cleaned)), true
}

func encodeURLPath(raw string) string {
	if raw == "" {
		return ""
	}
	parts := strings.Split(raw, "/")
	for i, part := range parts {
		if i == 0 && part == "" {
			continue
		}
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func SlugifyTag(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "tag"
	}
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		isAlphaNum := unicode.IsLetter(r) || unicode.IsNumber(r)
		if isAlphaNum {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash && b.Len() > 0 {
			b.WriteByte('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "tag"
	}
	return out
}

func NormalizeThemeName(theme string) string {
	theme = strings.ToLower(strings.TrimSpace(theme))
	if theme == "" {
		return "default"
	}
	for _, r := range theme {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_'
		if !ok {
			return "default"
		}
	}
	return theme
}

func ResolveTemplatePath(theme, rel string) string {
	rel = strings.TrimLeft(strings.ReplaceAll(rel, "\\", "/"), "/")
	t := NormalizeThemeName(theme)
	candidates := []string{
		filepath.Join("themes", t, "templates", filepath.FromSlash(rel)),
		filepath.Join("themes", "default", "templates", filepath.FromSlash(rel)),
	}
	for _, p := range candidates {
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}
	return ""
}

func ResolveStaticPath(theme, rel string) string {
	rel = strings.TrimLeft(strings.ReplaceAll(rel, "\\", "/"), "/")
	t := NormalizeThemeName(theme)
	candidates := []string{
		filepath.Join("themes", t, "static", filepath.FromSlash(rel)),
		filepath.Join("themes", "default", "static", filepath.FromSlash(rel)),
	}
	for _, p := range candidates {
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}
	return ""
}

func splitFrontMatter(content string) (map[string]string, string) {
	content = strings.TrimPrefix(content, "\uFEFF")
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(strings.TrimPrefix(lines[0], "\uFEFF")) != "---" {
		return map[string]string{}, content
	}

	meta := map[string]string{}
	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
		k, v, ok := strings.Cut(lines[i], ":")
		if !ok {
			continue
		}
		meta[strings.TrimSpace(strings.ToLower(k))] = strings.Trim(strings.TrimSpace(v), "\"")
	}
	if end == -1 {
		return map[string]string{}, content
	}
	return meta, strings.Join(lines[end+1:], "\n")
}

func parseList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	raw = strings.TrimPrefix(raw, "[")
	raw = strings.TrimSuffix(raw, "]")
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.Trim(strings.TrimSpace(p), "\"")
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseDateOrFallback(raw string, fallback time.Time) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}

	layouts := []string{time.RFC3339, "2006-01-02", "2006-01-02 15:04:05"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t
		}
	}
	return fallback
}

func fallbackTitle(title, slug string) string {
	title = strings.TrimSpace(title)
	if title != "" {
		return title
	}
	parts := strings.Split(strings.ReplaceAll(slug, "-", " "), " ")
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		runes := []rune(parts[i])
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}
	return strings.Join(parts, " ")
}

func fallbackAuthorName(author, fallback string) string {
	author = strings.TrimSpace(author)
	if author != "" {
		return author
	}
	return strings.TrimSpace(fallback)
}
