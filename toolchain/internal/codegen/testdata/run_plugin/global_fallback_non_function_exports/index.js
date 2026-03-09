exports.generate = 123;
module.exports = {
  generate: 456,
};

function generate() {
  return {
    files: [
      {
        path: "fallback.txt",
        content: "global",
      },
    ],
  };
}
