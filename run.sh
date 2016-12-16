ROOTCERF=~/go/src/github.com/wantonsolutions/obeah
ROOTPON=~/pro/go/src/github.com/wantonsolutions/obeah
RUN=true

if [ ! -d $ROOTCERF ]; then
    ROOT=$ROOTPON
    echo "working on pon"
else
    ROOT=$ROOTCERF
    echo "working on cerf"
fi


ARCHIVE=$ROOT/test/log_archive
# options [t.go , t2.go, t3.go, t4.go]
TESTFILE=t2.go

#install obeah
cd $ROOT
go install

#run on the test program
cd test
cp clean/$TESTFILE test.go
obeah -file=test.go -v


run the test program
go run test.go

#generate cfg
dot -Tpng -o runtimeCFG.png runtimeCFG.dot
display runtimeCFG.png &


#cleanup
if [ ! -d $ARCHIVE ]; then
    mkdir $ARCHIVE
fi

dir="$ARCHIVE/$(date +%m-%d_%H-%M-%S)/"
echo "moving latest to $dir"
mv $ARCHIVE/latest $dir
mkdir -p $ARCHIVE/latest

#archive the test file and its encoded state
mv test.go *.enc *.png *.dot $ARCHIVE/latest
#restore the testfile

