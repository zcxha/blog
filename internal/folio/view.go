package folio

import (
	"fmt"
	"html/template"
	"net/url"
	"strings"
)

type IndexPageData struct {
	Title           string
	BasePath        string
	AuthorGitHub    string
	StylePath       string
	FaviconPath     string
	SiteDescription string
	SEO             SEO
	Posts           []Post
	Pagination      Pagination
	Analytics       AnalyticsConfig
}

type PostPageData struct {
	Title        string
	BasePath     string
	AuthorGitHub string
	StylePath    string
	FaviconPath  string
	SEO          SEO
	Post         Post
	Comments     CommentConfig
	Analytics    AnalyticsConfig
}

type TagsPageData struct {
	Title        string
	BasePath     string
	AuthorGitHub string
	StylePath    string
	FaviconPath  string
	SEO          SEO
	CurrentTag   string
	Tags         []TagStat
	Posts        []Post
	Pagination   Pagination
	Analytics    AnalyticsConfig
}

type ArchivesPageData struct {
	Title        string
	BasePath     string
	AuthorGitHub string
	StylePath    string
	FaviconPath  string
	SEO          SEO
	Groups       []ArchiveGroup
	Pagination   Pagination
	Analytics    AnalyticsConfig
}

type SearchPageData struct {
	Title        string
	BasePath     string
	AuthorGitHub string
	StylePath    string
	FaviconPath  string
	SEO          SEO
	Analytics    AnalyticsConfig
}

type NotFoundPageData struct {
	Title        string
	BasePath     string
	AuthorGitHub string
	StylePath    string
	FaviconPath  string
	SEO          SEO
	Message      string
	Analytics    AnalyticsConfig
}

type PageLink struct {
	Number  int
	URL     string
	Current bool
}

type Pagination struct {
	CurrentPage int
	TotalPages  int
	PrevURL     string
	NextURL     string
	Pages       []PageLink
}

type CommentConfig struct {
	Enabled        bool
	Provider       string
	Repo           string
	RepoID         string
	Category       string
	CategoryID     string
	Mapping        string
	Theme          string
	Lang           string
	Label          string
	IssueTerm      string
	DiscussionTerm string
}

type AnalyticsConfig struct {
	Enabled   bool
	Provider  string
	Endpoint  string
	PublicURL string
}

func ParseTemplate(theme, pageRel string, tagResolver func(string) string) (*template.Template, error) {
	funcMap := template.FuncMap{
		"tagURL": tagResolver,
		"postURL": func(basePath, slug string) string {
			return WithBase(basePath, "/post/"+url.PathEscape(strings.TrimSpace(slug)))
		},
	}
	head := ResolveTemplatePath(theme, "partials/head-common.html")
	nav := ResolveTemplatePath(theme, "partials/nav.html")
	page := ResolveTemplatePath(theme, pageRel)
	if head == "" || nav == "" || page == "" {
		return nil, fmt.Errorf("theme templates not found: theme=%s page=%s", theme, pageRel)
	}
	files := []string{head, nav, page}
	return template.New(pageRel).Funcs(funcMap).ParseFiles(files...)
}
