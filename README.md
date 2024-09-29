# grepop

Pop `grep` powered by [Charm](https://github.com/charmbracelet) libraries.

![Made with VHS](https://vhs.charm.sh/vhs-2sjdwJZOpDo3t7Mf1cBcXQ.gif)

## Usage

```plaintext
$ grepop --help
Usage: grepop [option]... PATTERN

Examples:
  cat access.log | grepop ERROR

  unbuffer bat --paging never access.log | grepop ERROR

Options:
  -border-color uint
        Border foreground color (default 63)
  -border-template string
        Border Template (default "┏━┓\n┃ ┃\n┗━┛")
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
  -tab-width uint
        Tab Width (default 8)
```

See [./_examples](./_examples)
