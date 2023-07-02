Output images/intro.gif
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

Type "tube 8080 python -m http.server 8080"
Sleep 1s
Enter

Sleep 10s

Hide
Type "q"
Enter
Type "rm bin/tube"
Enter
Type "kill %1"
Enter
