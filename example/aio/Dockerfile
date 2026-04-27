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

# 设置 Python 版本参数 (支持 3.9 或 3.11)
ARG PYTHON_VERSION=3.9.18
ARG PYTHON_MAJOR_MINOR=3.9

# 配置国内镜像源（清华大学）
RUN sed -i 's|http://archive.ubuntu.com/ubuntu/|http://mirrors.tuna.tsinghua.edu.cn/ubuntu/|g' /etc/apt/sources.list && \
    sed -i 's|http://security.ubuntu.com/ubuntu/|http://mirrors.tuna.tsinghua.edu.cn/ubuntu/|g' /etc/apt/sources.list

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
ARG PYTHON_MAJOR_MINOR=3.9

# 配置国内镜像源（清华大学）
RUN sed -i 's|http://archive.ubuntu.com/ubuntu/|http://mirrors.tuna.tsinghua.edu.cn/ubuntu/|g' /etc/apt/sources.list && \
    sed -i 's|http://security.ubuntu.com/ubuntu/|http://mirrors.tuna.tsinghua.edu.cn/ubuntu/|g' /etc/apt/sources.list

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
    curl \
    wget \
    git \
    vim \
    supervisor \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /etc/supervisor/conf.d /var/log/supervisor

# 复制进程管理脚本
COPY --chown=root:root process-manager.sh /openyuanrong/process-manager.sh
RUN chmod +x /openyuanrong/process-manager.sh

# Install docker CLI (optional - if download fails, continue without it)
RUN install -m 0755 -d /etc/apt/keyrings \
    && (curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc \
    && chmod a+r /etc/apt/keyrings/docker.asc \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
    $(. /etc/os-release && echo "$UBUNTU_CODENAME") stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null \
    && apt-get update \
    && apt-get install -y --no-install-recommends docker-ce-cli \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*) || echo "Docker CLI installation failed - continuing without it"

# 安装 Node.js LTS（可选 - 用于 claude code / openclaw）
RUN (curl -fsSL https://deb.nodesource.com/setup_lts.x | bash - && \
    apt-get install -y nodejs && \
    rm -rf /var/lib/apt/lists/* && \
    npm config set registry https://registry.npmmirror.com && \
    npm install -g @anthropic-ai/claude-code openclaw) || echo "Node.js installation failed - continuing without it"

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

# 使用 BuildKit mount 安装 openyuanrong_sdk（不增加镜像层大小）
RUN --mount=type=bind,source=pkg/openyuanrong_sdk-0.7.0.dev0-cp39-cp39-linux_x86_64.whl,target=/tmp/openyuanrong_sdk-0.7.0.dev0-cp39-cp39-linux_x86_64.whl \
    if [ "${PYTHON_MAJOR_MINOR}" = "3.9" ]; then \
        pip install --no-cache-dir /tmp/openyuanrong_sdk-0.7.0.dev0-cp39-cp39-linux_x86_64.whl; \
    else \
        echo "Unsupported Python version: ${PYTHON_MAJOR_MINOR}"; \
        exit 1; \
    fi

RUN --mount=type=bind,source=pkg/openyuanrong-0.7.0.dev0-cp39-cp39-linux_x86_64.whl,target=/tmp/openyuanrong-0.7.0.dev0-cp39-cp39-linux_x86_64.whl \
    if [ "${PYTHON_MAJOR_MINOR}" = "3.9" ]; then \
        pip install --no-cache-dir /tmp/openyuanrong-0.7.0.dev0-cp39-cp39-linux_x86_64.whl; \
    else \
        echo "Unsupported Python version: ${PYTHON_MAJOR_MINOR}"; \
        exit 1; \
    fi

COPY --chown=root:root bootstrap /openyuanrong/
COPY --chown=root:root services.yaml /openyuanrong/
COPY --chown=root:root cert /openyuanrong/cert/
COPY --chown=root:root pkg/runtime-launcher /openyuanrong/
COPY --chown=root:root traefik /openyuanrong/traefik/
COPY --chown=root:root supervisord.conf /etc/supervisor/supervisord.conf
COPY --chown=root:root supervisord-entrypoint.sh /usr/local/bin/supervisord-entrypoint.sh
RUN chmod +x /usr/local/bin/supervisord-entrypoint.sh

# 复制文档到镜像
COPY --chown=root:root TRAEFIK_ETCD.md /openyuanrong/docs/

# 安装 traefik - 从 GitHub 下载
ARG TRAEFIK_VERSION=v3.1.2
RUN echo "Downloading traefik from GitHub..." && \
    wget -qO /tmp/traefik.tar.gz https://github.com/traefik/traefik/releases/download/${TRAEFIK_VERSION}/traefik_${TRAEFIK_VERSION}_linux_amd64.tar.gz || \
    curl -sL -o /tmp/traefik.tar.gz https://github.com/traefik/traefik/releases/download/${TRAEFIK_VERSION}/traefik_${TRAEFIK_VERSION}_linux_amd64.tar.gz && \
    tar -xzf /tmp/traefik.tar.gz -C /tmp traefik && \
    install -m 0755 /tmp/traefik /usr/local/bin/traefik && \
    rm -f /tmp/traefik /tmp/traefik.tar.gz && \
    traefik version

WORKDIR /root

ENV DEBIAN_FRONTEND=dialog

ENV YR_SERVER_ADDRESS="127.0.0.1:8888"
ENV YR_IAM_ADDRESS="127.0.0.1:31112"
ENV YR_CLIENT_AUTH_TYPE="one-way"
ENV YR_SERVER_NAME="openyuanrong.org"
ENV YR_ENABLE_TLS="true"
ENV YR_PRIVATE_KEY_FILE="/openyuanrong/cert/module.key"
ENV YR_CERT_FILE="/openyuanrong/cert/module.crt"
ENV YR_VERIFY_FILE="/openyuanrong/cert/ca.crt"
ENV YR_IN_CLUSTER="false"
ENV ENABLE_TOKEN="false"

# 使用 supervisord 作为入口
ENTRYPOINT ["/usr/bin/tini", "--", "/usr/local/bin/supervisord-entrypoint.sh"]