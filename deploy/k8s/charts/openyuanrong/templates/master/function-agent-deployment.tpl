{{- define "agent.deployment.template" }}
metadata:
  labels:
    app: function-agent
  name: function-agent
  namespace: {{ .Values.global.namespace }}
spec:
  progressDeadlineSeconds: 600
  replicas: {{ .Values.global.pool.poolSize }}
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: function-agent
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
  {{ include "agent.pod.template" . }}
{{- end }}

{{- define "agent.pod.template" }}
    metadata:
      creationTimestamp: null
      labels:
        app: function-agent
      annotations:
        yr-default: yr-default
    spec:
      {{- if .Values.global.pool.nodeAffinity }}
      affinity:
        {{- with .Values.global.pool.nodeAffinity }}
        nodeAffinity:
          {{- toYaml . | nindent 10 }}
        {{- end }}
      {{- end }}
      {{- if .Values.global.pool.accelerator }}
      nodeSelector:
        accelerator: {{ .Values.global.pool.accelerator }}
      {{- else if .Values.global.pool.nodeSelector }}
      {{- with .Values.global.pool.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- end }}
      {{- if .Values.global.controlPlane.tolerations }}
      {{- with .Values.global.controlPlane.tolerations }}
      tolerations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- end }}
      initContainers:
      - name: function-agent-init
        command:
          - /home/sn/bin/entrypoint-function-agent-init
        image: "{{ .Values.global.imageRegistry }}{{ .Values.global.images.agentInit }}"
        securityContext:
          runAsUser: 0
          capabilities:
            drop:
            - ALL
            add: # Add as needed based on the script entrypoint-function-agent-init.
            - NET_RAW
            - NET_ADMIN
            - SYS_ADMIN
            - CHOWN
            - SETGID
            - SETUID
            - DAC_OVERRIDE
            - FOWNER
            - FSETID
        env:
          - name: POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
          - name: ENABLE_IPV4_TENANT_ISOLATION
            value: "{{ .Values.global.tenantIsolation.ipv4.enable }}"
          - name: HOST_IP
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: status.hostIP
          - name: THIRD_PARTY_WHITELIST
            value: "{{ .Values.global.tenantIsolation.thirdPartyWhitelist }}"
          - name: SVC_CIDR
            value: "{{ .Values.global.kubernetes.svcCIDR }}"
          - name: POD_CIDR
            value: "{{ .Values.global.kubernetes.podCIDR }}"
          - name: HOST_CIDR
            value: "{{ .Values.global.kubernetes.hostCIDR }}"
          - name: TCP_PORT_WHITELIST
            value: "{{ .Values.global.tenantIsolation.ipv4.tcpPortWhitelist }}"
          - name: UDP_PORT_WHITELIST
            value: "{{ .Values.global.tenantIsolation.ipv4.udpPortWhitelist }}"
        volumeMounts:
          {{- if .Values.global.log.hostPath.enable }}
          - mountPath: "{{ .Values.global.log.functionSystem.path }}"
            name: varlog-runtime-manager
            subPathExpr: $(POD_NAME)
          - mountPath: "{{ .Values.global.log.runtime.path }}"
            name: servicelog
            subPathExpr: $(POD_NAME)
          - mountPath: "{{ .Values.global.log.userOutput.path }}"
            name: stdlog
            subPathExpr: $(POD_NAME)
          {{- else }}
          - mountPath: "{{ .Values.global.log.functionSystem.path }}"
            name: varlog-runtime-manager
          - mountPath: "{{ .Values.global.log.runtime.path }}"
            name: servicelog
          - mountPath: "{{ .Values.global.log.userOutput.path }}"
            name: stdlog
          {{- end }}
          - mountPath: /home/snuser/secret
            name: secret-dir
          - mountPath: /dcache
            name: pkg-dir
          - mountPath: /opt/function/code
            name: pkg-dir1
          {{- if .Values.global.sts.enable }}
          - mountPath: /home/snuser/alarms
            name: alarms-dir
          {{- end }}
          - mountPath: {{ .Values.global.observer.metrics.path.file }}
            name: metrics-dir
          - mountPath: /home/snuser/metrics
            name: runtime-metrics-dir
          - mountPath: {{ .Values.global.observer.metrics.path.failure }}
            name: varfailuremetrics
      containers:
      - name: runtime-manager
        command:
          - /home/sn/bin/entrypoint-runtime-manager
        {{- if and .Values.global.pool.accelerator (ne .Values.global.pool.accelerator "nvidia-gpu") (ne .Values.global.pool.accelerator "amd-gpu") }}
        envFrom:
        - configMapRef:
            name: function-agent-config
        {{- end }}
        env:
          - name: POD_IP
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: status.podIP
          - name: HOST_IP
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: status.hostIP
          - name: NODE_ID
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: spec.nodeName
          - name: POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
          - name: POD_ID
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.uid
          - name: RUNTIME_MGR_PORT
            value: {{ quote .Values.global.port.runtimeMgrPort }}
          - name: ENABLE_INHERIT_ENV
            value: "false"
          - name: FUNCTION_AGENT_PORT
            value: {{ quote .Values.global.port.functionAgentPort }}
          - name: RUNTIME_INIT_PORT
            value: {{ quote .Values.global.port.runtimeInitPort }}
          - name: DS_WORKER_PORT
            value: {{ quote .Values.global.port.worker }}
          - name: FUNCTION_PROXY_GRPC_PORT
            value: {{ quote .Values.global.port.functionProxyGrpcPort }}
          - name: RUNTIME_PORT_NUM
            value: {{ quote .Values.global.port.runtimePortNum }}
          - name: METRICS_COLLECTOR_TYPE
            value: {{ .Values.global.runtime.metricsCollectorType }}
          - name: DISK_USAGE_MONITOR_NOTIFY_FAILURE_ENABLE
            value: {{ quote .Values.global.runtime.diskUsageMonitor.notifyFailureEnable }}
          - name: DISK_USAGE_MONITOR_PATH
            value: {{ quote .Values.global.runtime.diskUsageMonitor.path }}
          - name: DISK_USAGE_LIMIT
            value: {{ quote .Values.global.runtime.diskUsageMonitor.limit }}
          - name: SNUSER_LIB_PATH
            value: {{ quote .Values.global.runtime.snuserLibPath }}
          - name: VIRTUAL_ENV_IDLE_TIME_LIMIT
            value: {{ quote .Values.global.runtime.virtualEnvIdleTimeLimit }}
          - name: REQUEST_ACK_ACC_MAX_SEC
            value: {{ quote .Values.global.runtime.requestAckAccMaxSec}}
          - name: SNUSER_DIR_DISK_USAGE_LIMIT
            value: {{ quote .Values.global.runtime.diskUsageMonitor.snuserDirSizeLimit }}
          - name: TMP_DIR_DISK_USAGE_LIMIT
            value: {{ quote .Values.global.runtime.diskUsageMonitor.tmpDirSizeLimit }}
          - name: DISK_USAGE_MONITOR_DURATION
            value: {{ quote .Values.global.runtime.diskUsageMonitor.duration }}
          - name: CPU4COMP
            value: {{ quote .Values.global.pool.requestCpu }}
          - name: MEM4COMP
            value: {{ quote .Values.global.pool.requestMemory }}
          - name: INIT_LABELS
            value: ""
          - name: LOG_PATH
            value: {{ .Values.global.log.functionSystem.path }}
          - name: LOG_LEVEL
            value: {{ .Values.global.log.functionSystem.level }}
          - name: LOG_PATTERN
            value: {{ quote .Values.global.log.functionSystem.pattern }}
          - name: LOG_COMPRESS_ENABLE
            value: {{ quote .Values.global.log.functionSystem.compress }}
          - name: LOG_ROLLING_MAXSIZE
            value: {{ quote .Values.global.log.functionSystem.rolling.maxSize }}
          - name: LOG_ROLLING_MAXFILES
            value: {{ quote .Values.global.log.functionSystem.rolling.maxfiles }}
          - name: LOG_ASYNC_LOGBUFSECS
            value: "10"
          - name: LOG_ASYNC_MAXQUEUESIZE
            value: "1024"
          - name: LOG_ASYNC_THREADCOUNT
            value: "1"
          - name: LOG_ALSOLOGTOSTDERR
            value: "false"
          - name: ENABLE_METRICS
            value: {{ quote .Values.global.observer.metrics.enable }}
          - name: METRICS_CONFIG
            value: {{ quote .Values.global.observer.metrics.metricsConfig }}
          - name: METRICS_CONFIG_FILE
            value: {{ quote .Values.global.observer.metrics.metricsConfigFile }}
          - name: RUNTIME_METRICS_CONFIG
            value: {{ quote .Values.global.observer.metrics.runtimeMetricsConfig }}
          - name: RUNTIME_METRICS_CONFIG_FILE
            value: {{ quote .Values.global.observer.metrics.runtimeMetricsConfigFile }}
          - name: ENABLE_TRACE
            value: {{ quote .Values.global.observer.trace.enable }}
          - name: TRACE_CONFIG
            value: {{ quote .Values.global.observer.trace.traceConfig }}
          - name: RUNTIME_TRACE_CONFIG
            value: {{ quote .Values.global.observer.trace.runtimeTraceConfig }}
          - name: RUNTIME_LOG_DIR
            value: {{ .Values.global.log.runtime.path }}
          - name: RUNTIME_LOG_LEVEL
            value: {{ .Values.global.log.runtime.level }}
          - name: PROMETHEUS_PUSH_GATEWAY_IP
            value: {{ quote .Values.global.observer.proGatewayIP }}
          - name: PROMETHEUS_PUSH_GATEWAY_PORT
            value: {{ quote .Values.global.observer.gatewayPort }}
          - name: JAVA_PRESTART_COUNT
            value: {{ quote .Values.global.runtime.prestartCount.java8 }}
          - name: JAVA11_PRESTART_COUNT
            value: {{ quote .Values.global.runtime.prestartCount.java11 }}
          - name: PYTHON36_PRESTART_COUNT
            value: {{ quote .Values.global.runtime.prestartCount.python36 }}
          - name: PYTHON37_PRESTART_COUNT
            value: {{ quote .Values.global.runtime.prestartCount.python37 }}
          - name: PYTHON38_PRESTART_COUNT
            value: {{ quote .Values.global.runtime.prestartCount.python38 }}
          - name: PYTHON39_PRESTART_COUNT
            value: {{ quote .Values.global.runtime.prestartCount.python39 }}
          - name: PYTHON310_PRESTART_COUNT
            value: {{ quote .Values.global.runtime.prestartCount.python310 }}
          - name: PYTHON311_PRESTART_COUNT
            value: {{ quote .Values.global.runtime.prestartCount.python311 }}
          - name: CPP_PRESTART_COUNT
            value: {{ quote .Values.global.runtime.prestartCount.cpp }}
          - name: JVM_CUSTOM_ARGS
            value: {{ quote .Values.global.runtime.jvmCustomArgs }}
          - name: JAVA8_DEFAULT_ARGS
            value: {{ quote .Values.global.runtime.defaultArgs.java8 }}
          - name: JAVA11_DEFAULT_ARGS
            value: {{ quote .Values.global.runtime.defaultArgs.java11 }}
          - name: JAVA17_DEFAULT_ARGS
            value: {{ quote .Values.global.runtime.defaultArgs.java17 }}
          - name: JAVA21_DEFAULT_ARGS
            value: {{ quote .Values.global.runtime.defaultArgs.java21 }}
          - name: SYSTEM_TIMEOUT
            value: {{ quote .Values.global.common.systemTimeout }}
          - name: CLUSTER_ID
            value: {{ quote .Values.global.clusterId }}
          - name: RUNTIME_GID
            value: "1003"
          - name: RUNTIME_UID
            value: "1003"
          - name: ENABLE_DS_CLIENT
            value: "0"
          - name: NPU_COLLECTION_MODE
            value: {{ quote .Values.global.runtime.npuCollectionMode }}
          - name: GPU_COLLECTION_ENABLE
            value: {{ quote .Values.global.runtime.gpuCollectionEnable }}
          - name: IS_PROTOMSG_TO_RUNTIME
            value: {{ quote .Values.global.runtime.isProtoMsgToRuntime }}
          - name: MASSIF_ENABLE
            value: {{ quote .Values.global.runtime.massifEnable }}
          - name: ENABLE_DISTRIBUTED_CONVERGENT_CALL_STACK
            value: {{ quote .Values.global.runtime.disConvCallStack.enable }}
          - name: RESOURCE_PATH
            value: /home/sn/resource
          - name: RUNTIME_HOME_DIR
            value: /home/snuser
          - name: RESOURCE_LABEL_PATH
            value: /home/sn/podInfo/labels
          - name: NPU_DEVICE_INFO_PATH
            value: /home/sn/config/topology-info.json
          - name: RUNTIME_DS_CONNECT_TIMEOUT
            value: {{ quote .Values.global.runtime.runtimeDsConnectTimeout }}
          {{- if .Values.global.log.runtime.expiration.enable }}
          - name: LOG_EXPIRATION_ENABLE
            value: {{ quote .Values.global.log.runtime.expiration.enable }}
          - name: LOG_EXPIRATION_CLEANUP_INTERVAL
            value: {{ quote .Values.global.log.runtime.expiration.cleanupInterval }}
          - name: LOG_EXPIRATION_TIME_THRESHOLD
            value: {{ quote .Values.global.log.runtime.expiration.timeThreshold }}
          - name: LOG_EXPIRATION_MAX_FILE_COUNT
            value: {{ quote .Values.global.log.runtime.expiration.maxFileCount }}
          - name: LOG_REUSE_ENABLE
            value: {{ quote .Values.global.log.runtime.expiration.logReuseEnable }}
          {{- end }}
          - name: USER_LOG_EXPORT_MODE
            value: {{ quote .Values.global.log.runtime.userLogExportMode }}
          - name: RUNTIME_DIRECT_CONNECTION_ENABLE
            value: "false"
          - name: ENABLE_CLEAN_STREAM_PRODUCER
            value: {{ quote .Values.global.runtime.cleanStreamProducerEnable }}
          {{- if .Values.global.runtime.oomKill.enable }}
          - name: OOM_KILL_ENABLE
            value: {{ quote .Values.global.runtime.oomKill.enable }}
          - name: MEMORY_DETECTION_INTERVAL
            value: {{ quote .Values.global.runtime.oomKill.memoryDetectionInterval }}
          - name: OOM_CONSECUTIVE_DETECTION_COUNT
            value: {{ quote .Values.global.runtime.oomKill.consecutiveDetectionCount }}
          - name: OOM_KILL_CONTROL_LIMIT
            value: {{ quote .Values.global.runtime.oomKill.controlLimit }}
          {{- end }}
          - name: KILL_PROCESS_TIMEOUT_SECONDS
            value: {{ quote .Values.global.runtime.killProcessTimeoutSeconds }}
          - name: RUNTIME_INSTANCE_DEBUG_ENABLE
            value: "false"
        image: "{{ .Values.global.imageRegistry | trimSuffix "/" }}/{{ .Values.global.images.runtimeManager }}"
        imagePullPolicy: IfNotPresent
        livenessProbe:
          failureThreshold: {{ .Values.global.pool.readinessProbeFailureThreshold }}
          exec:
            command:
              - /bin/bash
              - -c
              - /home/sn/bin/health-check $(RUNTIME_MGR_PORT) runtime-manager
          initialDelaySeconds: 1
          periodSeconds: 5
          successThreshold: 1
          timeoutSeconds: 5
        readinessProbe:
          failureThreshold: {{ .Values.global.pool.livenessProbeFailureThreshold }}
          exec:
            command:
              - /bin/bash
              - -c
              - /home/sn/bin/health-check $(RUNTIME_MGR_PORT) runtime-manager
          initialDelaySeconds: 1
          periodSeconds: 1
          successThreshold: 1
          timeoutSeconds: 5
        ports:
          - containerPort: 21005
            name: 21005tcp00
            protocol: TCP
          - name: prometheus-http
            containerPort: 9392
            protocol: TCP
        resources:
          limits:
            {{- if eq .Values.global.pool.accelerator "huawei-Ascend310" }}
            huawei.com/Ascend310: {{ .Values.global.pool.cardNum }}
            {{ else if eq .Values.global.pool.accelerator "huawei-Ascend310P" }}
            huawei.com/Ascend310P: {{ .Values.global.pool.cardNum }}
            {{ else if eq .Values.global.pool.accelerator "huawei-Ascend910" }}
            huawei.com/Ascend910: {{ .Values.global.pool.cardNum }}
            {{ else if eq .Values.global.pool.accelerator "huawei-Ascend910B" }}
            huawei.com/ascend-1980: {{ .Values.global.pool.cardNum }}
            {{ else if eq .Values.global.pool.accelerator "nvidia-gpu" }}
            nvidia.com/gpu: {{ .Values.global.pool.cardNum }}
            {{ else if eq .Values.global.pool.accelerator "amd-gpu" }}
            amd.com/gpu: {{ .Values.global.pool.cardNum }}
            {{- end }}
            cpu: {{ .Values.global.pool.limitCpu }}m
            memory: {{ .Values.global.pool.limitMemory }}Mi
            ephemeral-storage: {{ .Values.global.pool.limitEphemeralStorage }}Mi
          requests:
            {{- if eq .Values.global.pool.accelerator "huawei-Ascend310" }}
            huawei.com/Ascend310: {{ .Values.global.pool.cardNum }}
            {{ else if eq .Values.global.pool.accelerator "huawei-Ascend310P" }}
            huawei.com/Ascend310P: {{ .Values.global.pool.cardNum }}
            {{ else if eq .Values.global.pool.accelerator "huawei-Ascend910" }}
            huawei.com/Ascend910: {{ .Values.global.pool.cardNum }}
            {{ else if eq .Values.global.pool.accelerator "huawei-Ascend910B" }}
            huawei.com/ascend-1980: {{ .Values.global.pool.cardNum }}
            {{ else if eq .Values.global.pool.accelerator "nvidia-gpu" }}
            nvidia.com/gpu: {{ .Values.global.pool.cardNum }}
            {{ else if eq .Values.global.pool.accelerator "amd-gpu" }}
            amd.com/gpu: {{ .Values.global.pool.cardNum }}
            {{- end }}
            cpu: {{ .Values.global.pool.requestCpu }}m
            memory: {{ .Values.global.pool.requestMemory }}Mi
            ephemeral-storage: {{ .Values.global.pool.requestEphemeralStorage }}Mi
        securityContext:
          capabilities:
            add:
            - SYS_ADMIN
            - KILL
            - DAC_OVERRIDE
            - SETGID
            - SETUID
            drop:
            - ALL
        terminationMessagePath: /var/tmp/termination-log
        terminationMessagePolicy: File
        volumeMounts:
          - mountPath: /etc/localtime
            name: local-time
          {{- if .Values.global.log.hostPath.enable }}
          - mountPath: "{{ .Values.global.log.functionSystem.path }}"
            name: varlog-runtime-manager
            subPathExpr: $(POD_NAME)
          - mountPath: "{{ .Values.global.log.runtime.path }}"
            name: servicelog
            subPathExpr: $(POD_NAME)
          - mountPath: "{{ .Values.global.log.userOutput.path }}"
            name: stdlog
            subPathExpr: $(POD_NAME)
          {{- else }}
          - mountPath: "{{ .Values.global.log.functionSystem.path }}"
            name: varlog-runtime-manager
          - mountPath: "{{ .Values.global.log.runtime.path }}"
            name: servicelog
          - mountPath: "{{ .Values.global.log.userOutput.path }}"
            name: stdlog
          {{- end }}
          - mountPath: /home/snuser/secret
            name: secret-dir
          - mountPath: /dcache
            name: pkg-dir
          - mountPath: /opt/function/code
            name: pkg-dir1
          {{- if .Values.global.sts.enable }}
          - mountPath: /home/snuser/alarms
            name: alarms-dir
          {{- end }}
          - mountPath: /home/sn/metrics
            name: metrics-dir
          - mountPath: /home/snuser/metrics
            name: runtime-metrics-dir
          - mountPath: /home/snuser/config/python-runtime-log.json
            name: python-runtime-log-config
            readOnly: true
            subPath: python-runtime-log.json
          - mountPath: /home/snuser/config/runtime.json
            name: runtime-config
            subPath: runtime.json
            readOnly: true
          - mountPath: /home/snuser/runtime/java/log4j2.xml
            name: java-runtime-log4j2-config
            subPath: log4j2.xml
          - mountPath: /home/uds
            name: datasystem-socket
          - mountPath: /dev/shm
            name: datasystem-shm
          - mountPath: /home/sn/podInfo
            name: podinfo
      - name: function-agent
        command:
          - /home/sn/bin/entrypoint-function-agent
        env:
        - name: POD_IP
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.podIP
        - name: NODE_ID
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: HOST_IP
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.hostIP
        - name: POD_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.name
        - name: FSPROXY_PORT
          value: {{ quote .Values.global.port.functionProxyPort }}
        - name: FUNCTION_AGENT_PORT
          value: {{ quote .Values.global.port.functionAgentPort }}
        - name: PROMETHEUS_PUSH_GATEWAY_IP
          value: {{ quote .Values.global.observer.proGatewayIP }}
        - name: PROMETHEUS_PUSH_GATEWAY_PORT
          value: {{ quote .Values.global.observer.gatewayPort }}
        - name: DECRYPT_ALGORITHM
          value: {{ quote .Values.global.common.decryptAlgorithm }}
        - name: S3_ADDR
          value: {{ .Values.global.obsManagement.s3Endpoint | default "$(HOST_IP):30110" }}
        - name: S3_PROTOCOL
          value: {{ .Values.global.obsManagement.protocol }}
        - name: S3_CREDENTIAL_TYPE
          value: {{ .Values.global.obsManagement.credentialType }}
        - name: LOG_PATH
          value: {{ .Values.global.log.functionSystem.path }}
        - name: LOG_LEVEL
          value: {{ .Values.global.log.functionSystem.level }}
        - name: LOG_PATTERN
          value: {{ quote .Values.global.log.functionSystem.pattern }}
        - name: LOG_COMPRESS_ENABLE
          value: {{ quote .Values.global.log.functionSystem.compress }}
        - name: LOG_ROLLING_MAXSIZE
          value: {{ quote .Values.global.log.functionSystem.rolling.maxSize }}
        - name: LOG_ROLLING_MAXFILES
          value: {{ quote .Values.global.log.functionSystem.rolling.maxfiles }}
        - name: LOG_ASYNC_LOGBUFSECS
          value: {{ quote .Values.global.log.functionSystem.async.logBufSecs }}
        - name: LOG_ASYNC_MAXQUEUESIZE
          value: "51200"
        - name: LOG_ASYNC_THREADCOUNT
          value: "1"
        - name: LOG_ALSOLOGTOSTDERR
          value: "false"
        - name: ENABLE_METRICS
          value: {{ quote .Values.global.observer.metrics.enable }}
        - name: METRICS_CONFIG
          value: {{ quote .Values.global.observer.metrics.metricsConfig }}
        - name: METRICS_CONFIG_FILE
          value: {{ quote .Values.global.observer.metrics.metricsConfigFile }}
        - name: SYSTEM_TIMEOUT
          value: {{ quote .Values.global.common.systemTimeout }}
        - name: STS_CONFIG
          value: "{}"
        - name: SSL_ENABLE
          value: {{ quote .Values.global.mutualSSLConfig.sslEnable }}
        - name: SSL_BASE_PATH
          value: {{ quote .Values.global.mutualSSLConfig.sslBasePath }}
        - name: SSL_ROOT_FILE
          value: "ca.crt"
        - name: SSL_CERT_FILE
          value: "module.crt"
        - name: SSL_KEY_FILE
          value: "module.key"
        - name: SSL_PWD_FILE
          value: "cert_pwd"
        - name: SSL_DECRYPT_TOOL
          value: {{ quote .Values.global.mutualSSLConfig.sslDecryptTool }}
        - name: S3_DOWNLOAD_MAXZIPSIZE
          value: {{ quote .Values.global.common.zipFile.zipFileSizeMax }}
        - name: S3_DOWNLOAD_MAXUNZIPSIZE
          value: {{ quote .Values.global.common.zipFile.unzipFileSizeMax }}
        - name: S3_DOWNLOAD_MAXFILECOUNT
          value: {{ quote .Values.global.common.zipFile.fileCountsMax }}
        - name: S3_DOWNLOAD_MAXDIRDEPTH
          value: {{ quote .Values.global.common.zipFile.dirDepthMax }}
        - name: ENABLE_IPV4_TENANT_ISOLATION
          value: {{ quote .Values.global.tenantIsolation.ipv4.enable }}
        - name: DEPLOY_DIR
          value: {{ quote .Values.global.common.deployDir }}
        - name: SCC_ENABLE
          value: {{ quote .Values.global.scc.enable }}
        - name: SCC_ALGORITHM
          value: {{ quote .Values.global.scc.algorithm }}
        - name: SCC_PRIMARY_FILE
          value: "primary.ks"
        - name: SCC_STANDBY_FILE
          value: "standby.ks"
        - name: ENABLE_SIGNATURE_VALIDATION
          value: {{ quote .Values.global.common.zipFile.signatureValidationEnable }}
        - name: CODE_AGING_TIME
          value: {{ quote .Values.global.common.codeAgingTime }}
        - name: SYSTEM_AUTH_MODE
          value: {{ quote .Values.global.common.systemAuthMode }}
        - name: CUSTOM_RESOURCES
          value: ""
        - name: RESOURCE_PATH
          value: /home/sn/resource
        - name: SCC_BASE_PATH
          value: /home/sn/resource/scc
        - name: SCC_LOG_PATH
          value: /home/sn/log
        - name: DELEGATE_ENV_VAR
          value: '{"DISABLE_APIG_FORMAT":"true"}'
        image: "{{ .Values.global.imageRegistry | trimSuffix "/" }}/{{ .Values.global.images.functionAgent }}"
        imagePullPolicy: IfNotPresent
        livenessProbe:
          failureThreshold: {{ .Values.global.pool.livenessProbeFailureThreshold }}
          exec:
            command:
              - /bin/bash
              - -c
              - /home/sn/bin/health-check $(FUNCTION_AGENT_PORT) function-agent
          initialDelaySeconds: 10
          periodSeconds: 5
          successThreshold: 1
          timeoutSeconds: 5
        readinessProbe:
          failureThreshold: {{ .Values.global.pool.readinessProbeFailureThreshold }}
          exec:
            command:
              - /bin/bash
              - -c
              - /home/sn/bin/health-check $(FUNCTION_AGENT_PORT) function-agent
          initialDelaySeconds: 3
          periodSeconds: 1
          successThreshold: 1
          timeoutSeconds: 5
        ports:
        - containerPort: 58866
          name: 58866tcp00
          protocol: TCP
        resources:
          limits:
            cpu: "{{ .Values.global.resources.functionAgent.limits.cpu }}"
            memory: {{ .Values.global.resources.functionAgent.limits.memory }}
          requests:
            cpu: "{{ .Values.global.resources.functionAgent.requests.cpu }}"
            memory: {{ .Values.global.resources.functionAgent.requests.memory }}
        securityContext:
          capabilities:
            add:
            - NET_ADMIN
            - NET_RAW
            drop:
            - ALL
        terminationMessagePath: /var/tmp/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        {{- if .Values.global.common.secretKeyEnable }}
        - name: localauth
          mountPath: /home/sn/resource/cipher
          readOnly: true
        {{- end }}
        - mountPath: /etc/localtime
          name: local-time
        - mountPath: "{{ .Values.global.log.functionSystem.path }}"
          name: varlog-function-agent
          {{- if .Values.global.log.hostPath.enable }}
          subPathExpr: $(POD_NAME)
          {{- end}}
        - mountPath: /dcache
          name: pkg-dir
        - mountPath: /opt/function/code
          name: pkg-dir1
        - mountPath: {{ .Values.global.observer.metrics.path.file }}
          name: metrics-dir
        - mountPath: {{ .Values.global.observer.metrics.path.failure }}
          name: varfailuremetrics
        {{- if or (eq .Values.global.mutualSSLConfig.sslEnable true) (eq .Values.global.observer.metrics.sslEnable true) }}
        - mountPath: {{ .Values.global.mutualSSLConfig.sslBasePath }}
          name: module
          readOnly: true
        {{- end }}
        {{- if .Values.global.scc.enable }}
        - name: scc-ks
          mountPath: /home/sn/resource/scc
          readOnly: true
        {{- end }}
      dnsPolicy: ClusterFirst
      imagePullSecrets:
{{- include "functionSystem.imagePullSecrets" . | nindent 6 }}
      restartPolicy: Always
      automountServiceAccountToken: false
      schedulerName: default-scheduler
      securityContext:
        fsGroup: {{ .Values.global.runtime.fsGroup }}
        supplementalGroups:
          - 1000
          - 1002
      terminationGracePeriodSeconds: {{ .Values.global.pool.gracePeriodSeconds }}
      volumes:
      - name: local-time
        hostPath:
          path: /etc/localtime
      {{- if .Values.global.log.hostPath.enable }}
      - name: varlog-function-agent
        hostPath:
          path: "{{ .Values.global.log.hostPath.componentLog }}"
          type: DirectoryOrCreate
      - name: varlog-runtime-manager
        hostPath:
          path: "{{ .Values.global.log.hostPath.componentLog }}"
          type: DirectoryOrCreate
      - name: servicelog
        hostPath:
          path: "{{ .Values.global.log.hostPath.serviceLog }}"
          type: DirectoryOrCreate
      - name: stdlog
        hostPath:
          path: "{{ .Values.global.log.hostPath.userLog }}"
          type: DirectoryOrCreate
      {{- else }}
      - name: varlog-function-agent
        emptyDir: {}
      - name: varlog-runtime-manager
        emptyDir: {}
      - name: servicelog
        emptyDir: {}
      - name: stdlog
        emptyDir: {}
      {{- end }}
      - name: secret-dir
        emptyDir: {}
      - name: pkg-dir
        emptyDir: {}
      - name: pkg-dir1
        emptyDir: {}
      {{- if .Values.global.sts.enable }}
      - emptyDir: {}
        name: alarms-dir
      {{- end }}
      - emptyDir: {}
        name: metrics-dir
      - emptyDir: {}
        name: runtime-metrics-dir
      {{- if .Values.global.common.secretKeyEnable }}
      - name: localauth
        secret:
          secretName: local-secret
      {{- end }}
      {{- if .Values.global.observer.metrics.hostPath.failureFileEnable }}
      - name: varfailuremetrics
        hostPath:
          path: "{{ .Values.global.observer.metrics.hostPath.failureMetrics }}"
          type: DirectoryOrCreate
      {{- else }}
      - emptyDir: {}
        name: varfailuremetrics
      {{- end }}
      - name: resource-volume
        emptyDir: {}
      - configMap:
          defaultMode: 0440
          items:
            - key: python-runtime-log.json
              path: python-runtime-log.json
          name: function-agent-config
        name: python-runtime-log-config
      - configMap:
          defaultMode: 0440
          items:
            - key: runtime.json
              path: runtime.json
          name: function-agent-config
        name: runtime-config
      - configMap:
          defaultMode: 0440
          items:
            - key: log4j2.xml
              path: log4j2.xml
          name: function-agent-config
        name: java-runtime-log4j2-config
      - configMap:
          defaultMode: 0440
          items:
            - key: iptabelsRule
              path: iptabelsRule
          name: function-agent-config
        name: iptables-rules
      - hostPath:
          path: /home/uds
          type: ""
        name: datasystem-socket
      - hostPath:
          path: /dev/shm
          type: ""
        name: datasystem-shm
      - downwardAPI:
          defaultMode: 420
          items:
            - fieldRef:
                apiVersion: v1
                fieldPath: metadata.labels
              path: labels
            - fieldRef:
                apiVersion: v1
                fieldPath: metadata.annotations
              path: annotations
        name: podinfo
      {{- if or (eq .Values.global.mutualSSLConfig.sslEnable true) (eq .Values.global.observer.metrics.sslEnable true) }}
      - name: module
        secret:
          defaultMode: 0440
          secretName: {{ .Values.global.mutualSSLConfig.secretName }}
      {{- end }}
      {{- if .Values.global.scc.enable }}
      - name: scc-ks
        secret:
          defaultMode: 0440
          secretName: {{ .Values.global.scc.secretName }}
      - configMap:
          defaultMode: 0440
          items:
            - key: CONFIG
              path: scc.conf
          name: fs-scc-configmap
        name: scc-config
      {{- end }}
{{- end }}