# Folio

一个基于 Go + 文件系统的静态博客生成器。保留纯 GitHub Pages 部署，同时把内容渲染和站点集成拆成了可测试的独立层。

核心能力：

- Markdown：Goldmark CommonMark + GFM，支持表格、任务列表、脚注、删除线、自动链接和原生 HTML。
- LaTeX：`$...$` / `$$...$$` / `\(...\)` / `\[...\]`，公式内容不会再被 Markdown 的强调语法破坏；MathJax SVG 运行时随站点发布，不依赖公共 CDN。
- HTML：支持片段和完整的 Typora 导出文档；完整文档保持原样，只注入导航、数学、评论与统计集成。
- 构建校验：每次导出都会解析所有 HTML，并拒绝 UTF-8 错误、未渲染模板以及失效的站内链接/图片。
- 发布自动化：`publish` 在测试和构建前调用 Codex，为本次变更的文章更新分类与标签。
- 互动：Utterances / Giscus（GitHub 登录评论），以及 GoatCounter 的文章浏览、全站浏览和匿名点赞事件。

演示地址：<https://wofiporia.github.io/folio/>

## 分支与发布

- `main`：模板分支（不自动发布）
- `blog`：内容分支（自动发布到 GitHub Pages）

推荐流程：使用 `Use this template` 创建仓库，然后在 `blog` 分支写作并发布。

## 快速开始

1. 使用模板创建你的仓库。
2. 新建并切换到 `blog` 分支。
3. 在仓库 `Settings -> Pages` 中把 `Source` 设为 `GitHub Actions`。
4. 在 `Settings -> Environments -> github-pages` 里允许 `blog` 分支部署。
5. push 到 `blog`，等待 Actions 完成。

## 配置文件（`config.json`）

示例：

```jsonc
{
  "site_title": "Folio",
  "site_description": "一个基于 Go 和文件系统的轻量博客。",
  "site_url": "https://your-name.github.io/your-repo",
  "author_name": "Your Name",
  "author_github": "your-github-id",
  "theme": "default",
  "default_description": "这里什么都没有写。",
  "default_og_image": "",
  "default_og_type": "website",

  "comments_provider": "utterances",
  "comments_repo": "owner/repo",
  "comments_issue_term": "pathname",
  "comments_label": "comment",
  "comments_theme": "github-light",

  "analytics_provider": "goatcounter",
  "analytics_endpoint": "https://your-code.goatcounter.com/count",
  "analytics_public_url": "https://your-code.goatcounter.com"
}
```

主要字段：

- `site_title`：站点名。
- `site_description`：站点描述（首页与默认 SEO 描述）。
- `site_url`：站点完整 URL（用于 canonical / og:url）。
- `author_name`：文章未写作者时的默认作者名。
- `author_github`：作者 GitHub 地址（支持直接写用户名，程序会自动补全为 `https://github.com/<name>`）。
- `theme`：主题名（对应 `themes/<theme>`）。
- `default_description`：缺省 SEO 描述。
- `default_og_image`：缺省 OG 图片 URL。
- `default_og_type`：缺省 OG 类型。

评论字段：

- `comments_provider`：`utterances` 或 `giscus`。
- 若配置不完整，评论区会自动关闭，不影响页面渲染。
- Utterances 使用 GitHub Issues，访客通过 GitHub 授权后评论；首次使用需为仓库安装 <https://github.com/apps/utterances>。
- Giscus 使用 GitHub Discussions，并同时提供 reaction 点赞。启用 Discussions 后，在 <https://giscus.app/zh-CN> 生成 `repo_id`、`category`、`category_id` 并填入配置即可。

统计字段：

- `analytics_provider`：当前支持 `goatcounter`。
- `analytics_endpoint`：GoatCounter 的 `/count` 地址；留空时安全关闭统计与独立点赞按钮。
- `analytics_public_url`：可选。填写后导航栏出现“统计”入口。
- 普通页面路径就是文章浏览统计，`TOTAL` 显示全站浏览；点赞以 `like:<文章路径>` 自定义事件写入，浏览器本地会阻止同一访客重复点击。若要在页面上展示数字，请在 GoatCounter 设置中允许公开 visitor counts。

GitHub Pages 本身没有可写数据库，因此评论、全局浏览量和全局点赞数必须落在 GitHub 或统计服务中。这里没有把密钥放进前端，也不会在服务不可用时阻断正文渲染。

## 主题

当前内置主题：

- `default`：简约风。
- `kinetic`：更大胆的视觉与排版。

主题目录约定：

