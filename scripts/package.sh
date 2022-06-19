#!/bin/bash

set -e
PRJROOT=$(dirname "${BASH_SOURCE[0]}")/..
cd $PRJROOT

PKGNAME="$1"
PLATFORM="$2"
GOOS="$3"
GOARCH="$4"
VERSION='unknown'

if [ -f "./version.txt" ]; then
    VERSION=`cat ./version.txt`
elif [ -d ".git" ]; then
    VERSION=$(git describe --tags --abbrev=0)
fi

echo Packaging $PKGNAME $VERSION $PLATFORM $GOARCH

# Remove previous build directory, if needed.
bdir=$PKGNAME-$VERSION-$GOOS-$GOARCH
rm -rf packages/$bdir && mkdir -p packages/$bdir

case $PKGNAME in
    *) 
        declare -a BINS=( $PKGNAME )
        declare -a DOCS=("README.md", "server-config-sample.hcl")
        ;;
esac

for BIN in $BINS; do
    # Make the binaries.
    GOOS=$GOOS GOARCH=$GOARCH make $BIN

    # Copy the executable binaries.
    if [ "$GOOS" == "windows" ]; then
        mv tmp/$BIN packages/$bdir/$BIN.exe
    else
        mv tmp/$BIN packages/$bdir
    fi
done


# Copy documention and license.
for D in $DOCS; do
    cp $D packages/$bdir
done

# Compress the package.
cd packages
zip -r -q $bdir.zip $bdir
# if [ "$GOOS" == "linux" ]; then
# 	tar -zcf $bdir.tar.gz $bdir
# else
# 	zip -r -q $bdir.zip $bdir
# fi

# Remove build directory.
rm -rf $bdir
