#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

CUR_DIR=$(dirname $(readlink -f "$0"))
set -e

THIRD_PARTY_DIR="${CUR_DIR}/../../vendor/"
OPENSOURCE="${CUR_DIR}/openSource.txt"
MODULES="all"
DOWNLOAD_TEST_THIRDPARTY="ON"

LOCAL_OS=$(head -1 /etc/os-release | tail -1 | awk -F "\"" '{print $2}')_$(uname -m)
THIRD_PARTY_CACHE=${THIRD_PARTY_CACHE:-"https://build-logs.openeuler.openatom.cn:38080/temp-archived/openeuler/openYuanrong/deps/"}
echo -e "local os is $LOCAL_OS"


while getopts 'T:M:F:r' opt; do
    case "$opt" in
    T)
        THIRD_PARTY_DIR=$(readlink -f "${OPTARG}")
        ;;
    M)
        MODULES="${OPTARG}"
        ;;
    F)
        OPENSOURCE="${OPTARG}"
        ;;
    r)
        DOWNLOAD_TEST_THIRDPARTY="OFF"
        ;;
    *)
        log_error "Invalid command"
        exit 1
        ;;
    esac
done

if [ ! -d "${THIRD_PARTY_DIR}" ]; then
  mkdir -p "${THIRD_PARTY_DIR}"
fi

IFS=';' read -ra MODULES_MAP <<< "$MODULES"

function checksum_and_decompress() {
    local name="$1"
    local filename="$2"
    local savepath="$THIRD_PARTY_DIR"

    actual_sha256=$(shasum -a 256 "${filename}" | awk '{print $1}')
    if [ "$actual_sha256" != "$sha256" ]; then
        echo "=== download ${name}-${tag} to ${savepath} checksum failed ==="
        cd ..
        rm -rf "${savepath}/${name}"
        return 1
    fi

    echo "process file: ${filename}"
    case "$filename" in
    *.tar.gz)
        echo "use tar to decompress"
        mkdir "${savepath}/${name}"
        tar -zxvf ${filename} -C "$name" --strip-components=1 && rm ${filename}
        ;;
    *.zip)
        echo "use unzip to decompress"
        root_dir=$(unzip -l "${filename}" | awk '$NF ~ /\/$/ {print substr($NF, 1, length($NF)-1); exit}')
        unzip -q ${filename} && rm ${filename}
        mv "${root_dir}" "$name"
        ;;
    *)
        echo "File does not have a .tar.gz/.zip extension."
        ;;
    esac
}

function download_open_src() {
    local name="$1"
    local tag="$2"
    local repo="$3"
    local sha256="$4"
    local savepath="$THIRD_PARTY_DIR"
    echo -e "=== download opensrc ${name}-${tag} to ${savepath}... ==="

    if [ -d "$savepath"/"$name" ]; then
        echo -e "${name} has been downloaded to ${savepath}"
        return 0
    fi

    cd "$savepath"

    local filename="${name}"-"$(basename ${repo})"
    if [ -n "${THIRD_PARTY_CACHE}" ]; then
        if ! wget "${THIRD_PARTY_CACHE}/${filename}"; then
            echo -e "=== download ${name}-${tag} cache to ${savepath} failed ==="
            cd ..
            rm -rf "${savepath}/${name}"
            return 1
        fi
    else
        if ! curl -sS -L ${repo} -o ${filename} --retry 3; then
            echo -e "=== download ${name}-${tag} to ${savepath} failed ==="
            cd ..
            rm -rf "${savepath}/${name}"
            return 1
        fi
    fi

    checksum_and_decompress ${name} ${filename}
}

download_a_repo() {
    local name=$1
    local tag=$2
    local module=$3
    local repo=$4
    local sha256=$5
    local usage=$6

    if [[ "X${usage}" = "Xtest" && "X${DOWNLOAD_TEST_THIRDPARTY}" = "XOFF" ]]; then
        echo -e "${name} is not downloaded."
        return 0
    fi

    echo "begin download $name"
    if [ "${MODULES}" == "all" ]; then
        download_open_src "$name" "$tag" "$repo" "$sha256"
    else
        IFS=';' read -ra module_map <<< "$module"
        for item in "${MODULES_MAP[@]}"
        do
            for m in "${module_map[@]}"
            do
                echo "item is: ${item}, module is: ${m}"
                if [ "${item}" = "$m" ]; then
                    download_open_src "$name" "$tag" "$repo" "$sha256"
                fi
            done
        done
    fi
}

pids=()
while IFS=',' read -r name tag module repo sha256 usage; do
    # Start background process for each task.
    download_a_repo "$name" "$tag" "$module" "$repo" "$sha256" "$usage" 
    pid=$!
    echo "Task PID ${pid}: download repo $repo"
done < "${OPENSOURCE}"

echo "All downloads completed!"
