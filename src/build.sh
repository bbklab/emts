#!/usr/bin/env bash

export GOPATH=~/goshare/:~/emts
export GOBIN=${GOPATH}/bin/

packages=(
	"worker"
	"inc"
)

for p in ${packages[*]} 
do
	if go build ${p}; then
		:
	else
		echo -e "building package ${p} ... Fail"
	fi
done

if go build main.go; then
	:
else
	echo -e "main.go building ... Fail"
fi
