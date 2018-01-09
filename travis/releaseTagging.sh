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

notifyHipchat(){
    echo "A new Environment Operator release is now available." > /tmp/notification
    awk '/RELEASED/{p++}p==1' /tmp/environment-operator CHANGELOG.md >> /tmp/notification
    cat /tmp/notification | ./${TRAVIS_BUILD_DIR}/hipchat/hipchat_room_message -t ${hipchat_token} -r ${hipchat_room} -f "TravisCI"
}


#If it's not a PR, the branch is dev and it is a release , push dev to master and tag master.
if [ $TRAVIS_PULL_REQUEST == "false" ] && [ $TRAVIS_BRANCH == "dev" ] && [ $release == "true" ]; then

        echo "---------------------------------------------------------------------------------------------------------------------------------"
        echo "------------------------  Merge 'dev' to 'master' and Tag Release -----------------------------------------------------------"
        echo "---------------------------------------------------------------------------------------------------------------------------------"

        cloneMaster
        mergeIntoMaster
        tagmaster
        notifyHipchat
fi
