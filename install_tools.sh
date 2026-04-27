#!/bin/bash
set -e

source check_tools.sh || { echo "❌ check_tools.sh not found"; exit 1; }

# Early exit if already installed
if check_tools; then
    echo "✅ Build environment is ready."
    echo "👉 Run: source /etc/profile.d/buildtools.sh to load in current shell."
    exit 0
fi

# --- Detect architecture ---
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        PKG_ARCH="x64"
        OBS_ARCH="x64"
        ;;
    aarch64)
        PKG_ARCH="arm64"
        OBS_ARCH="arm"
        ;;
    *)
        echo "❌ Unsupported architecture: $ARCH" >&2
        exit 1
        ;;
esac

# --- install yum repos ---
sudo yum install -y \
    gcc gcc-c++ git make libstdc++-devel curl libtool \
    binutils binutils-devel libcurl-devel zlib zlib-devel automake expat-devel \
    python3-pip python3-mock patch openssl openssl-devel libssh-devel libffi-devel \
    sqlite-devel bison gawk texinfo glibc glibc-devel wget bzip2-devel sudo \
    rsync nfs-utils xz libuuid unzip util-linux-devel cpio libcap-devel libatomic \
    chrpath numactl-devel openeuler-lsb libasan dos2unix net-tools pigz cmake \
    protobuf protobuf-devel patchelf libibverbs-devel libibverbs

echo "✅ Detected architecture: $ARCH → using package suffix: $PKG_ARCH"

PKG_DIR="$(pwd)/offline-pkgs"
BUILD_TOOLS="/opt/buildtools"
REMOTE_BASE="https://openyuanrong.obs.cn-southwest-2.myhuaweicloud.com/build_tools/openeuler_22.03_LTS/${OBS_ARCH}"

mkdir -p "$PKG_DIR"

# --- Required packages ---
REQUIRED_PKGS=(
    "jdk8-linux-${PKG_ARCH}.tar.gz"
    "apache-maven-3.9.11-bin.tar.gz"
    "go-1.24.1-linux-${PKG_ARCH}.tar.gz"
    "python-3.9.11-linux-${PKG_ARCH}.tar.gz"
    "python-3.10.2-linux-${PKG_ARCH}.tar.gz"
    "python-3.11.4-linux-${PKG_ARCH}.tar.gz"
    "bazel-6.5.0-linux-${PKG_ARCH}.tar.gz"
    "ninja-1.12.0-linux-${PKG_ARCH}.tar.gz"
    "nodejs-20.19.0-linux-${PKG_ARCH}.tar.gz"
)

# --- Ensure all packages are present (download if missing) ---
echo "🔍 Checking and fetching required packages..."
for pkg in "${REQUIRED_PKGS[@]}"; do
    local_path="$PKG_DIR/$pkg"

    if [ -f "$local_path" ]; then
        echo "✅ Found locally: $pkg"
    else
        # Determine remote subpath: most go under ${PKG_ARCH}/, but Maven is arch-independent
        remote_url="${REMOTE_BASE}/${pkg}"
        echo "⬇️  Downloading missing package: $pkg"
        echo "   From: $remote_url"

        # Use curl if available, fallback to wget
        if command -v curl >/dev/null 2>&1; then
            curl -fL --retry 3 --connect-timeout 10 --max-time 60 "$remote_url" -o "$local_path"
        elif command -v wget >/dev/null 2>&1; then
            wget -q --timeout=10 --tries=3 -O "$local_path" "$remote_url"
        else
            echo "❌ Neither curl nor wget found. Cannot download $pkg." >&2
            exit 1
        fi

        if [ ! -s "$local_path" ]; then
            echo "❌ Download failed or empty file: $pkg" >&2
            rm -f "$local_path"
            exit 1
        fi

        echo "✅ Downloaded: $pkg"
    fi
done

echo "✅ All packages ready."

# --- Rest of installation logic (unchanged) ---
sudo mkdir -p "$BUILD_TOOLS"
sudo chown "$(whoami):$(whoami)" "$BUILD_TOOLS" || true

echo "📦 Installing JDK..."
sudo tar -xzf "$PKG_DIR/jdk8-linux-${PKG_ARCH}.tar.gz" -C "$BUILD_TOOLS"

echo "📦 Installing Maven..."
sudo tar -xzf "$PKG_DIR/apache-maven-3.9.11-bin.tar.gz" -C "$BUILD_TOOLS"

echo "📦 Installing Go..."
sudo tar -xzf "$PKG_DIR/go-1.24.1-linux-${PKG_ARCH}.tar.gz" -C "$BUILD_TOOLS"

echo "📦 Installing Python versions..."
sudo tar -xzf "$PKG_DIR/python-3.9.11-linux-${PKG_ARCH}.tar.gz" -C /
sudo tar -xzf "$PKG_DIR/python-3.10.2-linux-${PKG_ARCH}.tar.gz" -C /
sudo tar -xzf "$PKG_DIR/python-3.11.4-linux-${PKG_ARCH}.tar.gz" -C /
if [[ -f "$PKG_DIR/python-3.12.0-linux-${PKG_ARCH}.tar.gz" ]]; then
    sudo tar -xzf "$PKG_DIR/python-3.12.0-linux-${PKG_ARCH}.tar.gz" -C /
fi
if [[ -f "$PKG_DIR/python-3.13.0-linux-${PKG_ARCH}.tar.gz" ]]; then
    sudo tar -xzf "$PKG_DIR/python-3.13.0-linux-${PKG_ARCH}.tar.gz" -C /
fi

echo "📦 Installing Bazel..."
sudo tar -xzf "$PKG_DIR/bazel-6.5.0-linux-${PKG_ARCH}.tar.gz" -C /usr/local/bin/
sudo chmod +x /usr/local/bin/bazel

