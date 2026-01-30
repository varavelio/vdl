# VDL VS Code Extension Release Instructions

This document outlines the manual process for releasing a new version of the VDL VS Code extension. Given that the extension logic relies primarily on the VDL binary, releases are infrequent and managed manually.

## Prerequisites

- Ensure you are in the `editors/vscode` directory.
- Ensure you have the necessary permissions for the GitHub repository and Marketplace accounts.

---

## Phase 1: Versioning & Documentation

1.  **Update Version Number**
    - Open `editors/vscode/package.json`.
    - Update the `"version"` field following Semantic Versioning (e.g., `0.1.8`).
    - **Note:** Do not include a "v" prefix in the JSON file.

2.  **Update Changelog**
    - Open `editors/vscode/CHANGELOG.md`.
    - Add a new entry for the current version with a date and a list of changes.

3.  **Update README Summary**
    - Open `editors/vscode/README.md`.
    - Update the "Recent Changes" section to reflect the latest top 10 changes.
    - Ensure the link to the full `CHANGELOG.md` is correct and accessible.

4.  **Commit Changes**
    - Commit the version bump and documentation updates.
    - _Example message:_ `chore(vscode): bump version to 0.1.8`

---

## Phase 2: Build & Package

1.  **Generate the VSIX Package**
    Run the following command from the `editors/vscode` directory. This script cleans previous builds, compiles the code, and packages the extension.

    ```bash
    npm run package
    ```

2.  **Verify the Artifact**
    - Confirm that a file named `vdl-X.X.X.vsix` has been created in the directory.
    - (Optional) Verify the package contents to ensure only necessary files are included:
      ```bash
      npx vsce ls
      ```

---

## Phase 3: GitHub Release

1.  **Create a New Release**
    - Go to the [GitHub Releases page](https://github.com/varavelio/vdl/releases).
    - Click **Draft a new release**.

2.  **Tagging**
    - **Tag version:** Create a new tag using the format: `vX.X.X-vscode` (e.g., `v0.1.8-vscode`).
    - _Note:_ This suffix is mandatory to distinguish extension releases from the core VDL binary releases in the monorepo.

3.  **Release Details**
    - **Title:** `VDL VS Code Extension vX.X.X`
    - **Description:** Copy the relevant section from your `CHANGELOG.md`.
    - **Assets:** Upload the generated `.vsix` file to the release assets.

4.  **Publish**
    - Click **Publish release**.

---

## Phase 4: Marketplace Publishing

### 1. Visual Studio Marketplace (Official)

1.  Log in to the [VS Code Marketplace Management Portal](https://marketplace.visualstudio.com/manage).
2.  Select the **vdl** extension.
3.  Click **Update** (or the ellipsis `...` -> **Update**).
4.  Upload the `.vsix` file.
5.  Wait for the verification process to complete.

### 2. Open VSX Registry (Open Source / VSCodium)

1.  Log in to the [Open VSX Registry User Settings](https://open-vsx.org/user-settings/extensions).
2.  Locate the namespace/extension.
3.  Drag and drop the `.vsix` file into the upload area (or click to browse).
4.  Confirm the upload.

---

**Done!** The update should appear in VS Code for users shortly.
