#/usr/bin/env bash

# Find all files ending with .go that do not end with .peg.go (generated)
GOLINT_FILES=`find . -not -name *.peg.go -name *.go`
# Iterate over these (split by newline)

# Check each of them with golint
found=""
#IFS=$'\n' splits the string by its newlines, storing into an array
IFS=$'\n' GOLINT_ARRAY=($GOLINT_FILES)
for file in ${GOLINT_ARRAY[@]}; do
	GOLINT_RESULT=`golint $file`
	if [ "$GOLINT_RESULT" ]; then
		found="yes"
	fi
done
# If one of them produced output,
# run through again and print any output that occurs
if [ $found ]; then
	echo "FAIL: UNLINTED FILES:"
	echo "GOLINT FINDS"
	for file in ${GOLINT_ARRAY[@]}; do
		GOLINT_RESULT=`golint $file`
		if [ "$GOLINT_RESULT" ]; then
			echo "$GOLINT_RESULT"
		fi
	done
	exit -1
fi