function generate(input) {
  return {
    errors: [
      {
        message: `global:${input.ir.entryPoint}`,
      },
    ],
  };
}
