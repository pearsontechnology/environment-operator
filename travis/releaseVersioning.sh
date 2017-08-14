#!/bin/bash

git fetch
git fetch --tags

export releaseVersion=`grep -m 1 '### \*\*\[.*..*..*\].*\[RELEASED\]\*\*' ${TRAVIS_BUILD_DIR}/CHANGELOG.md | tail -n1 | cut -d "[" -f2 | cut -d "]" -f1`
export previousReleaseVersion=`grep -m 2 '### \*\*\[.*..*..*\].*\[RELEASED\]\*\*' ${TRAVIS_BUILD_DIR}/CHANGELOG.md | tail -n1 | cut -d "[" -f2 | cut -d "]" -f1`


if [ -z "$releaseVersion" ];then
  echo "export release=false" >> ~/environment.sh
else
  if ! git tag | grep -q $releaseVersion; then   #IF the version found is not already a git tag, it is a new release
    echo "export release=true" >> ~/environment.sh
    echo "export releaseVersion=$releaseVersion" >> ~/environment.sh

    echo "---------------------------------------------------------------------------------------------------------------------------------"
    echo "------------------------  Release Detected in ${TRAVIS_BUILD_DIR}/CHANGELOG.md"
    echo "------------------------  Found New Version (${releaseVersion}) to tag and release"
    echo "---------------------------------------------------------------------------------------------------------------------------------"

  else  #It's not a new release
    echo "export release=false" >> ~/environment.sh
    echo "export releaseVersion=$releaseVersion" >> ~/environment.sh

  fi
fi
