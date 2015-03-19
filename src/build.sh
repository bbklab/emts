#!/usr/bin/env bash

export GOPATH=~/goshare/:~/emts
export GOBIN=${GOPATH}/bin/

packages=(
	"worker"
	"inc"
)

for p in ${packages[*]} 
do
	echo -e "building package ${p} ... \c"
	if go build ${p}; then
		echo "Done"
	else
		echo "Fail"
	fi
done

echo -e "main.go building ... \c"
if go build main.go; then
	echo "Done"
else
	echo "Fail"
fi
