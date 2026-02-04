import { describe, expect, it } from "vitest";

import { slugify } from "./slugify.ts";

describe("slugify", () => {
  it("should convert to lowercase", () => {
    expect(slugify("Hello World")).toBe("hello-world");
  });

  it("should replace spaces with hyphens", () => {
    expect(slugify("hello world")).toBe("hello-world");
  });

  it("should handle multiple spaces", () => {
    expect(slugify("hello   world")).toBe("hello-world");
  });

  it("should trim whitespace", () => {
    expect(slugify("  hello world  ")).toBe("hello-world");
  });

  it("should remove special characters", () => {
    expect(slugify("hello@world!")).toBe("helloworld");
  });

  it("should preserve numbers", () => {
    expect(slugify("hello123world")).toBe("hello123world");
  });

  it("should handle multiple consecutive hyphens", () => {
    expect(slugify("hello--world")).toBe("hello-world");
  });

  it("should remove leading hyphens", () => {
    expect(slugify("-hello")).toBe("hello");
  });

  it("should remove trailing hyphens", () => {
    expect(slugify("hello-")).toBe("hello");
  });

  it("should handle empty string", () => {
    expect(slugify("")).toBe("");
  });

  it("should handle string with only spaces", () => {
    expect(slugify("   ")).toBe("");
  });

  it("should preserve hash symbols", () => {
    expect(slugify("hello#world")).toBe("hello#world");
  });

  it("should preserve multiple hash symbols", () => {
    expect(slugify("hello##world")).toBe("hello##world");
  });

  it("should preserve hash at the beginning", () => {
    expect(slugify("#hello")).toBe("#hello");
  });

  it("should preserve hash at the end", () => {
    expect(slugify("hello#")).toBe("hello#");
  });

  it("should handle complex string with hashes and spaces", () => {
    expect(slugify("Hello World #Section")).toBe("hello-world-#section");
  });

  it("should handle anchor-like patterns", () => {
    expect(slugify("docs#getting-started")).toBe("docs#getting-started");
  });

  it("should handle mixed special characters keeping only hashes", () => {
    expect(slugify("hello@#world!")).toBe("hello#world");
  });
});
