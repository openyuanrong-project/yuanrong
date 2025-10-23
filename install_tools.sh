#!/bin/bash
set -e

# ==============================================================================
# Build Tools Installer (Simplified, Full English)
# Installs ninja via yum, then creates 'ninja' symlink from 'ninja-build'
# All packages use Huawei Cloud mirrors.
# ==============================================================================
source check_tools.sh

# check environment
if check_tools; then
    echo "Build environment is ready."
    echo "Run: source /etc/profile.d/*.sh to load in current shell."
    exit 0
fi

# Configuration
TMP_DIR="/tmp"
BUILD_TOOLS="/opt/buildtools"

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        ARCH_TAG="x64"
        BAZEL_ARCH="x86_64"
        echo "Detected x86_64 architecture"
        ;;
    aarch64)
        ARCH_TAG="aarch64"
        BAZEL_ARCH="arm64"
        echo "Detected aarch64 (ARM64) architecture"
        ;;
    *)
        echo "Error: Unsupported architecture: $ARCH" >&2
        exit 1
        ;;
esac

# Ensure sudo is available
if ! command -v sudo &> /dev/null; then
    echo "Warning: sudo not found, installing..."
    yum install -y sudo || echo "Failed to install sudo, proceeding..."
fi

# Create build tools directory
echo "Creating $BUILD_TOOLS..."
sudo mkdir -p "$BUILD_TOOLS"
sudo chown "$(whoami):$(whoami)" "$BUILD_TOOLS" || true

# ==============================================================================
# 0. Install system dependencies including ninja-build
# ==============================================================================
echo "Installing system dependencies via yum..."
sudo yum install -y \
    gcc gcc-c++ git make libstdc++-devel curl libtool \
    binutils binutils-devel libcurl-devel zlib zlib-devel automake expat-devel \
    python3-pip python3-mock patch openssl openssl-devel libssh-devel libffi-devel \
    sqlite-devel bison gawk texinfo glibc glibc-devel wget bzip2-devel sudo \
    rsync nfs-utils xz libuuid unzip util-linux-devel cpio libcap-devel libatomic \
    chrpath numactl-devel openeuler-lsb libasan dos2unix net-tools pigz cmake \
    protobuf protobuf-devel patchelf golang ninja-build doxygen  || \
    echo "Dependency installation completed or partially failed"

# ==============================================================================
# 1. Setup 'ninja' command via symlink to 'ninja-build'
# ==============================================================================
echo "Setting up 'ninja' command..."
if [ -f "/usr/bin/ninja-build" ] && [ ! -f "/usr/bin/ninja" ]; then
    sudo ln -sf ninja-build /usr/bin/ninja
    echo "Created symlink: /usr/bin/ninja -> /usr/bin/ninja-build"
else
    echo "ninja command already exists or ninja-build not installed"
fi


# ==============================================================================
# 2. Define package names and URLs
# ==============================================================================
# JDK 8
JDK_TAR="OpenJDK8U-jdk_${ARCH_TAG}_linux_hotspot_8u382b05.tar.gz"
JDK_URL="https://mirrors.huaweicloud.com/eclipse/temurin-compliance/temurin/8/jdk8u382-b05/$JDK_TAR"

# Maven 3.9.11
MAVEN_TAR="apache-maven-3.9.11-bin.tar.gz"
MAVEN_URL="https://mirrors.huaweicloud.com/apache/maven/maven-3/3.9.11/binaries/$MAVEN_TAR"

# Go 1.21.4 source
GO_SRC_TAR="go1.21.4.src.tar.gz"
GO_REPO_URL="https://gitee.com/src-openeuler/golang.git"
GO_REPO_BRANCH="openEuler-24.03-LTS"

# Python versions
PYTHON_VERSIONS=("3.9.11" "3.10.2" "3.11.4")
declare -A PYTHON_URLS
for ver in "${PYTHON_VERSIONS[@]}"; do
    major_minor=$(echo "$ver" | cut -d. -f1-2)
    PYTHON_URLS["$ver"]="https://mirrors.huaweicloud.com/python/${ver}/Python-${ver}.tgz"
done

# Bazel 6.5.0
BAZEL_BINARY="bazel-6.5.0-linux-$BAZEL_ARCH"
BAZEL_URL="https://mirrors.huaweicloud.com/bazel/6.5.0/$BAZEL_BINARY"


# ==============================================================================
# 3. Install tools: JDK, Maven, Go, Python, Bazel
# ==============================================================================

# --- Install JDK 8 ---
if [ ! -d "$BUILD_TOOLS/jdk8" ]; then
    echo "Installing JDK 8..."
    if [ ! -f "$TMP_DIR/$JDK_TAR" ]; then
        wget -O "$TMP_DIR/$JDK_TAR" "$JDK_URL"
    fi
    tar -xf "$TMP_DIR/$JDK_TAR" -C "$BUILD_TOOLS"
    mv "$BUILD_TOOLS/jdk8u382-b05" "$BUILD_TOOLS/jdk8"
