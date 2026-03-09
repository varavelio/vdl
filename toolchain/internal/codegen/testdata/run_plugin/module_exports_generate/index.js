module.exports.generate = (input) => ({
  files: [
    {
      path: "module.txt",
      content: `module=${input.options.module}`,
    },
  ],
});
