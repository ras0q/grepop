# grepop

Interactive and pop `grep` powered by [Charm](https://github.com/charmbracelet) libraries.

## Usage

```bash
$ grepop --help
Usage: grep [option]... PATTERN

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