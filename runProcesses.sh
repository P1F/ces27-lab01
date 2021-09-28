#!/bin/bash

gnome-terminal --geometry 900x500+0+0 --title="SharedResource" -e 'sh -c "go run resource/SharedResource.go" & PIDSR=$!'
gnome-terminal --geometry 900x500+0+0 --title="Process 1" -e 'sh -c "go run Process.go 1 :10002 :10003 :10004 :10005" & PIDP1=$!'
gnome-terminal --geometry 900x500+1200+0 --title="Process 2" -e 'sh -c "go run Process.go 2 :10002 :10003 :10004 :10005" & PIDP2=$!'
gnome-terminal --geometry 900x500+0+1000 --title="Process 3" -e 'sh -c "go run Process.go 3 :10002 :10003 :10004 :10005" & PIDP3=$!'
gnome-terminal --geometry 900x500+1200+1000 --title="Process 4" -e 'sh -c "go run Process.go 4 :10002 :10003 :10004 :10005" & PIDP4=$!'
wait $PIDP4 $PIDP3 $PIDP2 $PIDP1 $PIDSR