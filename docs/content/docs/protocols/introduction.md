+++
title = "VDL Plugin Protocols"
description = "What VDL Plugin Protocols are, why they exist, and how they keep the ecosystem consistent."
template = "docs.html"
path = "docs/protocols/introduction"
weight = 1
+++

A **VDL Plugin Protocol** (**VPP**) is a written standard for a family of VDL plugins.

It defines what a plugin is allowed to understand, what it must validate, and what kind of output it must produce. In practical terms, a VPP turns a convention into a shared contract for the whole VDL ecosystem.

VDL itself is intentionally small. The core language gives you a compact way to describe schemas with declarations such as types, enums, constants and annotations. It does not try to hard-code every possible domain into the compiler. Instead, VDL grows indefinitely through plugins.

The plugin-first model is powerful, but it needs discipline. Without shared protocols, every plugin could interpret the same annotation differently, produce incompatible APIs, or apply different validation rules. VPPs prevent that fragmentation.

Plugins are not required to use a VPP. A plugin can be completely ad hoc, especially when it exists for one project, one team, or one private workflow. VPPs become useful when the same pattern appears across projects often enough that it is worth standardizing. At that point, a protocol lets multiple plugins implement the same idea across different languages and stacks, so VDL continues to feel familiar no matter what environment you are working in.

## Why VPPs Exist

VPPs exist so VDL can stay small while the ecosystem remains consistent.

The VDL compiler should focus on the language itself: parsing files, building an intermediate representation, resolving declarations, and exposing structured data to plugins. Domain-specific behavior belongs outside the compiler, where plugins can evolve independently.

A VPP defines the rules for that domain-specific behavior.

For example, if a protocol introduces an annotation, the protocol must explain what the annotation means, where it can be used, what arguments it accepts, what invalid usage looks like, and how compliant plugins should reflect it in generated artifacts.

This gives VDL three important properties.

1. **A small core:** the compiler does not grow every time the ecosystem needs a new capability.
2. **Portable semantics:** the same VDL file can be understood consistently by plugins for different languages and different purposes.
3. **Composable tooling:** plugins can build on top of the same conventions instead of inventing private dialects.

## What a VPP Defines

A VPP does not define new VDL syntax. It defines meaning on top of the existing language.

Protocols are built from standard VDL features such as annotations, declaration names, metadata, literals, and the resolved intermediate representation passed to plugins. The protocol says how those pieces must be interpreted.

A good VPP clearly specifies these areas.

1. **Vocabulary:** the annotations, metadata fields, declaration shapes, and names that belong to the protocol.
2. **Allowed targets:** where that vocabulary can appear, such as types, fields, enum members, constants, or protocol-specific containers.
3. **Validation rules:** what a compliant plugin must check before generating output.
4. **Output expectations:** what generated code, documentation, diagnostics, or metadata should look like.
5. **Interoperability rules:** how the protocol should compose with other protocols and plugins.

The goal is not to force every language target to generate identical source code. Go, TypeScript, Rust, Java, and other ecosystems have different idioms. The goal is to guarantee that all compliant plugins preserve the same semantics while generating idiomatic output for their target language.

## How VPPs Keep Plugins Compatible

Plugins are free to specialize, but they should not redefine shared meaning.

If a VPP says that an annotation has a specific behavior, a compliant plugin must follow that behavior. It may generate language-specific APIs, add useful helper files, or integrate with a framework, but it must not silently change the meaning of the protocol.

This is especially important when plugins are chained together. A foundational plugin may generate basic models, while another plugin may generate services, events, policies, adapters, or documentation that references those models. The only way that composition stays reliable is if each plugin can depend on stable protocol semantics.

In other words, a VPP is the agreement that lets independently developed plugins behave like part of one coherent system.

## What VPPs Are Not

A VPP is not a compiler feature request.

A VPP should not expand the VDL grammar. Most ecosystem capabilities should be modeled with existing VDL declarations and annotations, then implemented by plugins.

A VPP is also not a single implementation. Multiple plugins can implement the same protocol for different languages, frameworks, or runtime environments. The protocol is the standard and each plugin is one implementation of that standard.

## Defining a Protocol

When defining or proposing a VPP, write it as a contract for implementers.

The document should be precise enough that two independent plugin authors can read it, implement it separately, and produce compatible behavior. It should explain the intent in plain language, then define the exact rules that make the behavior testable.

Any person, team, or organization can define their own VPPs for internal use. Private protocols are useful when a company wants consistent behavior across its own plugins without waiting for an ecosystem-wide standard.

Official VDL protocols are different. To become an official VPP, a proposal must go through a formal community discussion on GitHub. That discussion is where the vocabulary, validation rules, edge cases, and long-term compatibility concerns are refined before the protocol becomes part of the public VDL standard set.

At minimum, a protocol should answer these questions.

1. What problem does this protocol solve?
2. Which VDL constructs does it use?
3. Which annotations or metadata fields does it define?
4. Where are those annotations or fields allowed?
5. Which argument types are accepted?
6. What must happen when invalid input is found?
7. What should compliant generators produce?
8. How should the protocol interact with other VPPs?

The best VPPs are narrow, explicit, and composable. They define one domain well, avoid leaking implementation details into the language, and leave room for plugins to generate idiomatic code in their target ecosystem.
