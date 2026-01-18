# URPC Formatter Tests

## Overview

This directory contains test files designed to verify the functionality of the
URPC Formatter.

## Test File Naming Convention

All test files must follow the naming convention:

`NNNN-<descriptive-name>.urpc`

Where:

- `NNNN` is a three-digit number (e.g., `0001`, `0015`, `1234`).
- `<descriptive-name>` briefly explains the purpose of the test (e.g.,
  `docstrings`, `nested-fields`).

Example: `0001-simple-comment.urpc`

## Test File Structure

Each test file must contain two sections separated by a specific comment:

1. The **unformatted** URPC code.
2. The **expected formatted** URPC code after the formatter runs.

The structure within the file must be as follows:

```urpc
<unformatted code>

// >>>>

<formatted code>
```

Note the empty lines before and after the separator comment those are required.

## Test Execution Process

A dedicated test script processes these files. The script performs the following
actions:

1. Reads the entire content of a `.urpc` test file.
2. Locates the mandatory separator comment: `// >>>>`.
3. Splits the file content into two parts based on this comment.

**Important:**

- The separator comment **must be exactly** `// >>>>`. Any variation will cause
  the test script to fail.
- During the splitting process, the line _immediately before_ the `// >>>>`
  comment and the line _immediately after_ it are **removed** and not included
  in either the unformatted or formatted code sections used for comparison.
  Ensure your unformatted and formatted code blocks account for this.
