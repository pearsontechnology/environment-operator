#!/bin/bash

source ~/environment.sh

cloneMaster(){
    rm -Rf /tmp/environment-operator > /dev/null 2>&1
    git clone --depth=50 --branch=master git@github.com:pearsontechnology/environment-operator.git /tmp/environment-operator
    cd /tmp/environment-operator
}

mergeIntoMaster(){
    git remote set-branches --add origin dev
    git fetch
    git merge --no-ff --no-edit origin/dev
}

tagmaster(){
    git tag -d $releaseVersion
    git push origin :refs/tags/$releaseVersion
    git tag -a -m $releaseVersion $releaseVersion
    git push origin master
    git push origin --tags
}


#If it's not a PR, the branch is dev and there is a releaseVersion, push dev to master and tag master.
if [ $TRAVIS_PULL_REQUEST == "false" ] && [ $TRAVIS_BRANCH == "dev" ] && [ ! -z "$releaseVersion" ]; then

        echo "---------------------------------------------------------------------------------------------------------------------------------"
        echo "------------------------  Merge 'dev' to 'master' and Tag Release -----------------------------------------------------------"
        echo "---------------------------------------------------------------------------------------------------------------------------------"

        cloneMaster
        mergeIntoMaster
        tagmaster
fi
