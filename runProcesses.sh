#!/bin/bash

gnome-terminal --geometry 900x500+0+0 --title="SharedResource" -e 'sh -c "go run resource/SharedResource.go"'
gnome-terminal --geometry 900x500+0+0 --title="Process 1" -e 'sh -c "go run Process.go 1 :10002 :10003 :10004 :10005"'
gnome-terminal --geometry 900x500+1200+0 --title="Process 2" -e 'sh -c "go run Process.go 2 :10002 :10003 :10004 :10005"'
gnome-terminal --geometry 900x500+0+1000 --title="Process 3" -e 'sh -c "go run Process.go 3 :10002 :10003 :10004 :10005"'
gnome-terminal --geometry 900x500+1200+1000 --title="Process 4" -e 'sh -c "go run Process.go 4 :10002 :10003 :10004 :10005"'
