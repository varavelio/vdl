import { describe, expect, it } from "vitest";

import { getLangExtension } from "./getLangExtension";

describe("getLangExtension", () => {
  it("should return the correct extension for a known language", () => {
    expect(getLangExtension("javascript")).toBe("js");
    expect(getLangExtension("python")).toBe("py");
    expect(getLangExtension("typescript")).toBe("ts");
    expect(getLangExtension("csharp")).toBe("cs");
    expect(getLangExtension("java")).toBe("java");
  });

  it("should be case-insensitive for known languages", () => {
    expect(getLangExtension("JavaScript")).toBe("js");
    expect(getLangExtension("PYTHON")).toBe("py");
  });

  it("should handle language aliases", () => {
    expect(getLangExtension("js")).toBe("js");
    expect(getLangExtension("py")).toBe("py");
    expect(getLangExtension("ts")).toBe("ts");
    expect(getLangExtension("sh")).toBe("sh");
    expect(getLangExtension("shell")).toBe("sh");
  });

  it("should return the fallback extension for an unknown language", () => {
    expect(getLangExtension("unknown-language")).toBe("txt");
  });

  it("should return the custom fallback extension for an unknown language", () => {
    expect(getLangExtension("unknown-language", "dat")).toBe("dat");
  });

  it("should return the correct extension for languages with the same extension", () => {
    expect(getLangExtension("yaml")).toBe("yaml");
    expect(getLangExtension("yml")).toBe("yaml");
    expect(getLangExtension("ansible")).toBe("yaml");
  });

  it("should handle an empty string as input", () => {
    expect(getLangExtension("")).toBe("txt");
  });

  it("should handle a language with leading/trailing spaces", () => {
    expect(getLangExtension("  javascript  ")).toBe("js");
  });

  it("should return the correct extension for Objective-C", () => {
    expect(getLangExtension("objectivec")).toBe("m");
    expect(getLangExtension("objc")).toBe("m");
  });

  it("should return sh for wget", () => {
    expect(getLangExtension("wget")).toBe("sh");
  });

  it("should return correct extension for vdl", () => {
    expect(getLangExtension("vdl")).toBe("vdl");
  });
});
