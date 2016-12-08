ROOTCERF=~/go/src/github.com/wantonsolutions/obeah
ROOTPON=~/pro/go/src/github.com/wantonsolutions/obeah

if [ ! -d $ROOTCERF ]; then
    ROOT=$ROOTPON
    echo "working on pon"
else
    ROOT=$ROOTCERF
    echo "working on cerf"
fi


ARCHIVE=$ROOT/test/log_archive
TESTFILE=t.go

#install obeah
cd $ROOT
go install

#run on the test program
cd test
obeah -file=$TESTFILE -v

#run the test program
go run t.go

#cleanup
if [ ! -d $ARCHIVE ]; then
    mkdir $ARCHIVE
fi

dir="$ARCHIVE/$(date +%m-%d_%H-%M-%S)/"
echo "moving left-overs to $dir"
mkdir -p "$dir"

#archive the test file and its encoded state
mv $TESTFILE *.enc $dir
#restore the testfile
cp ./clean/$TESTFILE ./

