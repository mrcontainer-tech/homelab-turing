# Vim Cheat Sheet

Quick reference for essential Vim shortcuts, nvimtree, and fuzzy finding.

---

## Modes

| Key | Action |
|-----|--------|
| `i` | Insert mode (before cursor) |
| `a` | Insert mode (after cursor) |
| `A` | Insert mode (end of line) |
| `o` | Insert mode (new line below) |
| `O` | Insert mode (new line above) |
| `v` | Visual mode (character) |
| `V` | Visual mode (line) |
| `Ctrl+v` | Visual block mode |
| `Esc` or `Ctrl+[` | Normal mode |

---

## Navigation

| Key | Action |
|-----|--------|
| `h` `j` `k` `l` | Left, Down, Up, Right |
| `w` | Next word |
| `b` | Previous word |
| `e` | End of word |
| `0` | Start of line |
| `^` | First non-blank character |
| `$` | End of line |
| `gg` | Start of file |
| `G` | End of file |
| `:{n}` or `{n}G` | Go to line n |
| `%` | Jump to matching bracket |
| `Ctrl+d` | Half page down |
| `Ctrl+u` | Half page up |
| `Ctrl+f` | Full page down |
| `Ctrl+b` | Full page up |

---

## Editing

| Key | Action |
|-----|--------|
| `x` | Delete character |
| `dd` | Delete line |
| `dw` | Delete word |
| `d$` or `D` | Delete to end of line |
| `yy` | Copy (yank) line |
| `yw` | Copy word |
| `p` | Paste after cursor |
| `P` | Paste before cursor |
| `u` | Undo |
| `Ctrl+r` | Redo |
| `cc` | Change entire line |
| `cw` | Change word |
| `r` | Replace single character |
| `>>` | Indent line |
| `<<` | Unindent line |
| `.` | Repeat last command |

---

## Search & Replace

| Key | Action |
|-----|--------|
| `/pattern` | Search forward |
| `?pattern` | Search backward |
| `n` | Next match |
| `N` | Previous match |
| `*` | Search word under cursor (forward) |
| `#` | Search word under cursor (backward) |
| `:%s/old/new/g` | Replace all in file |
| `:%s/old/new/gc` | Replace all with confirmation |
| `:noh` | Clear search highlighting |

---

## File Operations

| Key | Action |
|-----|--------|
| `:w` | Save |
| `:q` | Quit |
| `:wq` or `ZZ` | Save and quit |
| `:q!` | Quit without saving |
| `:e filename` | Open file |
| `:bn` | Next buffer |
| `:bp` | Previous buffer |
| `:bd` | Close buffer |
| `:ls` | List buffers |

---

## Visual Mode

| Key | Action |
|-----|--------|
| `v` then movement | Select text |
| `V` | Select entire lines |
| `y` | Yank (copy) selection |
| `d` | Delete selection |
| `>` | Indent selection |
| `<` | Unindent selection |
| `=` | Auto-indent selection |

---

## nvimtree (File Explorer)

### Opening nvimtree

| Key | Action |
|-----|--------|
| `:NvimTreeToggle` | Toggle file tree |
| `:NvimTreeFocus` | Focus file tree |
| `:NvimTreeFindFile` | Reveal current file |

Common keybinding: `<Space>e` or `Ctrl+n` (depends on config)

### Inside nvimtree

| Key | Action |
|-----|--------|
| `Enter` or `o` | Open file/folder |
| `a` | Create new file/folder |
| `d` | Delete file/folder |
| `r` | Rename file/folder |
| `x` | Cut file/folder |
| `c` | Copy file/folder |
| `p` | Paste file/folder |
| `y` | Copy filename |
| `Y` | Copy relative path |
| `gy` | Copy absolute path |
| `R` | Refresh tree |
| `W` | Collapse all folders |
| `E` | Expand all folders |
| `q` | Close nvimtree |
| `?` | Show help (all keybindings) |
| `H` | Toggle hidden files |
| `I` | Toggle .gitignore files |

