exports.generate = (input) => ({
  files: [
    {
      path: "input.txt",
      content: `${input.version}|${input.ir.entryPoint}|${input.options.module}`,
    },
  ],
});
