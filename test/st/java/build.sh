#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
set -e

BASE_DIR=$(cd "$(dirname "$0")"; pwd)
YUANRONG_DIR=""
PKG_DIR=${BASE_DIR}/../functions/pkg
STJAVA_LIB=$PKG_DIR/lib
STJAVA=$PKG_DIR/stjava
SDK_DIR=$BASE_DIR/../../../output/yuanrong/runtime/sdk/java

readonly USAGE="
Usage: bash build.sh [-h] [-y path-to-yuanrong]

Options:
    -h Output this help and exit.
    -y YuanRong dir.
Example:
  $ bash build.sh
"

usage() {
    echo -e "$USAGE"
}

get_opt() {
    while getopts 'y:h' opt; do
    case "$opt" in
    y)
        YUANRONG_DIR="${OPTARG}"
        SDK_DIR=$YUANRONG_DIR/runtime/sdk/java
        ;;
    h)
        usage
        exit 0
        ;;
    *)
        echo "invalid command"
        usage
        exit 1
        ;;
    esac
done
}

function install_sdk() {
    cd $SDK_DIR
    jarfile=$(find ./ -name "yr-api-sdk-*.jar" -print -quit)
    echo "[INFO] --java st-- Installing java sdk from ${SDK_DIR}/${jarfile}"
    mvn -q install:install-file -Dfile=${jarfile} -DartifactId=yr-api-sdk -DgroupId=org.yuanrong \
    -Dversion=v0.0.1 -Dpackaging=jar -DpomFile=./pom.xml
}

function package() {
    echo "[INFO] --java st-- Generating test code .jar file"
    cd $BASE_DIR
    mvn clean package -q -Dmaven.test.skip=true -Dmaven.wagon.http.ssl.insecure=true -Dmaven.wagon.http.ssl.allowall=true
}

function create_zip() {
    echo "[INFO] --java st-- Generating 'stjava.zip' file in test/st/functions/pkg/."
    # The file structure under test/st/functions/:
    # ├── service.yaml
    # ├── pkg
    # │   └──stjava.zip
    rm -rf $STJAVA_LIB && mkdir -p $STJAVA_LIB
    cp ${BASE_DIR}/target/*.jar $STJAVA_LIB/
    cd $PKG_DIR && zip -r stjava.zip lib/
    rm -rf $STJAVA_LIB
    cd ${BASE_DIR}
}

main() {
    get_opt "$@"
    install_sdk
    package
    create_zip
}

main "$@"