Output images/tui.gif
Set Width 1400
Set Height 600
Set Theme "Dracula"
Set FontFamily "Source Code Pro"
Set LetterSpacing 1

Hide
Type "export PS1='> '; export PATH=$PWD/bin:$PATH; mkdir bin"
Enter
Type "/bin/sh -c 'while true; do curl localhost:8080 &>/dev/null; sleep 0.75; done' &"
Enter
Type "go build -o bin/tube cmd/tube/main.go"
Enter
Type "clear"
Enter

Show
Type "tube 8080"
Sleep 1s
Enter

Sleep 3s

Type "e"
Sleep 2s
Type@125ms "python -m http.server 8080"
Enter

Sleep 3s

Type "r"

Sleep 3s

Hide
Type "q"
Type "rm bin/tube"
Enter
Type "kill %1"
Enter
