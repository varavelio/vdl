// Generates every documentation page, its raw Markdown twin, and the docs
// support routes (search index and llms.txt outputs).
export default function ({ files, parse }) {
  let paths = files.listFiles("content/docs/**/*.md");
  if (paths.length === 0) return [];

  let pages = [];

  for (const path of paths) {
    const pageRaw = files.readFile(path);
    const pageMd = parse.markdown(pageRaw);
    const pageIsDraft = Boolean(pageMd.frontmatter.draft);
    if (pageIsDraft) continue;

    // The `path` frontmatter field overrides the permalink derived from the
    // file path, which keeps published URLs decoupled from file names.
    const override = String(pageMd.frontmatter.path || "").replace(/^\/+|\/+$/g, "");
    const permalink = override
      ? `/${override}/`
      : files.toPermalink(path, { stripPrefix: "content/" });

    const pageMarkdown = pageMd.content;
    const pageContent = parse.renderComponents(pageMd.html);
    const title = String(pageMd.frontmatter.title || "Untitled");
    const description = String(pageMd.frontmatter.description || "");
    const weight = Number(pageMd.frontmatter.weight) || 999999;
    const icon = String(pageMd.frontmatter.icon || "");

    pages.push({
      permalink,
      template: "vara-docs",
      title,
      description,
      weight,
      icon,
      content: pageContent,
      docs_footer_links: [
        {
          title: "Edit this page",
          href: `https://github.com/varavelio/vdl/edit/main/docs/${path}`,
          icon: "pencil-line",
        },
      ],
    });

    pages.push({
      permalink: permalink + "index.md",
      template: "vara-docs-raw",
      title,
      description,
      weight,
      icon,
      content: pageMarkdown,
    });
  }

  pages.push({
    permalink: "/docs/vara-docs-search-index.json",
    template: "vara-docs-search-index",
    sitemap: false,
  });

  pages.push({
    permalink: "/docs/llms.txt",
    template: "vara-docs-llms-txt",
  });

  pages.push({
    permalink: "/docs/llms-full.txt",
    template: "vara-docs-llms-full-txt",
  });

  return pages;
}