echo "📦 Installing Ninja..."
sudo tar -xzf "$PKG_DIR/ninja-1.12.0-linux-${PKG_ARCH}.tar.gz" -C /

echo "📦 Installing Node.js..."
sudo tar -xzf "$PKG_DIR/nodejs-20.19.0-linux-${PKG_ARCH}.tar.gz" -C "$BUILD_TOOLS"

# --- Python symlinks ---
for ver in "3.9" "3.10" "3.11" "3.12" "3.13"; do
    sudo ln -sf "$BUILD_TOOLS/python$ver/bin/python$ver" "/usr/local/bin/python$ver"
    sudo ln -sf "$BUILD_TOOLS/python$ver/bin/python$ver" "$BUILD_TOOLS/python$ver/bin/python"
    sudo ln -sf "$BUILD_TOOLS/python$ver/bin/pip$ver" "/usr/local/bin/pip$ver"
done
sudo ln -sf /usr/local/bin/python3.9 /usr/bin/python3 2>/dev/null || true
sudo ln -sf /usr/bin/python3 /usr/bin/python 2>/dev/null || true

# --- Dynamic linker config ---
echo "/opt/buildtools/python3.9/lib" | sudo tee /etc/ld.so.conf.d/python3.9.conf > /dev/null
echo "/opt/buildtools/python3.10/lib" | sudo tee /etc/ld.so.conf.d/python3.10.conf > /dev/null
echo "/opt/buildtools/python3.11/lib" | sudo tee /etc/ld.so.conf.d/python3.11.conf > /dev/null
echo "/opt/buildtools/python3.12/lib" | sudo tee /etc/ld.so.conf.d/python3.12.conf > /dev/null
echo "/opt/buildtools/python3.13/lib" | sudo tee /etc/ld.so.conf.d/python3.13.conf > /dev/null
sudo ldconfig

# --- Global environment setup ---
sudo tee /etc/profile.d/buildtools.sh > /dev/null << 'EOF'
export JAVA_HOME=/opt/buildtools/jdk8
export MAVEN_HOME=/opt/buildtools/apache-maven-3.9.11
export GOROOT=/opt/buildtools/golang_go-1.24.1
export GOPATH=/opt/buildtools/go_workspace
export PATH=$JAVA_HOME/bin:$MAVEN_HOME/bin:$GOROOT/bin:$GOPATH/bin:/opt/buildtools/nodejs/bin:$PATH
export GO111MODULE=on
export GOPROXY=https://mirrors.huaweicloud.com/repository/goproxy/
export GONOSUMDB=*

export PYTHON_PATH_3911=/opt/buildtools/python3.9
export PYTHON_PATH_3102=/opt/buildtools/python3.10
export PYTHON_PATH_3114=/opt/buildtools/python3.11
export PYTHON_PATH_3120=/opt/buildtools/python3.12
export PYTHON_PATH_3130=/opt/buildtools/python3.13
export PATH=$PYTHON_PATH_3911/bin:$PYTHON_PATH_3102/bin:$PYTHON_PATH_3114/bin:$PYTHON_PATH_3120/bin:$PYTHON_PATH_3130/bin:$HOME/.local/bin:$PATH
export LD_LIBRARY_PATH=$PYTHON_PATH_3911/lib:$PYTHON_PATH_3102/lib:$PYTHON_PATH_3114/lib:$PYTHON_PATH_3120/lib:$PYTHON_PATH_3130/lib:/usr/lib:/usr/lib64:$LD_LIBRARY_PATH

export LANG=C.UTF-8
export TMOUT=0
EOF


# --- Configure pip mirrors ---
for py in "3.9" "3.10" "3.11" "3.12" "3.13"; do
    pip_cmd="pip$(echo $py | tr -d .)"
    if command -v "$pip_cmd" &> /dev/null; then
        "$pip_cmd" config --user set global.index-url https://mirrors.huaweicloud.com/repository/pypi/simple
        "$pip_cmd" config --user set global.trusted-host mirrors.huaweicloud.com
    fi
done



# ==============================================================================
# Configure pip to use Huawei Cloud mirror
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

if command -v pip3.12 &>/dev/null; then
    pip3.12 config --user set global.index-url https://mirrors.huaweicloud.com/repository/pypi/simple
    pip3.12 config --user set global.trusted-host mirrors.huaweicloud.com
    pip3.12 install -U wheel
fi

if command -v pip3.13 &>/dev/null; then
    pip3.13 config --user set global.index-url https://mirrors.huaweicloud.com/repository/pypi/simple
    pip3.13 config --user set global.trusted-host mirrors.huaweicloud.com
    pip3.13 install -U wheel
fi

# --- Configure npm ---
export PATH="/opt/buildtools/nodejs/bin:$PATH"
if command -v npm &> /dev/null; then
    npm config set registry https://mirrors.huaweicloud.com/repository/npm/
    npm config set strict-ssl false
fi

sudo chown -R "$(whoami):$(whoami)" /opt/buildtools

echo
echo "🎉 Installation completed successfully on $ARCH!"
echo "👉 Run this to activate environment:"
echo "   source /etc/profile.d/buildtools.sh"
echo
echo "✅ Verify your tools:"
echo "   java -version        # Should show JDK 8"
echo "   mvn -v               # Maven 3.9.11"
echo "   go version           # Go 1.24.1"
echo "   python3.9 --version  # Python 3.9.11"
echo "   node -v              # Node.js 20.19.0"
echo "   bazel --version      # Bazel 6.5.0"
echo "   ninja --version      # Ninja 1.12.0"