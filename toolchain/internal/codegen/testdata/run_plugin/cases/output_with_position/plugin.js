exports.generate = () => ({
  files: [
    {
      path: "generated.txt",
      content: "ok",
    },
  ],
  errors: [
    {
      message: "semantic issue",
      position: {
        file: "/schema/main.vdl",
        line: 7,
        column: 11,
      },
    },
  ],
});
