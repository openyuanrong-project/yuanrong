#!/bin/bash

check_tool() {
    local cmd="$1"
    local version_cmd="$2"
    local name="$3"
    local version="NOT INSTALLED"

    if command -v "$cmd" &> /dev/null; then
        if [ -n "$version_cmd" ]; then
            version=$(eval "$version_cmd" 2>&1 | head -n1)
        else
            case "$name" in
                "Java")
                    version=$(java -version 2>&1 | head -n1 | cut -d'"' -f2)
                    ;;
                "protoc")
                    version=$(protoc --version 2>&1 | awk '{print $2}')
                    ;;
                "Python 3.9")
                    version=$(python3.9 --version 2>&1 | awk '{print $2}')
                    ;;
                "Python 3.10")
                    version=$(python3.10 --version 2>&1 | awk '{print $2}')
                    ;;
                "Python 3.11")
                    version=$(python3.11 --version 2>&1 | awk '{print $2}')
                    ;;
                "glibc")
                    version=$(ldd --version 2>&1 | head -n1 | awk '{print $NF}')
                    ;;
                "Ninja")
                    version="OK (via ninja-build)"
                    ;;
                *)
                    version="OK"
                    ;;
            esac
        fi
    fi
    printf "%-20s %s\n" "$name" "$version"
}

check_tools() {
    local all_ok=true

    echo "=========================================="
    printf "%-20s %s\n" "Tool" "Version"
    echo "------------------------------------------"

    check_tool java "" "Java"
    check_tool mvn "mvn -version 2>&1 | grep 'Apache Maven' | awk '{print \$3}'" "Maven"
    check_tool go "go version 2>&1 | awk '{print \$3}'" "Go"
    check_tool bazel "bazel --version 2>&1 | grep -o 'bazel [0-9.]*' | awk '{print \$2}'" "Bazel"
    check_tool python3.9 "" "Python 3.9"
    check_tool python3.10 "" "Python 3.10"
    check_tool python3.11 "" "Python 3.11"
    check_tool ninja "" "Ninja"
    check_tool protoc "" "protoc"
    check_tool cmake "cmake --version 2>&1 | head -n1 | awk '{print \$3}'" "CMake"
    check_tool gcc "gcc --version 2>&1 | head -n1 | awk '{print \$3}'" "GCC"
    check_tool ldd "" "glibc"
    check_tool node "node -v" "Node.js"

    echo "=========================================="

    local cmd
    local tools=(
        "java"
        "mvn"
        "go"
        "bazel"
        "python3.9"
        "python3.10"
        "python3.11"
        "ninja"
        "protoc"
        "cmake"
        "gcc"
        "ldd"
        "node"
    )

    for cmd in "${tools[@]}"; do
        if ! command -v "$cmd" &> /dev/null; then
            all_ok=false
            break
        fi
    done

    if $all_ok; then
        return 0
    else
        return 1
    fi
}