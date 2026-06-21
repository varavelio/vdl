// Shared E2E tests plugin
// Writes the generator-facing IR to output.json for golden comparison.

function normalizeIR(ir) {
  const normalized = stripPositions(ir);
  normalized.entryPoint = "input.vdl";
  return normalized;
}

function stripPositions(value) {
  if (Array.isArray(value)) {
    return value.map(stripPositions);
  }

  if (value && typeof value === "object") {
    const out = {};
    for (const [key, item] of Object.entries(value)) {
      if (key !== "position") {
        out[key] = stripPositions(item);
      }
    }
    return out;
  }

  return value;
}

exports.generate = (input) => ({
  files: [
    {
      path: "output.json",
      content: `${JSON.stringify(normalizeIR(input.ir), null, 2)}\n`,
    },
  ],
});