else
    echo "JDK already installed at $BUILD_TOOLS/jdk8"
fi

# --- Install Bazel ---
if [ ! -f "/usr/local/bin/bazel" ]; then
    echo "Installing Bazel 6.5.0..."
    sudo wget -O /usr/local/bin/bazel "$BAZEL_URL"
    sudo chmod +x /usr/local/bin/bazel
else
    echo "Bazel already installed at /usr/local/bin/bazel"
fi

# --- Install Maven ---
if [ ! -d "$BUILD_TOOLS/apache-maven-3.9.11" ]; then
    echo "Installing Maven..."
    if [ ! -f "$TMP_DIR/$MAVEN_TAR" ]; then
        wget -O "$TMP_DIR/$MAVEN_TAR" "$MAVEN_URL"
    fi
    tar -xf "$TMP_DIR/$MAVEN_TAR" -C "$BUILD_TOOLS"

else
    echo "Maven already installed"
fi

# --- Build and Install Go from source ---
if [ ! -d "$BUILD_TOOLS/golang_go-1.21.4" ]; then
    echo "Building Go 1.21.4 from source..."
    if [ ! -f "$TMP_DIR/$GO_SRC_TAR" ]; then
        git clone "$GO_REPO_URL" -b "$GO_REPO_BRANCH" "$TMP_DIR/golang_repo"
        cp "$TMP_DIR/golang_repo/$GO_SRC_TAR" "$TMP_DIR/"
    fi
    tar -xzf "$TMP_DIR/$GO_SRC_TAR" -C "$TMP_DIR"
    cd "$TMP_DIR/go/src" && ./make.bash
    mv "$TMP_DIR/go" "$BUILD_TOOLS/golang_go-1.21.4"
    sudo yum remove -y golang || true
else
    echo "Go already built and installed"
fi

# --- Install Python versions ---
install_python() {
    local version=$1
    local major_minor=$(echo "$version" | cut -d. -f1-2)
    local tarball="Python-${version}.tgz"
    local url="${PYTHON_URLS[$version]}"

    if [ -d "$BUILD_TOOLS/python$major_minor" ]; then
        echo "Python $version already installed"
        return
    fi

    echo "Installing Python $version..."

    if [ ! -f "$TMP_DIR/$tarball" ]; then
        wget -O "$TMP_DIR/$tarball" "$url"
    fi

    tar -xzf "$TMP_DIR/$tarball" -C "$TMP_DIR"
    cd "$TMP_DIR/Python-$version"
    ./configure \
        --prefix="$BUILD_TOOLS/python$major_minor" \
        --with-openssl=/usr \
        --enable-loadable-sqlite-extensions \
        --enable-shared
    make -j $(nproc)
    sudo make install
    sudo chmod -R 755 "$BUILD_TOOLS/python$major_minor"
}

for ver in "${PYTHON_VERSIONS[@]}"; do
    install_python "$ver"
done

# ==============================================================================
# 4. Set environment variables (Huawei Cloud mirrors)
# ==============================================================================

# --- JAVA & MAVEN ---
sudo tee /etc/profile.d/buildtools-java-maven.sh > /dev/null << 'EOF'
export JAVA_HOME=/opt/buildtools/jdk8
export MAVEN_HOME=/opt/buildtools/apache-maven-3.9.11
export PATH=$JAVA_HOME/bin:$MAVEN_HOME/bin:$PATH
EOF

export JAVA_HOME=/opt/buildtools/jdk8
export MAVEN_HOME=/opt/buildtools/apache-maven-3.9.11
export PATH=$JAVA_HOME/bin:$MAVEN_HOME/bin:$PATH

mkdir -p "$MAVEN_HOME/conf"
    cat > "$MAVEN_HOME/conf/settings.xml" << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<settings xmlns="http://maven.apache.org/SETTINGS/1.0.0"
          xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
          xsi:schemaLocation="http://maven.apache.org/SETTINGS/1.0.0 http://maven.apache.org/xsd/settings-1.0.0.xsd">
  <mirrors>
    <mirror>
      <id>huaweicloud</id>
      <mirrorOf>central</mirrorOf>
      <url>https://repo.huaweicloud.com/repository/maven/</url>
    </mirror>
  </mirrors>
</settings>
EOF

# --- Go ---
sudo tee /etc/profile.d/buildtools-go.sh > /dev/null << 'EOF'
export GOROOT=/opt/buildtools/golang_go-1.21.4
export GOPATH=/opt/buildtools/go_workspace
export GO111MODULE=on
export GONOSUMDB=*
export GOPROXY=https://goproxy.cn,direct
export PATH=$GOROOT/bin:$GOPATH/bin:$PATH
EOF

