#!/bin/bash -e
cd "$(dirname "$0")"

rm -rf docs/app-dev
bash build.sh -O
rm -rf docs/app
cp -a docs/app-dev docs/app

# patch docs/index.html
VERSION=$(git rev-parse HEAD)

sed -E 's/\/\*VERSION\*\/".*"\/\*VERSION\*\//\/*VERSION*\/"'$VERSION'"\/*VERSION*\//g' \
  docs/index.html > docs/.index.html.tmp
mv -f docs/.index.html.tmp docs/index.html
