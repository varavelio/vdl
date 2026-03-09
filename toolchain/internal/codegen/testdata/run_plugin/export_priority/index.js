exports.generate = () => ({
  files: [
    {
      path: "winner.txt",
      content: "exports",
    },
  ],
});

module.exports = {
  generate: () => ({
    files: [
      {
        path: "winner.txt",
        content: "module",
      },
    ],
  }),
};

function generate() {
  return {
    files: [
      {
        path: "winner.txt",
        content: "global",
      },
    ],
  };
}