export GOROOT=/opt/buildtools/golang_go-1.21.4
export GOPATH=/opt/buildtools/go_workspace
mkdir -p "$GOPATH"
chmod 777 -R "$GOPATH"
export GO111MODULE=on
export GONOSUMDB=*
export GOPROXY=https://goproxy.cn,direct
export PATH=$GOROOT/bin:$GOPATH/bin:$PATH

go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# --- Python ---
sudo tee /etc/profile.d/python-env.sh > /dev/null << 'EOF'
export PYTHON_PATH_3911=/opt/buildtools/python3.9
export PYTHON_PATH_3102=/opt/buildtools/python3.10
export PYTHON_PATH_3114=/opt/buildtools/python3.11
export PATH=$PYTHON_PATH_3911/bin:$PYTHON_PATH_3102/bin:$PYTHON_PATH_3114/bin:$PATH
export LD_LIBRARY_PATH=$PYTHON_PATH_3911/lib:$PYTHON_PATH_3102/lib:$PYTHON_PATH_3114/lib:/usr/lib:/usr/lib64:$LD_LIBRARY_PATH
EOF

export PYTHON_PATH_3911=/opt/buildtools/python3.9
export PYTHON_PATH_3102=/opt/buildtools/python3.10
export PYTHON_PATH_3114=/opt/buildtools/python3.11
export PATH=$PYTHON_PATH_3911/bin:$PYTHON_PATH_3102/bin:$PYTHON_PATH_3114/bin:$PATH
export LD_LIBRARY_PATH=$PYTHON_PATH_3911/lib:$PYTHON_PATH_3102/lib:$PYTHON_PATH_3114/lib:/usr/lib:/usr/lib64:$LD_LIBRARY_PATH

# Create symlinks for Python and pip
sudo ln -sf "$BUILD_TOOLS/python3.9/bin/python3.9" "/usr/local/bin/python3.9"
sudo ln -sf "$BUILD_TOOLS/python3.9/bin/pip3.9" "/usr/local/bin/pip3.9"
sudo ln -sf "$BUILD_TOOLS/python3.10/bin/python3.10" "/usr/local/bin/python3.10"
sudo ln -sf "$BUILD_TOOLS/python3.10/bin/pip3.10" "/usr/local/bin/pip3.10"
sudo ln -sf "$BUILD_TOOLS/python3.11/bin/python3.11" "/usr/local/bin/python3.11"
sudo ln -sf "$BUILD_TOOLS/python3.11/bin/pip3.11" "/usr/local/bin/pip3.11"
sudo ln -sf /usr/local/bin/python3.9 /usr/bin/python3 || true
sudo ln -sf /usr/bin/python3 /usr/bin/python || true

# Update shared library cache
echo "$PYTHON_PATH_3911/lib" | sudo tee /etc/ld.so.conf.d/python3.9.conf > /dev/null
echo "$PYTHON_PATH_3102/lib" | sudo tee /etc/ld.so.conf.d/python3.10.conf > /dev/null
echo "$PYTHON_PATH_3114/lib" | sudo tee /etc/ld.so.conf.d/python3.11.conf > /dev/null
sudo ldconfig

# Set locale and disable timeout
echo 'export LANG=C.UTF-8' | sudo tee /etc/profile.d/lang.sh > /dev/null
export LANG=C.UTF-8

echo 'export TMOUT=0' | sudo tee /etc/profile.d/timeout.sh > /dev/null
export TMOUT=0


# ==============================================================================
# 5. Configure pip to use Huawei Cloud mirror
# ==============================================================================
echo "Configuring pip to use Huawei Cloud mirror..."
pip3.9 config --user set global.index-url https://mirrors.huaweicloud.com/repository/pypi/simple
pip3.9 config --user set global.trusted-host mirrors.huaweicloud.com
pip3.9 install wheel==0.36.2 six==1.10.0

pip3.10 config --user set global.index-url https://mirrors.huaweicloud.com/repository/pypi/simple
pip3.10 config --user set global.trusted-host mirrors.huaweicloud.com
pip3.10 install wheel==0.36.2

pip3.11 config --user set global.index-url https://mirrors.huaweicloud.com/repository/pypi/simple
pip3.11 config --user set global.trusted-host mirrors.huaweicloud.com
pip3.11 install wheel==0.36.2


# ==============================================================================
# 6. Verify essential tools
# ==============================================================================
if check_tools; then
    echo "Build environment is ready."
    echo "Run: source /etc/profile.d/*.sh to load in current shell."
    exit 0
else
    echo "Error, init environment failed"
    exit 1
fi