exports.generate = (input) => ({
  files: [
    {
      path: "special.txt",
      content: input.options.note,
    },
  ],
});