---

## Fuzzy Finder

### Telescope (Neovim)

| Key | Action |
|-----|--------|
| `:Telescope find_files` | Find files |
| `:Telescope live_grep` | Search in files (grep) |
| `:Telescope buffers` | List open buffers |
| `:Telescope help_tags` | Search help |
| `:Telescope git_files` | Git files only |
| `:Telescope oldfiles` | Recent files |

Common keybindings (if configured):
- `<Space>ff` - Find files
- `<Space>fg` - Live grep
- `<Space>fb` - Buffers
- `<Space>fh` - Help tags

### Inside Telescope

| Key | Action |
|-----|--------|
| `Ctrl+n` / `Ctrl+p` | Next/previous item |
| `Down` / `Up` | Next/previous item |
| `Enter` | Open selection |
| `Ctrl+x` | Open in horizontal split |
| `Ctrl+v` | Open in vertical split |
| `Ctrl+t` | Open in new tab |
| `Ctrl+u` | Scroll preview up |
| `Ctrl+d` | Scroll preview down |
| `Esc` | Close Telescope |

### fzf (Classic Vim)

| Command | Action |
|---------|--------|
| `:Files` | Find files |
| `:GFiles` | Git files |
| `:Buffers` | Open buffers |
| `:Rg` | Ripgrep search |
| `:Lines` | Search lines in loaded buffers |
| `:BLines` | Search lines in current buffer |

Common keybindings:
- `Ctrl+p` - Find files
- `Ctrl+f` - Search in files

---

## Window Management

| Key | Action |
|-----|--------|
| `:split` or `:sp` | Horizontal split |
| `:vsplit` or `:vsp` | Vertical split |
| `Ctrl+w h/j/k/l` | Navigate windows |
| `Ctrl+w w` | Cycle through windows |
| `Ctrl+w =` | Equalize window sizes |
| `Ctrl+w q` | Close window |
| `Ctrl+w o` | Close all other windows |

---

## Tabs

| Key | Action |
|-----|--------|
| `:tabnew` | New tab |
| `:tabn` or `gt` | Next tab |
| `:tabp` or `gT` | Previous tab |
| `:tabclose` | Close tab |
| `{n}gt` | Go to tab n |

---

## Useful Tips

### Combining Commands
- `d5w` - Delete 5 words
- `y3j` - Yank 3 lines down
- `c2w` - Change 2 words
- `5dd` - Delete 5 lines

### Marks
- `ma` - Set mark 'a' at cursor
- `'a` - Jump to mark 'a'
- `:marks` - List all marks

### Macros
- `qa` - Record macro in register 'a'
- `q` - Stop recording
- `@a` - Play macro 'a'
- `@@` - Repeat last macro

### Folding
- `zf` - Create fold (visual mode)
- `zo` - Open fold
- `zc` - Close fold
- `za` - Toggle fold
- `zR` - Open all folds
- `zM` - Close all folds

---

## Quick Config Check

Check your current Vim configuration:

```bash
# Check if nvimtree is installed
:help nvim-tree

# Check if telescope is installed
:help telescope.nvim

# Check if fzf is installed
:FZF

# List all installed plugins
:PlugStatus    # vim-plug
:Lazy          # lazy.nvim
:PackerStatus  # packer.nvim
```

---

## Learning Path

1. **Day 1-3**: Master basic navigation (`hjkl`, `w`, `b`, `0`, `$`, `gg`, `G`)
2. **Day 4-7**: Practice editing (`dd`, `yy`, `p`, `u`, `i`, `a`, `o`)
3. **Week 2**: Visual mode and search (`v`, `V`, `/`, `n`, `*`)
4. **Week 3**: File operations and buffers (`:w`, `:e`, `:bn`, `:bp`)
5. **Week 4+**: Advanced features (macros, splits, fuzzy finder, nvimtree)

**Practice tip**: Use `vimtutor` command for interactive lessons!
