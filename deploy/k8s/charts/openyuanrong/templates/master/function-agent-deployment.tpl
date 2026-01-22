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
            - /bin/sh
            - -c
            - |
                if [ whoami != "${USER_NAME}" ]; then
                    if [ -w /etc/passwd ]; then
                      echo "${USER_NAME}:x:$(id -u):$(id -g):${USER_NAME} user:${HOME}:/sbin/nologin" >> /etc/passwd
                    fi
                fi
                umask 0027
                if [ -f "${SNHOME}"/bin/alias/runtime_manager_alias.sh ]; then
                    source "${SNHOME}"/bin/alias/runtime_manager_alias.sh
                fi
                [ ! -d {{ quote .Values.global.log.functionSystem.path}} ] && mkdir -p {{ quote .Values.global.log.functionSystem.path}}
                python3 -m yr.cli.main -v launch --inherit-env --env-subst NODE_ID,POD_IP,HOST_IP,CPU4COMP,MEM4COMP runtime_manager
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
          - name: CPU4COMP
            value: {{ quote .Values.global.pool.requestCpu }}
          - name: MEM4COMP
            value: {{ quote .Values.global.pool.requestMemory }}
        image: "{{ .Values.global.imageRegistry | trimSuffix "/" }}/{{ .Values.global.images.runtimeManager }}"
        imagePullPolicy: IfNotPresent
        livenessProbe:
          failureThreshold: {{ .Values.global.pool.readinessProbeFailureThreshold }}
          exec:
            command:
              - /bin/bash
              - -c
              - /home/sn/bin/health-check {{ quote .Values.global.port.runtimeMgrPort }} runtime-manager
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
              - /home/sn/bin/health-check {{ quote .Values.global.port.runtimeMgrPort }} runtime-manager
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
          - name: config-volume
            mountPath: /etc/yuanrong/config.toml
            subPath: config.toml
      - name: function-agent
        command:
          - /bin/sh
          - -c
          - |
              if [ whoami != "${USER_NAME}" ]; then
                if [ -w /etc/passwd ]; then
                  echo "${USER_NAME}:x:$(id -u):$(id -g):${USER_NAME} user:${HOME}:/sbin/nologin" >> /etc/passwd
                fi
              fi
              umask 0027
              [ ! -d "{{ .Values.global.log.functionSystem.path }}" ] && mkdir -p "{{ .Values.global.log.functionSystem.path }}"
              python3 -m yr.cli.main -v launch --inherit-env --env-subst POD_IP,NODE_ID,HOST_IP,POD_NAME,S3_ACCESS_KEY,S3_SECRET_KEY function_agent
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
        - name: S3_ACCESS_KEY
        {{ if .Values.global.obsManagement.s3AccessKey }}
          value: {{ quote .Values.global.obsManagement.s3AccessKey }}
        {{ else }}
          valueFrom:
            secretKeyRef:
              name: minio-secret-yuanrong
              key: accesskey
        {{ end }}
        - name: S3_SECRET_KEY
        {{ if .Values.global.obsManagement.s3SecretKey }}
          value: {{ quote .Values.global.obsManagement.s3SecretKey }}
        {{ else }}
          valueFrom:
            secretKeyRef:
              name: minio-secret-yuanrong
              key: secretkey
        {{ end }}
        image: "{{ .Values.global.imageRegistry | trimSuffix "/" }}/{{ .Values.global.images.functionAgent }}"
        imagePullPolicy: IfNotPresent
        livenessProbe:
          failureThreshold: {{ .Values.global.pool.livenessProbeFailureThreshold }}
          exec:
            command:
              - /bin/bash
              - -c
              - /home/sn/bin/health-check {{ quote .Values.global.port.functionAgentPort }} function-agent
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
              - /home/sn/bin/health-check {{ quote .Values.global.port.functionAgentPort }} function-agent
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
        - name: config-volume
          mountPath: /etc/yuanrong/config.toml
          subPath: config.toml
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
      - name: config-volume
        configMap:
          defaultMode: 0440
          name: components-toml-config
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