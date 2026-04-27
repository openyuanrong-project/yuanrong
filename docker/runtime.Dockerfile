# ============================================
# 使用 BuildKit 特性优化镜像大小
# 构建时请确保启用 BuildKit: DOCKER_BUILDKIT=1 docker build ...
# ============================================

# ============================================
# 第一阶段: 构建 Python (源码编译)
# ============================================
FROM ubuntu:22.04 AS python-builder

# 设置非交互模式
ENV DEBIAN_FRONTEND=noninteractive

# 设置 Python 版本参数 (支持 3.11 或 3.11)
ARG PYTHON_VERSION=3.11.4
ARG PYTHON_MAJOR_MINOR=3.11

# 安装编译 Python 所需的依赖
RUN apt-get update && apt-get install -y \
    wget \
    build-essential \
    zlib1g-dev \
    libncurses5-dev \
    libgdbm-dev \
    libnss3-dev \
    libssl-dev \
    libreadline-dev \
    libffi-dev \
    libsqlite3-dev \
    libbz2-dev \
    liblzma-dev \
    && rm -rf /var/lib/apt/lists/*

# 下载并编译 Python（使用清华镜像加速）
RUN wget --timeout=30 --tries=3 \
        https://repo.huaweicloud.com/python/${PYTHON_VERSION}/Python-${PYTHON_VERSION}.tgz \
        -O /tmp/python.tgz || \
    wget --timeout=30 --tries=3 \
        https://mirrors.tuna.tsinghua.edu.cn/python/${PYTHON_VERSION}/Python-${PYTHON_VERSION}.tgz \
        -O /tmp/python.tgz || \
    wget --timeout=30 --tries=3 \
        https://www.python.org/ftp/python/${PYTHON_VERSION}/Python-${PYTHON_VERSION}.tgz \
        -O /tmp/python.tgz && \
    tar -xzf /tmp/python.tgz -C /tmp && \
    cd /tmp/Python-${PYTHON_VERSION} && \
    ./configure \
        --prefix=/usr/local \
        --enable-shared \
        --without-ensurepip \
        LDFLAGS="-Wl,-rpath=/usr/local/lib" && \
    make -j$(nproc) && \
    make install && \
    cd / && \
    rm -rf /tmp/Python-${PYTHON_VERSION} /tmp/python.tgz

# 下载并安装 pip（使用清华镜像）
RUN wget --timeout=30 --tries=3 \
        https://bootstrap.pypa.io/get-pip.py -O /tmp/get-pip.py && \
    /usr/local/bin/python${PYTHON_MAJOR_MINOR} /tmp/get-pip.py -i https://pypi.tuna.tsinghua.edu.cn/simple && \
    rm /tmp/get-pip.py

# 清理不必要的文件以减小镜像大小
RUN find /usr/local -type d -name '__pycache__' -exec rm -rf {} + 2>/dev/null || true && \
    find /usr/local -type d -name 'test' -exec rm -rf {} + 2>/dev/null || true && \
    find /usr/local -type d -name 'tests' -exec rm -rf {} + 2>/dev/null || true && \
    find /usr/local -type f -name '*.pyc' -delete && \
    find /usr/local -type f -name '*.pyo' -delete

# ============================================
# 第二阶段: 最终镜像 (Ubuntu 22.04 + Python + openyuanrong_sdk)
# ============================================
FROM ubuntu:22.04

# 设置非交互模式
ENV DEBIAN_FRONTEND=noninteractive

# 设置 Python 版本参数
ARG PYTHON_MAJOR_MINOR=3.11

# 安装运行时依赖（最小化）
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libssl3 \
    libsqlite3-0 \
    libreadline8 \
    libffi8 \
    libbz2-1.0 \
    liblzma5 \
    tini \
    && rm -rf /var/lib/apt/lists/*

# 从构建阶段复制 Python
COPY --from=python-builder /usr/local /usr/local

# 设置环境变量
ENV PATH="/usr/local/bin:$PATH" \
    LD_LIBRARY_PATH="/usr/local/lib" \
    PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1

# 配置 pip 使用清华镜像源（解决网络超时问题）
RUN pip${PYTHON_MAJOR_MINOR} config set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple && \
    pip${PYTHON_MAJOR_MINOR} config set global.timeout 100 && \
    ln -sf /usr/local/bin/python${PYTHON_MAJOR_MINOR} /usr/local/bin/python && \
    ln -sf /usr/local/bin/pip${PYTHON_MAJOR_MINOR} /usr/local/bin/pip

# 验证 Python 版本
RUN python --version

# 使用 BuildKit mount 安装 openyuanrong_sdk（不增加镜像层大小）
RUN --mount=type=bind,source=openyuanrong_sdk-0.7.0.dev0-cp311-cp311-linux_x86_64.whl,target=/tmp/openyuanrong_sdk-0.7.0.dev0-cp311-cp311-linux_x86_64.whl \
    if [ "${PYTHON_MAJOR_MINOR}" = "3.11" ]; then \
        pip install --no-cache-dir /tmp/openyuanrong_sdk-0.7.0.dev0-cp311-cp311-linux_x86_64.whl; \
    else \
        echo "Unsupported Python version: ${PYTHON_MAJOR_MINOR}"; \
        exit 1; \
    fi

# 验证安装
RUN python -c "import sys; print(f'Python version: {sys.version}')" && \
    pip list | grep openyuanrong || true

RUN echo '#!/bin/bash' > /home/entryfile.sh && \
    echo '' >> /home/entryfile.sh && \
    echo 'cd ${YR_RT_WORKING_DIR}' >> /home/entryfile.sh && \
    echo 'export LD_LIBRARY_PATH=./usr/lib/x86_64-linux-gnu/:./usr/local/lib:./usr/lib64:./lib' >> /home/entryfile.sh && \
    echo 'export PYTHONPATH=./usr/local/lib/python3.11/site-packages/' >> /home/entryfile.sh && \
    echo 'exec ./usr/lib/x86_64-linux-gnu/ld-linux-x86-64.so.2 \' >> /home/entryfile.sh && \
    echo '    ./usr/local/bin/python3 \' >> /home/entryfile.sh && \
    echo '    ./usr/local/lib/python3.11/site-packages/yr/main/yr_runtime_main.py \' >> /home/entryfile.sh && \
    echo '    $@' >> /home/entryfile.sh && \
    chmod +x /home/entryfile.sh