#!/bin/bash

gnome-terminal -e 'sh -c "go run SharedResource.go"'
gnome-terminal -e 'sh -c "go run Process.go 1 :10001 :10002 :10003 :10004"'
gnome-terminal -e 'sh -c "go run Process.go 2 :10001 :10002 :10003 :10004"'
gnome-terminal -e 'sh -c "go run Process.go 3 :10001 :10002 :10003 :10004"'
gnome-terminal -e 'sh -c "go run Process.go 4 :10001 :10002 :10003 :10004"'
