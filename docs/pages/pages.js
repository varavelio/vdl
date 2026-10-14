// Generates the landing page and the site support routes (404 and sitemap).
export default function ({ data, files, parse }) {
  let pages = [];

  const homePageRaw = files.readFile("content/index.md");
  const homePageMd = parse.markdown(homePageRaw);
  const homePageContent = parse.renderComponents(homePageMd.html);

  pages.push({
    permalink: "/",
    template: "vara-landing",
    title: homePageMd.frontmatter.title || data.site.title,
    description: homePageMd.frontmatter.description || data.site.description,
    content: homePageContent,
  });

  const pluginsRaw = files.readFile("content/plugins.md");
  const pluginsMd = parse.markdown(pluginsRaw);
  const pluginsContent = parse.renderComponents(pluginsMd.html);

  pages.push({
    permalink: "/plugins/",
    template: "vara-landing",
    title: pluginsMd.frontmatter.title || "Plugin Explorer",
    description: pluginsMd.frontmatter.description || data.site.description,
    content: pluginsContent,
  });

  pages.push({
    permalink: "/404.html",
    template: "vara-404",
    title: "Page not found",
    description: "The requested page could not be found.",
    sitemap: false,
    content: "",
  });

  pages.push({
    permalink: "/sitemap.xml",
    template: "vara-sitemap-xml",
  });

  return pages;
}
