# Koi — Web UI Design

> Current: Single HTML file with inline CSS/JS.
> Future: Modular structure with Monaco Editor and data visualization.

## Current Implementation

Single file: `web/index.html` (~300 lines)

### Layout

```
┌─────────────────────────────────────────────────┐
│ Header: 🐟 Koi | Run | Clear | Settings         │
├──────────┬──────────────────────┬───────────────┤
│ Sidebar  │ Editor               │ Console       │
│ (files)  │ (textarea)           │ (output)      │
│          │                      │               │
├──────────┴──────────────────────┴───────────────┤
│ Status bar: Ready | Koi v0.1.0-mvp              │
└─────────────────────────────────────────────────┘
```

### Features

- **Code Editor**: Textarea with tab support, Ctrl+Enter to run
- **File Tree**: Browse virtual filesystem, click to navigate
- **Console**: Captured io.print output, error highlighting
- **Settings Modal**: Edit config (security, UI) via web form

## Planned Improvements

### Monaco Editor

Replace textarea with VS Code's editor component:
- Lua syntax highlighting
- Auto-completion for math/fs/io/os APIs
- Error markers
- Multi-file tabs

### Data Visualization

Interactive charts for computation results:
- Matrix heatmaps
- Function plots
- Statistical distributions
- 3D surface plots

### Real-Time Output

WebSocket-based streaming:
- Live console output during execution
- Filesystem change notifications
- Execution progress for long tasks

### Responsive Design

Current layout works on desktop. Planned:
- Collapsible sidebar on mobile
- Touch-friendly controls
- Adaptive font sizes
