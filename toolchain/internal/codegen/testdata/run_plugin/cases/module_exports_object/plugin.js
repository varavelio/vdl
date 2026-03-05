module.exports = {
  generate: (input) => ({
    files: [
      {
        path: "module-object.txt",
        content: input.options.target,
      },
    ],
  }),
};
