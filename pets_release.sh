#!/usr/bin/env bash

# builds and zips pets binaries for macOS and Linux
# should be run from .../go/.../windmilleng/pets

# Requirements:
# - gothub (https://github.com/itchio/gothub)
# - valid github token in environ: $GITHUB_TOKEN="abcd..."

if [[ -z "$(which gothub)" ]]; then
    echo "This script requires gothub (https://github.com/itchio/gothub)."
    echo "Install with: go get github.com/itchio/gothub"
fi

if [[ -z ${GITHUB_TOKEN} ]]; then
    echo "\$GITHUB_TOKEN not set, aborting."
    exit 1
fi

mkdir dist; cd dist

# macOS bin
echo "Building and zipping macOS binary..."
env GOOS=darwin GOARCH=amd64 go build -v -o pets github.com/windmilleng/pets
zip -m pets_macos_amd64.zip pets

echo "Building and zipping Linux binary..."
env GOOS=linux GOARCH=amd64 go build -v -o pets github.com/windmilleng/pets
zip -m pets_linux_amd64.zip pets

echo "Enter tag for new release: "
read TAG

echo "Enter description for new release (leave blank to just use tag): "
read DESC

if [[ -z ${DESC} ]]; then
  DESC=${TAG}
fi

echo "Pushing release to GitHub..."
gothub release --user windmilleng --repo pets --tag ${TAG} --description ${DESC}
gothub upload --user windmilleng --repo pets --tag ${TAG} --name pets_linux_amd64.zip --file pets_linux_amd64.zip
gothub upload --user windmilleng --repo pets --tag ${TAG} --name pets_macos_amd64.zip --file pets_macos_amd64.zip

echo "Verify your release at: https://github.com/windmilleng/pets/releases/latest"
