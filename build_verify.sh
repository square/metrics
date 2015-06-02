#/usr/bin/env bash 


# Invoke gofmt on the current directory.
# -l flag causes it to only list files
GOFMT_RESULTS=`gofmt -l .`
if [ "$GOFMT_RESULTS" ]
then
	echo "gofmt finds changes:"
	echo "$GOFMT_RESULTS"
	exit -1
fi


# Find all files ending with .go that do not end with .peg.go (generated)
GOLINT_FILES=`find . -not -name *.peg.go -name *.go`
# Iterate over these (split by newline)

# Check each of them with golint
found=""
#IFS=$'\n' splits the string by its newlines, storing into an array
IFS=$'\n' GOLINT_ARRAY=($GOLINT_FILES)
for file in ${GOLINT_ARRAY[@]}; do
	GOLINT_RESULT=`golint $file`
	if [ "$GOLINT_RESULT" ]
	then
		found="yes"
	fi
done
# If one of them produced output,
# run through again and print any output that occurs
if [ $found ]
then
	echo "Golint has messages:"
	for file in ${GOLINT_ARRAY[@]}; do
		GOLINT_RESULT=`golint $file`
		if [ "$GOLINT_RESULT" ]
		then
			echo "$GOLINT_RESULT"
		fi
	done
	exit -1
fi

#Lastly, make sure calling ./query/build.sh doesn't cause ./query/language.peg.go to change

hash=$(md5 ./query/language.peg.go)
./query/build.sh
newhash=$(md5 ./query/language.peg.go)

if [ "$hash" != "$newhash" ]
then
	echo "Build scripts have not been run:"
	echo "There were changes to query/language.peg without query/language.peg.go being updated"
	exit -1
fi









