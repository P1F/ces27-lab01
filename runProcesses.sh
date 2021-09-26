#!/bin/bash

gnome-terminal --title="SharedResource" -e 'sh -c "go run SharedResource.go"'
gnome-terminal --title="Process 1" -e 'sh -c "go run Process.go 1 :10002 :10003 :10004 :10005"'
gnome-terminal --title="Process 2" -e 'sh -c "go run Process.go 2 :10002 :10003 :10004 :10005"'
gnome-terminal --title="Process 3" -e 'sh -c "go run Process.go 3 :10002 :10003 :10004 :10005"'
gnome-terminal --title="Process 4" -e 'sh -c "go run Process.go 4 :10002 :10003 :10004 :10005"'
