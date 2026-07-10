Output ../example.gif

Set Shell "bash"
Set FontSize 16
Set Width 700
Set Height 600
Set Padding 20
Set Theme "Dracula"

Hide
Type "clear && go build -o /tmp/navstack-example . && /tmp/navstack-example"
Enter
Sleep 3s
Show
Sleep 1.2s

# --- TailView phase: only the current step is ever shown ---
Enter
Sleep 1.2s

Escape
Sleep 1.2s

Enter
Sleep 1.2s

Enter
Sleep 1.5s

Enter
Sleep 1.2s

# --- SequenceView phase: completed steps stay stacked above the active one ---
Enter
Sleep 1.2s

Escape
Sleep 1.2s

Enter
Sleep 1.2s

Enter
Sleep 1.5s

Hide
Enter
Sleep 2s
