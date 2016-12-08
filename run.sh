ROOT=~/go/src/github.com/wantonsolutions/obeah

cd $ROOT
go install
cd test
obeah -file=t.go -v
