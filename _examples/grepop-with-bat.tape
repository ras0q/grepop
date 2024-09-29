Output grepop-with-bat.gif

Source config.tape

Require unbuffer
Require bat

Type "unbuffer bat --plain --plain --theme OneHalfDark --language go go.mod | grepop go" Sleep 500ms Enter

Sleep 5s