```text
themes/
└── <theme>/
    ├── templates/
    │   ├── index.html
    │   ├── post.html
    │   ├── tags.html
    │   ├── archives.html
    │   ├── search.html
    │   ├── 404.html
    │   └── partials/
    │       ├── head-common.html
    │       └── nav.html
    └── static/
        ├── style.css
        └── favicon.png
```

说明：模板和静态资源都支持自动回退到 `themes/default`。

## 写作

文章支持两种文件格式和一套 LaTeX 数学语法，都会进入首页、标签页、归档页、搜索索引和静态构建：

- `posts/*.md`：Markdown 模式
- `posts/*.html`：HTML 模式
- Markdown 内的 `$...$`、`$$...$$`、`\(...\)`、`\[...\]`：LaTeX 数学模式

两种格式都支持可选 Front Matter：

```markdown
---
title: "我的第一篇文章"
author: "Your Name"
date: "2026-03-03T10:00:00Z"
tags: ["博客", "Go"]
category: "开发工具"
draft: false
---
```

- `author` 可选；不填时回退到 `config.json` 的 `author_name`。
- `date` 可选；不填或解析失败时回退到文章文件的修改时间。
- `draft: true` 的文章不会出现在前台。
- `category` 和 `tags` 会展示并写入搜索索引；通常由发布前的 Codex 分类器维护，也可以手工修改。
- 若是 `.html` 文件且没写 `title`，程序会尝试从 HTML 的 `<title>` 或第一个 `<h1>` 自动提取标题。
- `.html` 文件会按原始 HTML 文档直接输出，不再套博客正文排版；`.md` 文件会继续按 Markdown 渲染。
- HTML 文章里若使用 `./images/...`，程序会自动映射到站点的文章图片目录；静态构建时 `posts/images` 也会一起发布。
- 不要同时创建同名的 `posts/foo.md` 和 `posts/foo.html`，因为它们会映射到同一个文章链接 `/post/foo`。

## 页面与功能

- 首页：`/`
- 文章页：`/post/{slug}`
- 标签页：`/tags`
- 归档页：`/archives`
- 搜索页：`/search`（前端读取 `search-index.json`）
- SEO：`description`、Open Graph、`canonical`、`article:published_time`

## 本地开发

启动本地服务：

```bash
go run .
```

访问：`http://localhost:8080`

导出静态站点：

```bash
go run ./cmd/build -out dist -base-path /your-repo-name
```

可选参数：

- `-config`：指定配置文件路径（默认 `config.json`）
- `-site-url`：导出时覆盖站点 URL

## 一键发布

仓库根目录新增了：

- `publish.cmd`：Windows 下可直接双击
- `scripts/publish.ps1`：PowerShell 发布脚本

脚本会默认执行：

1. 找出本次新增/修改的博文，通过 `codex exec` 生成结构化分类与标签，并安全更新 Front Matter
2. `go test ./...`
3. `go run ./cmd/build` 做一次本地构建与站内链接检查
4. `git add -A`
5. 自动提交
6. `git push origin HEAD:blog`

也就是说，写完文章后直接双击 `publish.cmd`，或者在终端运行：

```powershell
.\publish.cmd
```

可选参数示例：

```powershell
.\publish.cmd -Message "post: 更新 scan 与并行思考"
.\publish.cmd -SkipTests
.\publish.cmd -SkipBuild
.\publish.cmd -SkipAI
```

说明：

- 默认推送到远端 `blog` 分支，因为 GitHub Pages 工作流当前监听的是 `blog`。
- 如果当前本地分支不是 `blog`，脚本也会把当前 `HEAD` 推到远端 `blog`，所以使用前确认这正是你想发布的内容。
- AI 分类要求已安装并登录 Codex CLI；失败会中止发布。确实需要人工标签时可显式使用 `-SkipAI`。
- 可单独运行 `powershell -File scripts/classify-posts.ps1 -All` 为全部历史文章重新分类。

## 开发检查

安装并确认 Go 版本：

```bash
go version
```

运行开发检查：

```bash
make test
```

检查包含 `gofmt`、`go vet`、单元/集成测试和完整静态构建验证，不依赖额外的全局 lint 工具。

## GitHub Pages Base Path

可在仓库变量中设置 `PAGES_BASE_PATH`（`Settings -> Secrets and variables -> Actions -> Variables`）。

常见值：

- 项目页：`/your-repo-name`
- 用户主页根路径：`/`
- 自定义子路径：`/blog`

优先级：

1. `PAGES_BASE_PATH`
2. 自动回退 `/<repo-name>`

## 项目结构

```text
folio/
├── main.go
├── cmd/build/main.go
├── internal/folio/
│   ├── folio.go
│   ├── view.go
│   ├── routing.go
│   └── static_build.go
├── config.json
├── posts/
├── test/
├── themes/
├── .github/workflows/pages.yml
└── README.md
```
