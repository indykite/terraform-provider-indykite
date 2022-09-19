module.exports = {
  extends: ["@commitlint/config-conventional"],
  plugins: ['commitlint-plugin-function-rules'],
  rules: {
    //   0 - Disabled, 1 - Warning, 2 - Error
    "body-max-line-length": [2, "always", 72],
    "body-leading-blank": [1, "always"],
    "header-max-length": [0], // level: disabled [2, "always", 72]
    "function-rules/header-max-length": [
      2, // level: error
      'always',
      (parsed) => {
        if (parsed.scope === 'deps' || parsed.scope === 'dependencies') {
          return [true];
        }
        if (parsed.header.length <= 72) {
          return [true];
        }
        return [false, 'header must not be longer than 72 characters'];
      },
    ],
    "subject-max-length": [2, "always", 50],
    "subject-full-stop": [2, "never", "."],
    "subject-case": [2, "always", ["lower-case"]],
    "type-enum": [
      2,
      "always",
      ["build", "chore", "ci", "docs", "feat", "fix", "perf", "refactor", "revert", "style", "test"],
    ],
    "scope-enum": [
      2,
      "always",
      ["logging", "sdk", "docs", "dependencies", "auth", "deps", "build", "test", "ci"],
    ],
  },
};
