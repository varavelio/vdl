# VDL for Neovim

VDL has built-in LSP support so you can use the native Neovim client directly.

## Setup

Add this snippet to your `init.lua`:

```lua
-- Register vdl file type
vim.filetype.add({
  extension = {
    vdl = 'vdl',
  },
})

-- Create a group to avoid duplication
local vdl_group = vim.api.nvim_create_augroup('vdl_lsp', { clear = true })

-- Start the LSP client when opening a .vdl file
vim.api.nvim_create_autocmd('FileType', {
  pattern = 'vdl',
  group = vdl_group,
  callback = function()
    vim.lsp.start({
      name = 'vdl',
      cmd = { 'vdl', 'lsp' },
      -- Find the project root looking for vdl.yaml or .git
      root_dir = vim.fs.dirname(vim.fs.find({'vdl.yaml', '.git'}, { upward = true })[1]),
    })
  end,
})
```

## Requirements

- You must have the `vdl` binary installed and in your PATH.
