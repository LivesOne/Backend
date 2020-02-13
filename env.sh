#!/bin/bash

oldGOPATH=$GOPATH

#export GOPATH=$PWD:$GOPATH

if [ $GOPATH ]; then  
    export GOPATH=$PWD:$GOPATH
else  
    export GOPATH=$PWD
fi  
