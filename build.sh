#!/bin/bash
set -ex
base_pkg=github.com/tgascoigne/ragekit

function build_for() {
	goos=$1
	goarch=$2

	outdir=$(readlink -f ${ARTIFACTS_DIR})/${goos}-${goarch}
	mkdir -p $outdir

	(cd $outdir;
	for f in `go list $base_pkg/cmd/...`; do
		GOOS=$goos GOARCH=$goarch go build $f;
	done;
	cd ..;
	zip -9 -rD ragekit-${goos}-${goarch}.zip $outdir;
	rm -rf $outdir)
}

go get $base_pkg/...
#build_for windows 386 
build_for windows amd64 
#build_for linux 386 
build_for linux amd64
#build_for darwin 386 
build_for darwin amd64
