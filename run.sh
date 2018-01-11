export GOPATH=$PWD
rm -rf bin/main
go build main
go install main
bin/main
