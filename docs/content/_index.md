+++
title = "Varavel Definition Language"
description = "VDL is an open-source, type-safe, multi-language, and easily extensible schema definition language and code generation toolchain."
+++

{{ header_base(
show_title=false,
container="lg",
menu="Overview|/docs/essentials/about/,Language|/docs/language/start-here/,Plugins|/docs/guides/plugins/,Reference|/docs/reference/spec/",
cta_text="Install",
cta_url="/docs/guides/installation/"
) }}

{{ hero_split(
container="lg",
eyebrow="Contract-first generation",
title="One human-readable schema. Type-safe code across your stack.",
description="VDL lets teams write compact contracts that humans can review and machines can validate. From one .vdl file, generate type-safe models, RPC clients and servers, schemas, docs, and custom artifacts for multiple languages.",
action_1_text="Get started",
action_1_url="/docs/",
action_2_text="See available plugins",
action_2_url="/docs/guides/plugins/",
caption="Open source, deterministic, and designed for normal development and CI workflows.",
panel_label="How it works",
panel_title="Readable contract to type-safe outputs",
panel_desc="A VDL project starts with declarations, not framework code. The analyzer enforces the model, then plugins generate the files each application needs.",
metric_1_value=".vdl",
metric_1_label="source files",
metric_2_value="Plugin",
metric_2_label="transforms IR",
metric_3_value="Code",
metric_3_label="type-safe outputs",
check_1="Write contracts that are easy to read in code review",
check_2="Validate references, naming, spreads, literals, and required type relationships",
check_3="Generate type-safe code for frontend, backend, services, and tooling"
) }}

{{ metrics_strip(
container="lg",
label="What VDL gives you",
stat_1_value="Readable",
stat_1_label="contracts for humans",
stat_2_value="Type-safe",
stat_2_label="generated code",
stat_3_value="Multi-lang",
stat_3_label="plugin outputs",
stat_4_value="CI-ready",
stat_4_label="format and checks"
) }}

{{ features_grid(
container="lg",
eyebrow="The core idea",
title="Keep the contract simple. Move the output logic into plugins.",
description="VDL separates what your systems agree on from how each toolchain consumes it. The source stays readable, while generated artifacts stay precise and type-safe.",
columns="3",
item_1_icon="braces",
item_1_title="Describe the shared model",
item_1_desc="Use type, enum, const, include, docstrings, arrays, maps, inline objects, references, and spreads to describe the contract itself in a review-friendly format.",
item_2_icon="workflow",
item_2_title="Add domain intent",
item_2_desc="Use annotations such as @rpc, @event, @deprecated, or your own team-specific tags without changing the language grammar.",
item_3_icon="plug",
item_3_title="Generate type-safe outputs",
item_3_desc="Official and custom easy to write plugins receive the resolved IR and generate language-specific code, schemas, docs, or internal artifacts."
) }}

{{ steps_cards(
container="lg",
eyebrow="Workflow",
title="A straightforward path from schema to artifacts",
description="The day-to-day loop is intentionally small: write the contract, configure plugins, and generate checked output.",
columns="3",
item_1_icon="terminal",
item_1_title="Install the CLI",
item_1_desc="Install VDL with the shell installer, Homebrew, PowerShell, npm, Docker, or release binaries.",
item_1_href="/docs/guides/installation/",
item_1_link="Install VDL",
item_2_icon="braces",
item_2_title="Model the contract",
item_2_desc="Create readable .vdl files for shared types, HTTP RPC operations, event payloads, documentation, and metadata.",
item_2_href="/docs/language/start-here/",
item_2_link="Learn the language",
item_3_icon="zap",
item_3_title="Run generation",
item_3_desc="Point vdl.config.vdl at your schema and plugins, then run vdl generate locally or in CI.",
item_3_href="/docs/guides/configuration/",
item_3_link="Configure a project"
) }}

{{ features_grid(
container="lg",
bg="base-200",
eyebrow="Use cases",
title="Use VDL anywhere systems need to agree on data",
description="VDL is most useful when several parts of a product need the same contract, but each part needs it in its own language or format.",
columns="3",
item_1_icon="box",
item_1_title="Share types across apps",
item_1_desc="Generate type-safe models for frontend, backend, workers, and microservices from the same source contract.",
item_2_icon="code-xml",
item_2_title="Build type-safe HTTP RPC",
item_2_desc="Model request and response shapes once, then generate typed clients, servers, and OpenAPI documents for frontend-to-backend communication.",
item_3_icon="workflow",
item_3_title="Coordinate microservices",
item_3_desc="Keep service boundaries explicit with shared schemas, generated packages, and stable contracts that can be checked in CI.",
item_4_icon="book-open",
item_4_title="Document and validate data",
item_4_desc="Generate JSON Schema, schema explorers, metadata exports, and documentation from docstrings and resolved types.",
item_5_icon="plug",
item_5_title="Standardize event contracts",
item_5_desc="Use @event declarations to generate subject builders and catalogs for event-driven systems.",
item_6_icon="terminal",
item_6_title="Private team artifacts",
item_6_desc="Write custom plugins for your framework conventions, templates, catalogs, policies, or platform-specific glue code."
) }}

{{ resource_cards(
container="lg",
eyebrow="Read next",
title="Choose the next step that matches your goal",
description="The documentation starts with concepts, then moves into installation, configuration, plugin usage, and reference details.",
item_1_type="Overview",
item_1_title="About VDL",
item_1_desc="Understand the project philosophy, the plugin-first model, and common use cases.",
item_1_href="/docs/essentials/about/",
item_1_meta="Read overview",
item_2_type="Language",
item_2_title="Start Here",
item_2_desc="Learn the VDL syntax in order, from files and comments to naming and validation.",
item_2_href="/docs/language/start-here/",
item_2_meta="Learn syntax",
item_3_type="Project setup",
item_3_title="Project Configuration",
item_3_desc="Configure vdl.config.vdl, plugin sources, output directories, remotes, hooks, and lock files.",
item_3_href="/docs/guides/configuration/",
item_3_meta="Configure generation",
item_4_type="Extension",
item_4_title="Creating Plugins",
item_4_desc="Build custom JavaScript or TypeScript plugins that consume the resolved VDL IR.",
item_4_href="/docs/guides/creating-plugins/",
item_4_meta="Write a plugin"
) }}

{{ cta_banner(
container="lg",
eyebrow="Ready to try it",
title="Install VDL, write one schema, and generate your first outputs.",
description="Start small: define a type, add a plugin, and run the generator. The same workflow scales to shared models, RPC services, event contracts, documentation, and private platform tooling.",
action_1_text="Install VDL",
action_1_url="/docs/guides/installation/",
action_2_text="Open the language guide",
action_2_url="/docs/language/start-here/",
caption="Generated files are regular files: inspect them, test them, commit them, or regenerate them in CI."
) }}

{{ footer_simple(
container="lg",
links="Docs|/docs/,GitHub|https://github.com/varavelio/vdl,Varavel|https://varavel.com",
show_github="false"
) }}
