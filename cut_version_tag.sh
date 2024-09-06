#!/usr/bin/env bash

echo "checking out main branch"
git checkout main
echo "pulling latest changes"
git pull

version=`cat VERSION`
full_version="$version"

echo "current git HEAD is \"$(git log --oneline |head -1)\""
read -p "Would you like to create and push the tag ${full_version} at the current head of the master branch? (y/n) " proceed

if [[ ${proceed} == "y" ]]; then
    git tag "${full_version}"
    git push --tags
fi