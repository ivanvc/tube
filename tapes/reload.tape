Output images/reload.gif
Set Width 1400
Set Height 600
Set Theme "Dracula"
Set FontFamily "Source Code Pro"
Set LetterSpacing 1

Hide
Type "export PS1='> '; export PATH=$PWD/bin:$PATH; mkdir bin"
Enter
Type "go build -o bin/tube cmd/tube/main.go"
Enter
Type "clear"
Enter
Type "/bin/sh -c 'sleep 5; touch go.mod' &"
Enter
Type "tube -watch 8080 python -m http.server 8080"
Enter
Show

Sleep 10s

Hide
Type "q"
Type "rm bin/tube"
Enter
