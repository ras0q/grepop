# grepop

Pop `grep` powered by [Charm](https://github.com/charmbracelet) libraries.

![preview](./preview.gif)

## Usage

```plaintext
$ grepop --help
Usage: grepop [option]... PATTERN

Examples:
  cat access.log | grepop ERROR

Options:
  -border-color uint
        Border foreground color (default 63)
  -color uint
        Foreground color (default 212)
  -debug
        Debug mode
  -height uint
        Percentage of terminal height (default 90)
  -no-border
        Disable popup border
  -sleep uint
        Milliseconds to wait for output (default 500)
```

### Examples

General

```bash
cat main.go | grepop error
```

Colorized output

```bash
unbuffer bat --pager never main.go | grepop error
```
