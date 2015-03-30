#!/usr/bin/env bash

basedir="$(cd $(dirname $0);pwd)"
GOPATH=~/goshare/:${basedir}/../

packages=(
	"worker"
	"inc"
)

# Func Def
check_rc() {
  if [ $? == 0 ]; then
        echo -e " -- $(date +%F_%T)  succed!  ${*}"
  else
        echo -e " -- $(date +%F_%T)  failed!  ${*}"; exit 1
  fi  
}

for p in ${packages[*]} 
do
	go build ${p}
	check_rc "building package ${p}"
done

go build emts.go 
check_rc "building execution main"

tempdir="./temp"
destdir="${tempdir}/emts"

/bin/mkdir -p ${destdir}
check_rc "mkdir ${destdir}"

/bin/cp -a ./emts ${destdir}
check_rc "copy emts"

/bin/cp -a ./conf/ ./c/ ./share/ ./sinfo/  ${destdir}
check_rc "copy related dirs"

cd ${tempdir}
check_rc "changing into ${tempdir}"

/bin/tar -czvf ../emts.tgz "emts" 2>&1 1>/dev/null
check_rc "make tarball on emts/"

cd ../
check_rc "changing back"

/bin/rm -rf ./emts ${tempdir}
check_rc "remove useless files"

