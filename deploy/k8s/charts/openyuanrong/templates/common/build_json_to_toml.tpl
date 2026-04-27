{{- define "yr.fs_log_config.json" -}}
filepath: {{ quote .Values.global.log.functionSystem.path }}
level: {{ quote .Values.global.log.functionSystem.level }}
pattern:
  separator: " | "
  placeholders:
    - flags: "%Y-%m-%dT%H:%M:%S.%f"
    - flags: "%l"
    - flags: "%s:%#"
    - env: "POD_NAME"
    - env: "CLUSTER_ID"
    - flags: ""
compress: true
rolling:
  maxsize: 1000
  maxfiles: 3
async:
  logBufSecs: 30
  maxQueueSize: 51200
  threadCount: 1
alsologtostderr: false
{{- end }}


{{- define "yr.runtime_prestart_config.json" -}}
java1.8:
  prestartCount: {{ quote .Values.global.runtime.prestartCount.java8 }}
  customArgs: {{ quote .Values.global.runtime.jvmCustomArgs }}
java11:
  prestartCount: {{ quote .Values.global.runtime.prestartCount.java11 }}
  customArgs: {{ quote .Values.global.runtime.jvmCustomArgs }}
python3.6:
  prestartCount: {{ quote .Values.global.runtime.prestartCount.python36 }}
python3.7:
  prestartCount: {{ quote .Values.global.runtime.prestartCount.python37 }}
python3.8:
  prestartCount: {{ quote .Values.global.runtime.prestartCount.python38 }}
python3.9:
  prestartCount: {{ quote .Values.global.runtime.prestartCount.python39 }}
python3.10:
  prestartCount: {{ quote .Values.global.runtime.prestartCount.python310 }}
python3.11:
  prestartCount: {{ quote .Values.global.runtime.prestartCount.python311 }}
cpp11:
  prestartCount: {{ quote .Values.global.runtime.prestartCount.cpp }}
{{- end }}


{{- define "yr.runtime_default_config.json" -}}
java1.8: {{ quote .Values.global.runtime.defaultArgs.java8 }}
java11: {{ quote .Values.global.runtime.defaultArgs.java11 }}
java17: {{ quote .Values.global.runtime.defaultArgs.java17 }}
java21: {{ quote .Values.global.runtime.defaultArgs.java21 }}
{{- end }}
