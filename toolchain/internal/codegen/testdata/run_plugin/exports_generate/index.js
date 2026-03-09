exports.generate = (input) => ({
  files: [
    {
      path: "exports.txt",
      content: `version=${input.version};target=${input.options.target}`,
    },
  ],
});
