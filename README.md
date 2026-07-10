# bubbles

A collection of components for [Bubble Tea v2](https://github.com/charmbracelet/bubbletea).

Each component lives in its own directory with its own Go module, so it can be
imported independently.

## Components

- [navstack](./navstack) — a generic navigation stack for building multi-screen
  wizards, with `Push`/`Replace`/`Pop` and built-in back-navigation.

  ![navstack example](./navstack/example.gif)

- [prompt](./prompt) — a configurable inline prompt bubble that accepts one of
  a set of single-character keys and echoes the chosen key inline.

  ![prompt example](./prompt/example.gif)

---

Sponsored by [imgproxy](https://imgproxy.net).
