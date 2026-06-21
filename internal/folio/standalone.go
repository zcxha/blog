package folio

import (
	"html"
	"regexp"
	"strings"
)

var (
	reClosingHead = regexp.MustCompile(`(?i)</head\s*>`)
	reClosingBody = regexp.MustCompile(`(?i)</body\s*>`)
)

// PrepareStandalonePostDocument keeps exported HTML documents intact while
// adding the same math, analytics, likes, navigation, and comments used by
// Markdown posts.
func PrepareStandalonePostDocument(post Post, basePath string, cfg AppConfig) string {
	doc := PreparePostDocumentForRender(post, basePath)
	includeMath := !strings.Contains(strings.ToLower(doc), "<mjx-container") && !strings.Contains(strings.ToLower(doc), "mathjax.js")
	head := standaloneHead(basePath, BuildAnalyticsConfig(cfg), includeMath)
	body := standaloneFooter(post, basePath, BuildCommentConfig(cfg, post), BuildAnalyticsConfig(cfg))

	if reClosingHead.MatchString(doc) {
		doc = reClosingHead.ReplaceAllStringFunc(doc, func(string) string { return head + "\n</head>" })
	} else {
		doc = "<!doctype html><html lang=\"zh-CN\"><head><meta charset=\"utf-8\">" + head + "</head><body>" + doc + "</body></html>"
	}
	if reClosingBody.MatchString(doc) {
		doc = reClosingBody.ReplaceAllStringFunc(doc, func(string) string { return body + "\n</body>" })
	} else {
		doc += body
	}
	return doc
}

func standaloneHead(basePath string, analytics AnalyticsConfig, includeMath bool) string {
	var out strings.Builder
	out.WriteString(`<style id="folio-integrations">
.folio-extras{box-sizing:border-box;max-width:860px;margin:48px auto 24px;padding:20px;border-top:1px solid #d8d2c5;font:14px/1.6 system-ui,sans-serif;color:#555}
.folio-extras__row{display:flex;align-items:center;flex-wrap:wrap;gap:12px}.folio-extras a{color:inherit}.folio-extras button{border:1px solid #b9b2a4;border-radius:999px;background:#fff;padding:8px 13px;cursor:pointer}.folio-extras button.is-liked{color:#b92f54;border-color:#d55b78}.folio-extras__comments{margin-top:24px}
</style>`)
	if includeMath {
		out.WriteString(`<script>window.MathJax={tex:{tags:"ams",inlineMath:[["$","$"],["\\(","\\)"]],displayMath:[["$$","$$"],["\\[","\\]"]]}};</script>`)
		out.WriteString(`<script defer src="` + html.EscapeString(WithBase(basePath, "/static/vendor/mathjax/tex-svg.js")) + `"></script>`)
	}
	out.WriteString(`<script defer src="` + html.EscapeString(WithBase(basePath, "/static/site.js")) + `"></script>`)
	if analytics.Enabled {
		out.WriteString(`<script data-goatcounter="` + html.EscapeString(analytics.Endpoint) + `" async src="https://gc.zgo.at/count.js"></script>`)
	}
	return out.String()
}

func standaloneFooter(post Post, basePath string, comments CommentConfig, analytics AnalyticsConfig) string {
	var out strings.Builder
	out.WriteString(`<aside class="folio-extras" aria-label="博客互动"><div class="folio-extras__row">`)
	out.WriteString(`<a href="` + html.EscapeString(WithBase(basePath, "/")) + `">← 返回博客</a>`)
	if analytics.Enabled {
		out.WriteString(`<button type="button" data-like-button data-like-path="` + html.EscapeString(post.Slug) + `" aria-pressed="false"><span data-like-icon aria-hidden="true">♡</span> <span data-like-label>喜欢这篇文章</span></button>`)
		out.WriteString(`<span>本文浏览 <strong data-page-view-count></strong></span><span>点赞 <strong data-like-count data-like-path="` + html.EscapeString(post.Slug) + `"></strong></span><span>全站浏览 <strong data-site-view-count></strong></span>`)
	}
	if analytics.PublicURL != "" {
		out.WriteString(`<a href="` + html.EscapeString(analytics.PublicURL) + `" target="_blank" rel="noopener noreferrer">查看统计</a>`)
	}
	out.WriteString(`</div>`)
	if comments.Enabled {
		out.WriteString(`<div class="folio-extras__comments">` + commentEmbed(comments) + `</div>`)
	}
	out.WriteString(`</aside>`)
	return out.String()
}

func commentEmbed(c CommentConfig) string {
	attr := func(value string) string { return html.EscapeString(value) }
	if c.Provider == "giscus" {
		return `<script src="https://giscus.app/client.js" data-repo="` + attr(c.Repo) + `" data-repo-id="` + attr(c.RepoID) +
			`" data-category="` + attr(c.Category) + `" data-category-id="` + attr(c.CategoryID) + `" data-mapping="` + attr(c.Mapping) +
			`" data-strict="0" data-reactions-enabled="1" data-emit-metadata="0" data-input-position="top" data-theme="` + attr(c.Theme) +
			`" data-lang="` + attr(c.Lang) + `" data-loading="lazy" crossorigin="anonymous" async></script>`
	}
	if c.Provider == "utterances" {
		label := ""
		if c.Label != "" {
			label = ` label="` + attr(c.Label) + `"`
		}
		return `<script src="https://utteranc.es/client.js" repo="` + attr(c.Repo) + `" issue-term="` + attr(c.IssueTerm) + `"` + label +
			` theme="` + attr(c.Theme) + `" crossorigin="anonymous" async></script>`
	}
	return ""
}
