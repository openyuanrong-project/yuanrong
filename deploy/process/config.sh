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

set -o pipefail

BASE_DIR=$(dirname "$(readlink -f "$0")")
readonly BASE_DIR

[[ ! -f "${BASE_DIR}/utils.sh" ]] && echo "${BASE_DIR}/utils.sh is not exist" && exit 1
. "${BASE_DIR}"/utils.sh

CURRENT_PID=$$
NODE_ID="$(trim_hostname)-${CURRENT_PID}"

declare -A port_policy_table

LONG_OPS_INFO="master,ip_address:,cpu_num:,memory_num:,shared_memory_num:,deploy_path:,master_info_output:,master_info:,\
services_path:,master_ip:,driver_gateway_enable:,only_check_param,\
cpu_reserved:,state_storage_type:,system_timeout:,runtime_conn_timeout_s:,\
status_collect_enable:,status_collect_interval:,enable_trace:,trace_config:,runtime_trace_config:,enable_metrics:,metrics_config:,metrics_config_file:,runtime_init_call_timeout_seconds:,custom_resources:,\
labels:,control_plane_port_min:,control_plane_port_max:,data_plane_port_min:,data_plane_port_max:,block:,\
ds_worker_unique_enable,log_level:,\
enable_distributed_master,\
log_root:,fs_log_path:,fs_log_level:,fs_log_compress_enable:,fs_log_rolling_max_size:,fs_log_rolling_max_files:,fs_log_rolling_retention_days:,\
fs_log_async_log_buf_secs:,fs_log_async_log_max_queue_size:,fs_log_async_log_thread_count:,fs_also_log_to_stderr:,\
runtime_log_level:,runtime_log_path:,runtime_log_rolling_max_size:,runtime_log_rolling_max_files:,runtime_default_config:,etcd_log_path:,enable_dis_conv_call_stack:,\
runtime_std_rolling_enable:,\
ds_log_level:,ds_log_path:,ds_log_rolling_max_size:,ds_log_rolling_max_files:,\
is_schedule_tolerate_abnormal:,max_instance_cpu_size:,max_instance_memory_size:,\
min_instance_cpu_size:,min_instance_memory_size:,\
ds_spill_enable:,ds_spill_directory:,ds_spill_size_limit:,\
ds_rpc_thread_num:,ds_node_timeout_s:,ds_node_dead_timeout_s:,\
ds_heartbeat_interval_ms:,ds_client_dead_timeout_s:,ds_max_client_num:,ds_memory_reclamation_time_second:,\
ds_arena_per_tenant:,ds_enable_fallocate:,ds_enable_huge_tlb:,ds_enable_thp:,\
enable_faas_frontend:,faas_frontend_http_port:,faas_frontend_grpc_port:,enable_function_scheduler:,\
enable_meta_service:,meta_service_port:,\
function_agent_port:,function_proxy_port:,\
function_proxy_grpc_port:,global_scheduler_port:,runtime_init_port:,\
function_agent_litebus_thread:,function_master_litebus_thread:,function_proxy_litebus_thread:,\
meta_store_mode:,\
metrics_collector_type:,runtime_heartbeat_enable:,runtime_heartbeat_timeout_ms:,port_policy:,ds_worker_port:,\
npu_collection_mode:,gpu_collection_enable:,\
enable_multi_master,etcd_addr_list:,cluster_list:,\
ds_master_port:,\
curve_key_path:,runtime_ds_auth_enable:,runtime_ds_encrypt_enable:,runtime_ds_connect_timeout:,\
ds_component_auth_enable:,etcd_ssl_base_path:,cache_storage_auth_type:,cache_storage_auth_enable:,\
is_partial_watch_instances:,\
ssl_base_path:,ssl_enable:,ssl_root_file:,ssl_cert_file:,ssl_key_file:,\
runtime_max_heartbeat_timeout_times:,runtime_port_num:,runtime_recover_enable:,runtime_direct_connection_enable:,runtime_instance_debug_enable:,is_protomsg_to_runtime:,massif_enable:,\
etcd_mode:,etcd_ip:,etcd_port:,etcd_server_cert_path:,etcd_client_cert_path:,etcd_client_cert_file:,etcd_client_key_file:,\
etcd_peer_port:,etcd_compact_retention:,etcd_auth_type:,etcd_cert_file:,etcd_key_file:,etcd_ca_file:,\
local_schedule_plugins:,domain_schedule_plugins:,enable_print_perf:,enable_meta_store:,enable_persistence:,enable_jemalloc:,enable_inherit_env:,\
etcd_proxy_enable:,etcd_proxy_nums:,etcd_proxy_port:,etcd_no_fsync:,node_id:,function_agent_alias:,function_proxy_unique_enable,\
enable_separated_redirect_runtime_std:,schedule_relaxed:,user_log_export_mode:,\
max_priority:,enable_preemption:,kill_process_timeout_seconds:,\
dashboard_port:,dashboard_grpc_port:,enable_dashboard:,enable_collector:,\
prometheus_address:,prometheus_ssl_enable:,prometheus_ssl_base_path:,prometheus_ssl_root_file:,prometheus_ssl_cert_file:,prometheus_ssl_key_file:,\
dashboard_ssl_enable:,dashboard_ssl_base_path:,dashboard_ssl_cert_file:,dashboard_ssl_key_file:,\
memory_detection_interval:,oom_kill_enable:,oom_kill_control_limit:,oom_consecutive_detection_count:,\
fs_health_check_retry_times:,fs_health_check_retry_interval:,fs_health_check_timeout:,disable_nc_check,\
runtime_home_dir:,\
etcd_table_prefix:,etcd_target_name_override:,\
ds_l2_cache_type:,ds_sfs_path:,ds_log_monitor_enable:,zmq_chunk_sz:,enable_lossless_data_exit_mode:,\
meta_store_max_flush_concurrency:,meta_store_max_flush_batch_size:,\
runtime_metrics_config:,\
help"
FS_LOG_CONFIG="{\"filepath\": \"{{logConfigPath}}\",\"level\": \"{{logLevel}}\",\"compress\": {{logCompressEnable}}, \
\"rolling\": {\"maxsize\": {{logRollingMaxSize}},\"maxfiles\": {{logRollingMaxFiles}},\"retentionDays\": {{logRollingRetentionDays}}}, \
\"async\": {\"logBufSecs\": {{logAsyncBufSecs}},\"maxQueueSize\": {{logAsyncMaxQueueSize}},\"threadCount\": {{logAsyncThreadCount}}}, \
\"alsologtostderr\": {{logAlsologtostderr}}}"
FAAS_LOG_CONFIG="{\"filepath\": \"{{logConfigPath}}\",\"level\": \"{{logLevel}}\"}"
LOG_ROTATE_CONFIG="{{logConfigPath}}/runtime-*.out {{logConfigPath}}/runtime-*.err {
    size {{logRollingMaxSize}}M
    rotate {{logRollingMaxFiles}}
    copytruncate
    compress
    missingok
    notifempty
    dateext
    dateformat -%Y%m%d%H%M%S
    maxage {{logRollingRetentionDays}}
}
"

# General Configuration
FUNCTION_MASTER_IP=""
ENABLE_MASTER="false"
IP_ADDRESS=""
HOST_IP=""
MASTER_INFO=
SERVICES_PATH="${BASE_DIR}/services.yaml"
DS_WORKER_UNIQUE_ENABLE="false"
FUNCTION_PROXY_UNIQUE_ENABLE="false"
UNIQUE_KEY_PREFIX=""
INSTALL_DIR_PARENT=$(readlink -m "${BASE_DIR}/../deploy")
DATA_PLANE_INSTALL_DIR=""
MASTER_INFO_OUT_FILE=""
CPU_ALL=0
MEM_ALL=0
CPU4COMP=0
MEM4COMP=0
MEM4DATA=0
CUSTOM_RESOURCES=""
LABELS=""
CPU_RESERVED_FOR_DS_WORKER=0
readonly MEM_RESERVED=0
readonly STD_LOG_SUFFIX="_std.log"
CONTROL_PLANE_PORT_MIN=10000
CONTROL_PLANE_PORT_MAX=20000
DATA_PLANE_PORT_MIN=20000
DATA_PLANE_PORT_MAX=40000
ONLY_CHECK_PARAM="false"
STATUS_COLLECT_ENABLE="false"
STATUS_COLLECT_INTERVAL=300
ENABLE_TRACE=false
TRACE_CONFIG=""
RUNTIME_TRACE_CONFIG=""
ENABLE_METRICS=true
METRICS_CONFIG=""
METRICS_CONFIG_FILE=$(readlink -m '${BASE_DIR}/../../../functionsystem/config/metrics/metrics_config.json')
RUNTIME_METRICS_CONFIG=""
STATE_STORAGE_TYPE="datasystem"
PULL_RESOURCE_INTERVAL=1000
ENABLE_DIS_CONV_CALL_STACK=false
BLOCK=false
ENABLE_MULTI_MASTER="false"
ELECTION_MODE="standalone"
FS_HEALTH_CHECK_RETRY_TIMES=60
FS_HEALTH_CHECK_RETRY_INTERVAL=0
FS_HEALTH_CHECK_TIMEOUT=1
DISABLE_NC_CHECK="false"
ETCD_TABLE_PREFIX=""
ETCD_TARGET_NAME_OVERRIDE=""
# Logging Configuration
LOG_ROOT="default"
FS_LOG_PATH=""
FS_LOG_LEVEL="INFO"
FS_LOG_COMPRESS_ENABLE=true
FS_LOG_ROLLING_MAX_SIZE=40
FS_LOG_ROLLING_MAX_FILES=10
FS_LOG_ASYNC_LOG_BUF_SECS=30
FS_LOG_ASYNC_LOG_MAX_QUEUE_SIZE=51200
FS_LOG_ASYNC_LOG_THREAD_COUNT=1
FS_ALSO_LOG_TO_STDERR=false
FS_LOG_ROLLING_RETENTION_DAYS=30
RUNTIME_LOG_PATH=""
RUNTIME_LOG_LEVEL="INFO"
RUNTIME_LOG_ROLLING_MAX_SIZE=40
RUNTIME_LOG_ROLLING_MAX_FILES=20
RUNTIME_STD_ROLLING_ENABLE="false"
DS_LOG_PATH=""
DS_LOG_LEVEL_STR="INFO"
DS_LOG_ROLLING_MAX_FILES=32
DS_LOG_ROLLING_MAX_SIZE=40
ETCD_LOG_PATH=""
# Function System Configuration
ACCESSOR_ENABLE="false"
MIN_INSTANCE_CPU_SIZE=0
MIN_INSTANCE_MEMORY_SIZE=0
MAX_INSTANCE_CPU_SIZE=16000
MAX_INSTANCE_MEMORY_SIZE=1073741824
MAX_VALUE_FOR_CPU_AND_MEM=1073741824
ACCESSOR_HTTP_PORT=31220
ACCESSOR_GRPC_PORT=31221
DRIVER_GATEWAY_ENABLE="false"
FUNCTION_AGENT_PORT=58866
FUNCTION_PROXY_PORT=22772
FUNCTION_PROXY_GRPC_PORT=22773
GLOBAL_SCHEDULER_PORT=22770
RUNTIME_INIT_PORT=21006
DASHBOARD_PORT=9080
DASHBOARD_GRPC_PORT=9081
COLLECTOR_PORT=9082
META_SERVICE_PORT=31111
PROMETHEUS_ADDRESS=""
PROMETHEUS_SSL_ENABLE="false"
PROMETHEUS_SSL_BASE_PATH=""
PROMETHEUS_SSL_ROOT_FILE="ca.crt"
PROMETHEUS_SSL_CERT_FILE="client.crt"
PROMETHEUS_SSL_KEY_FILE="client.key"
DASHBOARD_SSL_ENABLE="false"
DASHBOARD_SSL_BASE_PATH=""
DASHBOARD_SSL_CERT_FILE="server.crt"
DASHBOARD_SSL_KEY_FILE="server.key"
METRICS_COLLECTOR_TYPE="proc"
MERGE_PROCESS_ENABLE="true"
RUNTIME_HEARTBEAT_ENABLE=true
RUNTIME_HEARTBEAT_TIMEOUT_MS=100000
RUNTIME_MAX_HEARTBEAT_TIMEOUT_TIMES=18
RUNTIME_CONN_TIMEOUT_S=30
RUNTIME_RECOVER_ENABLE=false
RUNTIME_PORT_NUM=65535
RUNTIME_INIT_CALL_TIMEOUT_SECONDS=600
RUNTIME_DEFAULT_CONFIG=""
NPU_COLLECTION_MODE="all"
GPU_COLLECTION_ENABLE=false
IS_SCHEDULE_TOLERATE_ABNORMAL=false
ENABLE_PRINT_RESOURCE_VIEW=false
PORT_POLICY="RANDOM"
RUNTIME_DIRECT_CONNECTION_ENABLE="false"
RUNTIME_INSTANCE_DEBUG_ENABLE="false"
SYS_FUNC_RETRY_PERIOD=5000
SYSTEM_TIMEOUT=1800000
FUNCTION_AGENT_LITEBUS_THREAD=20
FUNCTION_PROXY_LITEBUS_THREAD=20
FUNCTION_MASTER_LITEBUS_THREAD=20
FUNCTION_AGENT_ALIAS=""
LOCAL_SCHEDULE_PLUGINS="[\"Label\", \"ResourceSelector\", \"Default\", \"Heterogeneous\"]"
DOMAIN_SCHEDULE_PLUGINS="[\"Label\", \"ResourceSelector\", \"Default\", \"Heterogeneous\"]"
SCHEDULE_RELAXED=-1
ENABLE_PREEMPTION=false
FUNCTION_PROXY_UNREGISTER_WHILE_STOP=true
MAX_PRIORITY=0
RUNTIME_METRICS_CONFIG="false"
# Use snlib to adapt old or new runtime params
if [ -d "${BASE_DIR}/../../runtime/service/cpp/snlib" ]; then
  IS_PROTOMSG_TO_RUNTIME=false
else
  IS_PROTOMSG_TO_RUNTIME=true
fi
MASSIF_ENABLE="false"
ENABLE_PRINT_PERF="false"

ENABLE_DASHBOARD="false"
ENABLE_COLLECTOR="false"

ENABLE_META_STORE="false"
ENABLE_META_SERVICE="false"
ENABLE_PERSISTENCE="false"
META_STORE_MODE="local"
META_STORE_EXCLUDED_KEYS=","
ENABLE_JEMALLOC="true"
ENABLE_INHERIT_ENV="false"
IS_PARTIAL_WATCH_INSTANCES="false"
META_STORE_MAX_FLUSH_CONCURRENCY=100
META_STORE_MAX_FLUSH_BATCH_SIZE=50
readonly JEMALLOC_LIB_PATH=$(readlink -m "${BASE_DIR}/../../functionsystem/lib/libjemalloc.so")
# Faas-pattern Configuration
ENABLE_FAAS_FRONTEND="false"
FAAS_FRONTEND_HTTP_PORT=8888
FAAS_FRONTEND_GRPC_PORT=31223
ENABLE_FUNCTION_SCHEDULER="false"
# Data System Configuration
DS_MASTER_IP=""
DS_MASTER_PORT=12123
DS_WORKER_PORT=31501
DS_SPILL_ENABLE=false
DS_SPILL_DIRECTORY=""
DS_SPILL_SIZE_LIMIT=20480
DS_RPC_THREAD_NUM=32
DS_NODE_TIMEOUT_S=1800
DS_HEARTBEAT_INTERVAL_MS=300000
DS_CLIENT_DEAD_TIMEOUT_S=60
DS_NODE_DEAD_TIMEOUT_S=864000
DS_MAX_CLIENT_NUM=1000
DS_MEMORY_RECLAMATION_TIME_SECOND=5
DS_ARENA_PER_TENANT=16
DS_ENABLE_FALLOCATE="true"
DS_ENABLE_HUGE_TLB="false"
DS_L2_CACHE_TYPE="none"
DS_SFS_PATH=""
DS_ENABLE_THP="false"
ENABLE_DISTRIBUTED_MASTER="false"
# Third Party Configuration
ETCD_MODE="inner"
ETCD_IP=""
ETCD_PORT=32379
ETCD_PEER_PORT=32380
ETCD_AUTH_TYPE="Noauth"
ETCD_SSL_BASE_PATH=""
ETCD_CERT_FILE="server.crt"
ETCD_KEY_FILE="server.key"
ETCD_CA_FILE="ca.crt"
ETCD_CLIENT_CERT_FILE="client.crt"
ETCD_CLIENT_KEY_FILE="client.key"
ETCD_PROXY_ENABLE="false"
ENABLE_ETCD_AUTH="false"
ETCD_PROXY_NUMS=0
ETCD_PROXY_PORT=23790
ETCD_NO_FSYNC=true
ETCD_COMPACT_RETENTION=100000
MASTER_USED_DS_PORT=""
ETCD_ADDR_LIST=""
ETCD_CLUSTER_ADDRESS=""
ETCD_NAME="etcd0"
ETCD_INITIAL_CLUSTER=""
ETCD_HTTP_PROTOCOL="http://"
ETCD_INITIAL_CLUSTER_STATE="new"
ETCD_ENDPOINTS_ADDR=""
ETCD_NODE_NUM=0
ETCD_IS_FIRST_DEPLOY="true"
ETCD_DEPLOY_FAIL_CNT=0
ETCD_ADDR_LIST_FOR_MASTER_INFO=""
# ssl config
STS_CONFIG="{}"
SSL_BASE_PATH=""
SCC_ENABLE="false"
SSL_ENABLE="false"
SSL_ROOT_FILE="ca.crt"
SSL_CERT_FILE="module.crt"
SSL_KEY_FILE="module.key"
# ds auth config
DS_COMPONENT_AUTH_ENABLE="false"
CURVE_KEY_PATH=""
RUNTIME_DS_AUTH_ENABLE="false"
RUNTIME_DS_ENCRYPT_ENABLE="false"
RUNTIME_DS_CONNECT_TIMEOUT="1800"
CACHE_STORAGE_AUTH_TYPE="Noauth"
CACHE_STORAGE_AUTH_ENABLE="false"
SEPARATED_REDIRECT_RUNTIME_STD="true"
USER_LOG_EXPORT_MODE="file"
# Runtime OOM kill
MEMORY_DETECTION_INTERVAL="1000"
OOM_KILL_ENABLE="false"
OOM_KILL_CONTROL_LIMIT="0"
OOM_CONSECUTIVE_DETECTION_COUNT="3"
# kill -SIGKILL process after kill -SIGINT
KILL_PROCESS_TIMEOUT_SECONDS="0"
# runtime user home dir
RUNTIME_USER_HOME_DIR=""
DS_LOG_MONITOR_ENABLE="false"
ZMQ_CHUNK_SZ="1048576"
ENABLE_LOSSLESS_DATA_EXIT_MODE="false"
# Runtime ssl config
ENABLE_MTLS="false"
PRIVATE_KEY_PATH=""
CERTIFICATE_FILE_PATH=""
VERIFY_FILE_PATH=""
PRIMARY_KEY_STORE_FILE=""
STANDBY_KEY_STORE_FILE=""
ENCRYPT_RUNTIME_PUBLIC_KEY_CONTEXT=""
ENCRYPT_RUNTIME_PRIVATE_KEY_CONTEXT=""
ENCRYPT_DS_PUBLIC_KEY_CONTEXT=""

function usage() {
  echo -e "General Options:"
  echo -e "     -a, --ip_address                                    node ip address"
  echo -e "     -c, --cpu_num                                       overall cpu cores (1/1000 core) in current script context"
  echo -e "     -d, --deploy_path                                   it is recommended to use an empty directory as deploy_path"
  echo -e "     -g, --driver_gateway_enable                         enable driver gateway, options:true/false (default false)"
  echo -e "     -m, --memory_num                                    overall memory (MB) in current script context(should be larger than shared_memory_num)"
  echo -e "     -s, --shared_memory_num                             data system shared memory (MB) should be reserved in current script context"
  echo -e "     -o, --master_info_output                            master info output file, absolute path, for agent, it will output the agent info"
  echo -e "     --master                                            deploy control plane, otherwise only deploy data plane"
  echo -e "     --master_info                                       master info"
  echo -e "     -p, --services_path                                 auto load function path (default /home/sn/config/services.yaml)"
  echo -e "     -w, --ds_worker_unique_enable                       enable all data planes on one node share one ds-worker"
  echo -e "     -f, --function_proxy_unique_enable                  enable all data planes on one node share one function-proxy"
  echo -e "     -e, --enable_multi_master                           enable multi master"
  echo -e "     --control_plane_port_min                            min port for control plane component (default 10000)"
  echo -e "     --control_plane_port_max                            max port for control plane component (default 20000)"
  echo -e "     --data_plane_port_min                               min port for data plane component (default 20000)"
  echo -e "     --data_plane_port_max                               max port for data plane component (default 40000)"
  echo -e "     --master_ip                                         master ip for data plane component"
  echo -e "     --status_collect_enable                             enable process status collect, options:true/false(default false)"
  echo -e "     --status_collect_interval                           interval of process status collect, unit second(default 300)"
  echo -e "     --enable_trace                                      enable trace, options:true/false(default false)"
  echo -e "     --trace_config                                      function system trace config, should be json string"
  echo -e "     --runtime_trace_config                              runtime trace config, should be json string"
  echo -e "     --enable_metrics                                    enable metrics, options:true/false(default true)"
  echo -e "     --metrics_config                                    metrics config, should be json string"
  echo -e "     --metrics_config_file                               metrics config file path, should be absolute path"
  echo -e "     --cpu_reserved                                      cpu reserved for data system worker(default 0)"
  echo -e "     --state_storage_type                                state storage type, options: disable/datasystem (default disable)"
  echo -e "     --pull_resource_interval                            set the interval of pull resource, unit ms(default 1000)"
  echo -e "     --enable_dis_conv_call_stack                        enable distributed convergent call stack, options:true/false(default true)"
  echo -e "     --block                                             deploy by the blocking way or run in background(default false for master and true for agent)"
  echo -e "     --custom_resources                                  will pass to function agent as custom resources"
  echo -e "     --labels                                            will pass to function agent as init labels"
  echo -e "     --disable_nc_check                                  disable nc check port"
  echo -e "     --etcd_mode                                         etcd deploy mode, options:inner/outter"
  echo -e "     --zmq_chunk_sz                                      Parallel payload split chunk size, unit: MB (default 1048756)"
  echo -e "Logging Options:"
  echo -e "     -l, --log_level                                     log level for the whole cluster (default INFO)"
  echo -e "     --log_root                                          log root directory, should be absolute path, default is 'deploy_path/NODE_ID/log'"
  echo -e "     --fs_log_level                                      log level for function system (default INFO)"
  echo -e "     --fs_log_compress_enable                            log compress for function system (default true)"
  echo -e "     --fs_log_path                                       log subdirectory for function system"
  echo -e "     --fs_log_rolling_max_size                           rolling log max size for function system, unit: MB (default 40)"
  echo -e "     --fs_log_rolling_max_files                          rolling log max files for function system (default 10)"
  echo -e "     --fs_log_rolling_retention_days                     maximum duration for storing compressed logs of function system, unit: days (default 30)"
  echo -e "     --fs_log_async_log_buf_secs                         time when logs are stored in the buffer before flush (default 30)"
  echo -e "     --fs_log_async_log_max_queue_size                   async log max queue size (default 51200)"
  echo -e "     --fs_log_async_log_thread_count                     async log thread count (default 1)"
  echo -e "     --fs_also_log_to_stderr                             log to stderr for function system(default false)"
  echo -e "     --runtime_log_path                                  log subdirectory for runtime"
  echo -e "     --runtime_log_level                                 runtime log level"
  echo -e "     --runtime_log_rolling_max_size                      rolling log max size for runtime, unit: MB (default 40)"
  echo -e "     --runtime_log_rolling_max_files                     rolling log max files for runtime (default 20)"
  echo -e "     --runtime_std_rolling_enable                        enable rolling log for runtime std log, depends on logrotate (default false)"
  echo -e "     --enable_separated_redirect_runtime_std             enable to redirect standard output of runtime separated. which does not support rotation & compress etc. {runtimeID}.out {runtimeID}.err"
  echo -e "     --user_log_export_mode                              user log export mode. options: file(export to file)/std(export to std), (default file)"
  echo -e "     --ds_log_level                                      data system log level, options: DEBUG/INFO/WARN/ERROR (default INFO)"
  echo -e "     --ds_log_path                                       log subdirectory for data system"
  echo -e "     --ds_log_rolling_max_size                           rolling log max size for data system, unit: MB (default 40)"
  echo -e "     --ds_log_rolling_max_files                          rolling log max files for data system (default 32)"
  echo -e "     --ds_log_monitor_enable                             enable data system log monitor(default false)"
  echo -e "     --etcd_log_path                                     log subdirectory for etcd"
  echo -e "Function System Options:"
  echo -e "     --min_instance_cpu_size                             instance min CPU size in request allowed (unit:1/1000 core, default 300)"
  echo -e "     --min_instance_memory_size                          instance min memory size in request allowed (unit:MB, default 128)"
  echo -e "     --max_instance_cpu_size                             instance max CPU size in request allowed (unit:1/1000 core, default 16000)"
  echo -e "     --max_instance_memory_size                          instance max memory size in request allowed (unit:MB, default 128)"
  echo -e "     --function_agent_port                               function agent port (default 58866)"
  echo -e "     --function_proxy_port                               function proxy port (default 22772)"
  echo -e "     --function_proxy_grpc_port                          function proxy port for driver (default 22773)"
  echo -e "     --global_scheduler_port                             global scheduler port (default 22770)"
  echo -e "     --runtime_init_port                                 runtime init port (default 21006)"
  echo -e "     --dashboard_port                                    dashboard port (default 9080)"
  echo -e "     --dashboard_grpc_port                               dashboard grpc port (default 9081)"
  echo -e "     --collector_port                                    collector port (default 9082)"
  echo -e "     --prometheus_address                                prometheus address(default empty, format: prometheus_ip:prometheus_port)"
  echo -e "     --prometheus_ssl_enable                             prometheus ssl enabled, options: true/false (default false)"
  echo -e "     --prometheus_ssl_base_path                          prometheus ssl base path, configure absolute path"
  echo -e "     --prometheus_ssl_root_file                          prometheus ssl root ca file name, default is ca.crt"
  echo -e "     --prometheus_ssl_cert_file                          prometheus ssl module cert file name, default is client.crt"
  echo -e "     --prometheus_ssl_key_file                           prometheus ssl module key file name, default is client.key"
  echo -e "     --dashboard_ssl_enable                              dashboard ssl enabled, options: true/false (default false)"
  echo -e "     --dashboard_ssl_base_path                           dashboard ssl base path, configure absolute path"
  echo -e "     --dashboard_ssl_cert_file                           dashboard ssl module cert file name, default is server.crt"
  echo -e "     --dashboard_ssl_key_file                            dashboard ssl module key file name, default is server.key"
  echo -e "     --metrics_collector_type                            runtime manager metrics collector type (default proc)"
  echo -e "     --port_policy                                       assign port policy, options: RANDOM, FIX(default RANDOM)"
  echo -e "     --runtime_heartbeat_enable                          enable heartbeat between function_proxy and runtime (default true)"
  echo -e "     --runtime_heartbeat_timeout_ms                      heartbeat timeout between function_proxy and runtime, unit ms(default 100000)"
  echo -e "     --runtime_max_heartbeat_timeout_times               max heartbeat timeout times between function_proxy and runtime (default 18)"
  echo -e "     --runtime_conn_timeout_s                            runtime first connect timeout, unit second (default 30)"
  echo -e "     --runtime_port_num                                  runtime port num in runtime manager (default 65535)"
  echo -e "     --runtime_recover_enable                            enable recover runtime (default false)"
  echo -e "     --runtime_direct_connection_enable                  enable runtime direct connection (default false)"
  echo -e "     --runtime_instance_debug_enable                     enable runtime instance debug (default false)"
  echo -e "     --runtime_default_config                            runtime default config"
  echo -e "     --is_protomsg_to_runtime                            send protobuf message to runtime, otherwise send json message (default false)"
  echo -e "     --npu_collection_mode                               collect npu info mode (default all), support all, count, hbm, sfmd, topo, off"
  echo -e "     --gpu_collection_enable                             enable collect gpu info (default false)"
  echo -e "     --local_schedule_plugins                            local schedule plugins (default enable all plugins)"
  echo -e "     --domain_schedule_plugins                           domain schedule plugins (default enable all plugins)"
  echo -e "     --runtime_init_call_timeout_seconds                 init runtime call timeout, unit second(default 600)"
  echo -e "     --is_schedule_tolerate_abnormal                     is schedule tolerate abnormal, options: true/false(default false)"
  echo -e "     --system_timeout                                    timeout including request and timeout, unit ms(default 1800000)"
  echo -e "     --function_agent_litebus_thread                     function agent litebus thread count(default 20)"
  echo -e "     --function_master_litebus_thread                    function master litebus thread count(default 20)"
  echo -e "     --function_proxy_litebus_thread                     function proxy litebus thread count(default 20)"
  echo -e "     --function_agent_alias                              function agent alias(default empty)"
  echo -e "     --enable_print_perf                                 function proxy enable to print perf info"
  echo -e "     --enable_dashboard                                  for to enable dashboard(default false)"
  echo -e "     --enable_collector                                  for to enable collector(default false)"
  echo -e "     --enable_meta_service                               for to enable meta service(default false)"
  echo -e "     --meta_service_port                                 meta_service port (default 31111)"
  echo -e "     --enable_meta_store                                 for to enable meta store(default false)"
  echo -e "     --enable_persistence                                enable meta store to persist into etcd (default false)"
  echo -e "     --meta_store_max_flush_concurrency                  max flush concurrency for meta store backup(default 1000)"
  echo -e "     --meta_store_max_flush_batch_size                   max flush batch size for meta store backup(default 100)"
  echo -e "     --meta_store_mode                                   meta-store mode (default local)"
  echo -e "     --enable_jemalloc                                   to enable jemalloc for function master and function proxy (default true)"
  echo -e "     --enable_inherit_env                                to enable create runtime with inherit env from runtime-manager (default false)"
  echo -e "     --is_partial_watch_instances                        only watch partial instances in function proxy (default false)"
  echo -e "     --fs_health_check_timeout                           timeout of function system component health check, unit s (default 1)"
  echo -e "     --fs_health_check_retry_interval                    retry interval of function system component health check, unit s (default 0)"
  echo -e "     --fs_health_check_retry_times                       retry times of function system component health check (default 60)"
  echo -e "     --enable_faas_frontend                              enable faasfrontend, options:true/false (default false)"
  echo -e "     --faas_frontend_http_port                           faas frontend http port (default 8888)"
  echo -e "     --faas_frontend_grpc_port                           faas frontend grpc port (default 31223)"
  echo -e "     --enable_function_scheduler                         enable function scheduler, options:true/false (default false)"
  echo -e "     --schedule_relaxed                                  enable the relaxed scheduling policy. When the relaxed number of available nodes or pods is selected, the scheduling progress exits without traversing all nodes or pods.(default 1)"
  echo -e "     --max_priority                                      schedule max priority (default 0)"
  echo -e "     --enable_preemption                                 enable schedule preemption while higher priority, only valid while max_priority > 0 (default false)"
  echo -e "     --kill_process_timeout_seconds                      time interval send kill -9 after send kill -2, unit second(default 5)"
  echo -e "     --runtime_home_dir                                  runtime home dir(default is Home environment variable of the OpenYuanrong component deployment user)"
  echo -e "Data System Options:"
  echo -e "     --ds_master_port                                    data system master listening port (default 12123)"
  echo -e "     --ds_worker_port                                    data system worker listening port (default 31501)"
  echo -e "     --ds_spill_enable                                   enable data system spill (default false)"
  echo -e "     --ds_spill_directory                                data system spill directory"
  echo -e "     --ds_spill_size_limit                               data system spill size limit, unit: MB (default 20480)"
  echo -e "     --ds_rpc_thread_num                                 data system rpc thread num(default 32)"
  echo -e "     --ds_node_timeout_s                                 data system master node timeout, unit second(default 1800)"
  echo -e "     --ds_heartbeat_interval_ms                          data system worker heartbeat interval, unit ms(default 300000)"
  echo -e "     --ds_node_dead_timeout_s                            data system master node dead timeout, unit second(default 864000)"
  echo -e "     --ds_client_dead_timeout_s                          data system master client dead timeout, unit second(default 60)"
  echo -e "     --ds_max_client_num                                 data system worker max client number (default 1000)"
  echo -e "     --ds_memory_reclamation_time_second                 data system worker memory reclamation time, unit second(default 5)"
  echo -e "     --ds_component_auth_enable                          data system component auth enabled(default false)"
  echo -e "     --ds_arena_per_tenant                               data system tenant manager the number of arena"
  echo -e "     --ds_enable_fallocate                               data system fallocate enabled(default true)"
  echo -e "     --ds_enable_huge_tlb                                data system shared memory huge tlb enabled(default false)"
  echo -e "     --ds_enable_thp                                     data system shared memory thp enabled(default false)"
  echo -e "     --ds_l2_cache_type                                  data system l2 cache type(default 'none')"
  echo -e "     --ds_sfs_path                                       data system l2 cache sfs path(default '')"
  echo -e "     --curve_key_path                                    data system curve key path, the dir must contains worker.key,worker.key_secret,client.key,client.key_secret"
  echo -e "     --runtime_ds_auth_enable                            runtime and data system auth enabled(default false)"
  echo -e "     --runtime_ds_encrypt_enable                         runtime and data system encrypt enable(default false)"
  echo -e "     --runtime_ds_connect_timeout                        runtime and data system connection timeout(s)(default 1800)"
  echo -e "     --enable_distributed_master                         to enable distributed deploy master(default false)"
  echo -e "     --enable_lossless_data_exit_mode                    to enable lossless data exit mode(default false)"
  echo -e "     --cache_storage_auth_type                           for cache storage service auth type, eg: Noauth, ZMQ, AK/SK"
  echo -e "     --cache_storage_auth_enable                         for cache storage service auth"
  echo -e "Third Party Options:"
  echo -e "     --etcd_ip                                           ip address used to start etcd (default node ip)"
  echo -e "     --etcd_port                                         port used to start etcd"
  echo -e "     --etcd_peer_port                                    port used to listen other etcd message in cluster"
  echo -e "     --etcd_auth_type                                    etcd authorization type, options: Noauth/TLS (default Noauth)"
  echo -e "     --etcd_cert_file                                    certificate used for SSL/TLS config to core_etcd server default is server.crt"
  echo -e "     --etcd_key_file                                     key file for server certificate, default is server.key"
  echo -e "     --etcd_ca_file                                      trusted certificate file authority default is ca.crt"
  echo -e "     --etcd_ssl_base_path                                etcd ssl file base directory, configure absolute path"
  echo -e "     --etcd_client_cert_file                             client certificate used for SSL/TLS config to proxy connect to core_etcd, default is client.crt"
  echo -e "     --etcd_client_key_file                              key file for etcd client certificate, default is client.key"
  echo -e "     --etcd_proxy_enable                                 etcd proxy enable, options: true/false (default false)"
  echo -e "     --etcd_proxy_nums                                   while proxy is enabled, etcd_proxy_nums of etcd proxy will be starting"
  echo -e "     --etcd_proxy_port                                   specify proxy endpoint, default 23790"
  echo -e "     --etcd_no_fsync                                     set etcd_no_fsync true will improve I/O performance, default true"
  echo -e "     --etcd_compact_retention                            set etcd compact retention revision num, default 100000"
  echo -e "     --massif_enable                                     starting the runtime with valgrind massif (default false)"
  echo -e "     --etcd_addr_list                                    etcd cluster address list, such as ip1,ip2 or ip1:port1,ip2:port2 or ip1:port1:peer_port1,ip2:port2:peer_port2"
  echo -e "     --etcd_table_prefix                                 etcd table prefix"
  echo -e "     --etcd_target_name_override                         etcd target name for ssl verify"
  echo -e "TLS Config Options:"
  echo -e "     --ssl_base_path                                     ssl base path, configure absolute path"
  echo -e "     --ssl_enable                                        ssl enabled, options: true/false (default false)"
  echo -e "     --ssl_root_file                                     ssl root ca file name, default is ca.crt"
  echo -e "     --ssl_cert_file                                     ssl module cert file name, default is module.crt"
  echo -e "     --ssl_key_file                                      ssl module key file name, default is module.key"
  echo -e "Runtime OOM kill Options:"
  echo -e "     --memory_detection_interval                         memory detection interval for runtime process, unit in mili seconds, default is 1000 ms, min is 100ms"
  echo -e "     --oom_kill_enable                                   enable runtime oom kill base on process memory usage"
  echo -e "     --oom_kill_control_limit                            configure the control limit for the runtime OOM kill based on process memory usage, unit is MB."
  echo -e "     --oom_consecutive_detection_count                   number of consecutive times the memory usage must exceed the control limit before triggering OOM kill"
  echo -e "     --runtime_metrics_config                            runtime_metrics_config, default is false"
}

function help_msg() {
  echo -e "try 'bash [deploy.sh|yr_master.sh|yr_agent.sh] -h' or 'bash  [deploy.sh|yr_master.sh|yr_agent.sh] --help'  for more information"
}

function parse_opt() {
  option_args=$(getopt -o a:b:c:g:l:m:s:d:o:wn:p:fhe -l $LONG_OPS_INFO -- "$@")
  if [ $? -ne 0 ]; then
    return 1
  fi
  eval set -- "$option_args"
  while true; do
    case $1 in
    -a|--ip_address) IP_ADDRESS=$2 && shift 2 ;;
    -c|--cpu_num) CPU_ALL=$2 && shift 2 ;;
    -m|--memory_num) MEM_ALL=$2 && shift 2 ;;
    -s|--shared_memory_num) MEM4DATA=$2 && shift 2 ;;
    -d|--deploy_path) INSTALL_DIR_PARENT=$(readlink -m "$2") && shift 2 ;;
    -g|--driver_gateway_enable) DRIVER_GATEWAY_ENABLE=$2 && shift 2 ;;
    -n|--node_id) NODE_ID=$2 && shift 2 ;;
    -o|--master_info_output) MASTER_INFO_OUT_FILE=$2 && shift 2 ;;
    -w|--ds_worker_unique_enable) DS_WORKER_UNIQUE_ENABLE=true && shift 1 ;;
    -f|--function_proxy_unique_enable) FUNCTION_PROXY_UNIQUE_ENABLE=true && shift 1 ;;
    -e|--enable_multi_master) ENABLE_MULTI_MASTER=true && shift 1 ;;
    -h|--help) usage && exit 0 ;;
    --master) ENABLE_MASTER="true" && shift 1 ;;
    --master_info) MASTER_INFO=$2 && shift 2 ;;
    --enable_distributed_master) ENABLE_DISTRIBUTED_MASTER="true" && shift 1 ;;
    -p|--services_path) SERVICES_PATH=$2 && shift 2 ;;
    --custom_resources) CUSTOM_RESOURCES=$2 && shift 2 ;;
    --labels) LABELS=$2 && shift 2 ;;
    --disable_nc_check) DISABLE_NC_CHECK="true" && shift 1 ;;
    --control_plane_port_min) CONTROL_PLANE_PORT_MIN=$2 && shift 2 ;;
    --control_plane_port_max) CONTROL_PLANE_PORT_MAX=$2 && shift 2 ;;
    --data_plane_port_min) DATA_PLANE_PORT_MIN=$2 && shift 2 ;;
    --data_plane_port_max) DATA_PLANE_PORT_MAX=$2 && shift 2 ;;
    --master_ip) FUNCTION_MASTER_IP=$2 && DS_MASTER_IP=$2 && shift 2 ;;
    --state_storage_type) STATE_STORAGE_TYPE=$2 && shift 2 ;;
    --pull_resource_interval) PULL_RESOURCE_INTERVAL=$2 && shift 2 ;;
    --only_check_param) ONLY_CHECK_PARAM="true" && shift ;;
    --enable_trace) ENABLE_TRACE=$2 && shift 2 ;;
    --trace_config) TRACE_CONFIG=$2 && shift 2 ;;
    --runtime_trace_config) RUNTIME_TRACE_CONFIG=$2 && shift 2 ;;
    --enable_dis_conv_call_stack) ENABLE_DIS_CONV_CALL_STACK=$2 && shift 2 ;;
    --enable_metrics) ENABLE_METRICS=$2 && shift 2 ;;
    --metrics_config) METRICS_CONFIG=$2 && shift 2 ;;
    --metrics_config_file) METRICS_CONFIG_FILE=$2 && shift 2 ;;
    --cpu_reserved) CPU_RESERVED_FOR_DS_WORKER=$2 && shift 2 ;;
    --status_collect_enable) STATUS_COLLECT_ENABLE=$2 && shift 2 ;;
    --status_collect_interval) STATUS_COLLECT_INTERVAL=$2 && shift 2 ;;
    -l|--log_level) FS_LOG_LEVEL=$2 && RUNTIME_LOG_LEVEL=$2 && DS_LOG_LEVEL_STR=$2 && shift 2 ;;
    --log_root) LOG_ROOT=$2 && shift 2 ;;
    --fs_log_level) FS_LOG_LEVEL=$2 && shift 2 ;;
    --fs_log_path) FS_LOG_PATH=$2 && shift 2 ;;
    --fs_log_compress_enable) FS_LOG_COMPRESS_ENABLE=$2 && shift 2 ;;
    --fs_log_rolling_max_size) FS_LOG_ROLLING_MAX_SIZE=$2 && shift 2 ;;
    --fs_log_rolling_max_files) FS_LOG_ROLLING_MAX_FILES=$2 && shift 2 ;;
    --fs_log_rolling_retention_days) FS_LOG_ROLLING_RETENTION_DAYS=$2 && shift 2 ;;
    --fs_log_async_log_buf_secs) FS_LOG_ASYNC_LOG_BUF_SECS=$2 && shift 2 ;;
    --fs_log_async_log_max_queue_size) FS_LOG_ASYNC_LOG_MAX_QUEUE_SIZE=$2 && shift 2 ;;
    --fs_log_async_log_thread_count) FS_LOG_ASYNC_LOG_THREAD_COUNT=$2 && shift 2 ;;
    --fs_also_log_to_stderr) FS_ALSO_LOG_TO_STDERR=$2 && shift 2 ;;
    --runtime_log_level) RUNTIME_LOG_LEVEL=$2 && shift 2 ;;
    --runtime_log_path) RUNTIME_LOG_PATH=$2 && shift 2 ;;
    --runtime_log_rolling_max_size) RUNTIME_LOG_ROLLING_MAX_SIZE=$2 && shift 2 ;;
    --runtime_log_rolling_max_files) RUNTIME_LOG_ROLLING_MAX_FILES=$2 && shift 2 ;;
    --runtime_std_rolling_enable) RUNTIME_STD_ROLLING_ENABLE=$2 && shift 2 ;;
    --enable_separated_redirect_runtime_std) SEPARATED_REDIRECT_RUNTIME_STD=$2 && shift 2 ;;
    --user_log_export_mode) USER_LOG_EXPORT_MODE=$2 && shift 2 ;;
    --ds_log_level) DS_LOG_LEVEL_STR=$2 && shift 2 ;;
    --ds_log_path) DS_LOG_PATH=$2 && shift 2 ;;
    --ds_log_rolling_max_size) DS_LOG_ROLLING_MAX_SIZE=$2 && shift 2 ;;
    --ds_log_rolling_max_files) DS_LOG_ROLLING_MAX_FILES=$2 && shift 2 ;;
    --ds_log_monitor_enable) DS_LOG_MONITOR_ENABLE=$2 && shift 2 ;;
    --etcd_log_path) ETCD_LOG_PATH=$2 && shift 2 ;;
    --fs_health_check_timeout) FS_HEALTH_CHECK_TIMEOUT=$2 && shift 2 ;;
    --fs_health_check_retry_times) FS_HEALTH_CHECK_RETRY_TIMES=$2 && shift 2 ;;
    --fs_health_check_retry_interval) FS_HEALTH_CHECK_RETRY_INTERVAL=$2 && shift 2 ;;
    --enable_faas_frontend) ENABLE_FAAS_FRONTEND=$2 && shift 2 ;;
    --faas_frontend_http_port) FAAS_FRONTEND_HTTP_PORT=$2 && port_policy_table["faas_frontend_http_port"]="FIX" && shift 2 ;;
    --faas_frontend_grpc_port) FAAS_FRONTEND_GRPC_PORT=$2 && port_policy_table["faas_frontend_grpc_port"]="FIX" && shift 2 ;;
    --enable_function_scheduler) ENABLE_FUNCTION_SCHEDULER=$2 && shift 2 ;;
    --enable_meta_service) ENABLE_META_SERVICE=$2 && shift 2 ;;
    --meta_servce_port) META_SERVICE_PORT=$2 && port_policy_table["meta_service_port"]="FIX" && shift 2 ;;
    --min_instance_cpu_size) MIN_INSTANCE_CPU_SIZE=$2 && shift 2 ;;
    --min_instance_memory_size) MIN_INSTANCE_MEMORY_SIZE=$2 && shift 2 ;;
    --max_instance_cpu_size) MAX_INSTANCE_CPU_SIZE=$2 && shift 2 ;;
    --max_instance_memory_size) MAX_INSTANCE_MEMORY_SIZE=$2 && shift 2 ;;
    --function_agent_port) FUNCTION_AGENT_PORT=$2 && port_policy_table["function_agent_port"]="FIX" && shift 2 ;;
    --function_proxy_port) FUNCTION_PROXY_PORT=$2 && port_policy_table["function_proxy_port"]="FIX" && shift 2 ;;
    --function_proxy_grpc_port) FUNCTION_PROXY_GRPC_PORT=$2 && port_policy_table["function_proxy_grpc_port"]="FIX" && shift 2 ;;
    --runtime_init_port) RUNTIME_INIT_PORT=$2 && port_policy_table["runtime_init_port"]="FIX" && shift 2 ;;
    --global_scheduler_port) GLOBAL_SCHEDULER_PORT=$2 && port_policy_table["global_scheduler_port"]="FIX" && shift 2 ;;
    --metrics_collector_type) METRICS_COLLECTOR_TYPE=$2 && shift 2 ;;
    --port_policy) PORT_POLICY=$2 && shift 2 ;;
    --runtime_heartbeat_enable) RUNTIME_HEARTBEAT_ENABLE=$2 && shift 2 ;;
    --runtime_heartbeat_timeout_ms) RUNTIME_HEARTBEAT_TIMEOUT_MS=$2 && shift 2 ;;
    --runtime_max_heartbeat_timeout_times) RUNTIME_MAX_HEARTBEAT_TIMEOUT_TIMES=$2 && shift 2 ;;
    --runtime_port_num) RUNTIME_PORT_NUM=$2 && shift 2 ;;
    --runtime_recover_enable) RUNTIME_RECOVER_ENABLE=$2 && shift 2 ;;
    --runtime_direct_connection_enable) RUNTIME_DIRECT_CONNECTION_ENABLE=$2 && shift 2 ;;
    --runtime_instance_debug_enable) RUNTIME_INSTANCE_DEBUG_ENABLE=$2 && shift 2 ;;
    --runtime_default_config) RUNTIME_DEFAULT_CONFIG=$2 && shift 2 ;;
    --runtime_metrics_config)  RUNTIME_METRICS_CONFIG=$2 && shift 2 ;;
    --npu_collection_mode) NPU_COLLECTION_MODE=$2 && shift 2 ;;
    --gpu_collection_enable) GPU_COLLECTION_ENABLE=$2 && shift 2 ;;
    --runtime_init_call_timeout_seconds) RUNTIME_INIT_CALL_TIMEOUT_SECONDS=$2 && shift 2 ;;
    --is_schedule_tolerate_abnormal) IS_SCHEDULE_TOLERATE_ABNORMAL=$2 && shift 2 ;;
    --system_timeout) SYSTEM_TIMEOUT=$2 && shift 2 ;;
    --runtime_conn_timeout_s) RUNTIME_CONN_TIMEOUT_S=$2 && shift 2 ;;
    --function_agent_litebus_thread) FUNCTION_AGENT_LITEBUS_THREAD=$2 && shift 2 ;;
    --function_proxy_litebus_thread) FUNCTION_PROXY_LITEBUS_THREAD=$2 && shift 2 ;;
    --function_master_litebus_thread) FUNCTION_MASTER_LITEBUS_THREAD=$2 && shift 2 ;;
    --function_agent_alias) FUNCTION_AGENT_ALIAS=$2 && shift 2 ;;
    --local_schedule_plugins) LOCAL_SCHEDULE_PLUGINS=$2 && shift 2 ;;
    --domain_schedule_plugins) DOMAIN_SCHEDULE_PLUGINS=$2 && shift 2 ;;
    --enable_print_perf) ENABLE_PRINT_PERF=$2 && shift 2 ;;
    --enable_meta_store) ENABLE_META_STORE=$2 && shift 2 ;;
    --enable_dashboard) ENABLE_DASHBOARD=$2 && shift 2 ;;
    --enable_collector) ENABLE_COLLECTOR=$2 && shift 2 ;;
    --enable_persistence) ENABLE_PERSISTENCE=$2 && shift 2 ;;
    --meta_store_mode) META_STORE_MODE=$2 && shift 2 ;;
    --enable_jemalloc) ENABLE_JEMALLOC=$2 && shift 2 ;;
    --enable_inherit_env) ENABLE_INHERIT_ENV=$2 && shift 2 ;;
    --meta_store_max_flush_concurrency) META_STORE_MAX_FLUSH_CONCURRENCY=$2 && shift 2 ;;
    --meta_store_max_flush_batch_size) META_STORE_MAX_FLUSH_BATCH_SIZE=$2 && shift 2 ;;
    --is_partial_watch_instances) IS_PARTIAL_WATCH_INSTANCES=$2 && shift 2 ;;
    --ds_master_port) DS_MASTER_PORT=$2 && port_policy_table["ds_master_port"]="FIX" && shift 2 ;;
    --ds_worker_port) DS_WORKER_PORT=$2 && port_policy_table["ds_worker_port"]="FIX" && shift 2 ;;
    --ds_spill_enable) DS_SPILL_ENABLE=$2 && shift 2 ;;
    --ds_spill_directory) DS_SPILL_DIRECTORY=$2 && shift 2 ;;
    --ds_spill_size_limit) DS_SPILL_SIZE_LIMIT=$2 && shift 2 ;;
    --ds_rpc_thread_num) DS_RPC_THREAD_NUM=$2 && shift 2 ;;
    --ds_node_dead_timeout_s) DS_NODE_DEAD_TIMEOUT_S=$2 && shift 2 ;;
    --ds_client_dead_timeout_s) DS_CLIENT_DEAD_TIMEOUT_S=$2 && shift 2 ;;
    --ds_node_timeout_s) DS_NODE_TIMEOUT_S=$2 && shift 2 ;;
    --ds_heartbeat_interval_ms) DS_HEARTBEAT_INTERVAL_MS=$2 && shift 2 ;;
    --ds_max_client_num) DS_MAX_CLIENT_NUM=$2 && shift 2 ;;
    --ds_memory_reclamation_time_second) DS_MEMORY_RECLAMATION_TIME_SECOND=$2 && shift 2 ;;
    --ds_arena_per_tenant) DS_ARENA_PER_TENANT=$2 && shift 2 ;;
    --ds_enable_fallocate) DS_ENABLE_FALLOCATE=$2 && shift 2 ;;
    --ds_enable_huge_tlb) DS_ENABLE_HUGE_TLB=$2 && shift 2 ;;
    --ds_enable_thp) DS_ENABLE_THP=$2 && shift 2 ;;
    --ds_l2_cache_type) DS_L2_CACHE_TYPE=$2 && shift 2 ;;
    --ds_sfs_path) DS_SFS_PATH=$2 && shift 2 ;;
    --zmq_chunk_sz) ZMQ_CHUNK_SZ=$2 && shift 2 ;;
    --enable_lossless_data_exit_mode) ENABLE_LOSSLESS_DATA_EXIT_MODE=$2 && shift 2 ;;
    --curve_key_path) CURVE_KEY_PATH=$2 && shift 2 ;;
    --ds_component_auth_enable) DS_COMPONENT_AUTH_ENABLE=$2 && shift 2 ;;
    --runtime_ds_auth_enable) RUNTIME_DS_AUTH_ENABLE=$2 && shift 2 ;;
    --runtime_ds_encrypt_enable) RUNTIME_DS_ENCRYPT_ENABLE=$2 && shift 2 ;;
    --cache_storage_auth_type) CACHE_STORAGE_AUTH_TYPE=$2 && shift 2 ;;
    --cache_storage_auth_enable) CACHE_STORAGE_AUTH_ENABLE=$2 && shift 2 ;;
    --etcd_mode) ETCD_MODE=$2 && shift 2 ;;
    --etcd_ip) ETCD_IP=$2 && shift 2 ;;
    --etcd_port) ETCD_PORT=$2 && port_policy_table["etcd_port"]="FIX" && shift 2 ;;
    --etcd_peer_port) ETCD_PEER_PORT=$2 && port_policy_table["etcd_peer_port"]="FIX" && shift 2 ;;
    --etcd_auth_type) ETCD_AUTH_TYPE=$2 && shift 2 ;;
    --etcd_cert_file) ETCD_CERT_FILE=$2 && shift 2 ;;
    --etcd_key_file) ETCD_KEY_FILE=$2 && shift 2 ;;
    --etcd_ca_file) ETCD_CA_FILE=$2 && shift 2 ;;
    --etcd_client_cert_file) ETCD_CLIENT_CERT_FILE=$2 && shift 2 ;;
    --etcd_client_key_file) ETCD_CLIENT_KEY_FILE=$2 && shift 2 ;;
    --etcd_ssl_base_path) ETCD_SSL_BASE_PATH=$2 && shift 2 ;;
    --etcd_proxy_enable) ETCD_PROXY_ENABLE=$2 && shift 2 ;;
    --etcd_proxy_nums) ETCD_PROXY_NUMS=$2 && shift 2 ;;
    --etcd_proxy_port) ETCD_PROXY_PORT=$2 && shift 2 ;;
    --etcd_no_fsync) ETCD_NO_FSYNC=$2 && shift 2 ;;
    --etcd_compact_retention) ETCD_COMPACT_RETENTION=$2 && shift 2 ;;
    --etcd_addr_list) ETCD_ADDR_LIST=$2 && shift 2 ;;
    --etcd_target_name_override) ETCD_TARGET_NAME_OVERRIDE=$2 && shift 2 ;;
    --etcd_table_prefix) ETCD_TABLE_PREFIX=$2 && shift 2 ;;
    --is_protomsg_to_runtime)  IS_PROTOMSG_TO_RUNTIME=$2 && shift 2 ;;
    --massif_enable) MASSIF_ENABLE=$2 && shift 2 ;;
    --block) BLOCK=$2 && shift 2 ;;
    --ssl_base_path) SSL_BASE_PATH=$2 && shift 2 ;;
    --ssl_enable) SSL_ENABLE=$2 && shift 2 ;;
    --ssl_root_file) SSL_ROOT_FILE=$2 && shift 2 ;;
    --ssl_cert_file) SSL_CERT_FILE=$2 && shift 2 ;;
    --ssl_key_file) SSL_KEY_FILE=$2 && shift 2 ;;
    --dashboard_port) DASHBOARD_PORT=$2 && port_policy_table["dashboard_port"]="FIX" && shift 2 ;;
    --dashboard_grpc_port) DASHBOARD_GRPC_PORT=$2 && port_policy_table["dashboard_grpc_port"]="FIX" && shift 2 ;;
    --collector_port) COLLECTOR_PORT=$2 && port_policy_table["collector_port"]="FIX" && shift 2 ;;
    --meta_service_port) META_SERVICE_PORT=$2 && port_policy_table["meta_service_port"]="FIX" && shift 2 ;;
    --prometheus_address) PROMETHEUS_ADDRESS=$2 && shift 2 ;;
    --prometheus_ssl_enable) PROMETHEUS_SSL_ENABLE=$2 && shift 2 ;;
    --prometheus_ssl_base_path) PROMETHEUS_SSL_BASE_PATH=$2 && shift 2 ;;
    --prometheus_ssl_root_file) PROMETHEUS_SSL_ROOT_FILE=$2 && shift 2 ;;
    --prometheus_ssl_cert_file) PROMETHEUS_SSL_CERT_FILE=$2 && shift 2 ;;
    --prometheus_ssl_key_file) PROMETHEUS_SSL_KEY_FILE=$2 && shift 2 ;;
    --dashboard_ssl_enable) DASHBOARD_SSL_ENABLE=$2 && shift 2 ;;
    --dashboard_ssl_base_path) DASHBOARD_SSL_BASE_PATH=$2 && shift 2 ;;
    --dashboard_ssl_cert_file) DASHBOARD_SSL_CERT_FILE=$2 && shift 2 ;;
    --dashboard_ssl_key_file) DASHBOARD_SSL_KEY_FILE=$2 && shift 2 ;;
    --schedule_relaxed) SCHEDULE_RELAXED=$2 && shift 2 ;;
    --memory_detection_interval) MEMORY_DETECTION_INTERVAL=$2 && shift 2 ;;
    --oom_kill_enable) OOM_KILL_ENABLE=$2 && shift 2 ;;
    --runtime_ds_connect_timeout) RUNTIME_DS_CONNECT_TIMEOUT=$2 && shift 2 ;;
    --oom_kill_control_limit) OOM_KILL_CONTROL_LIMIT=$2 && shift 2 ;;
    --oom_consecutive_detection_count) OOM_CONSECUTIVE_DETECTION_COUNT=$2 && shift 2 ;;
    --max_priority) MAX_PRIORITY=$2 && shift 2 ;;
    --enable_preemption) ENABLE_PREEMPTION=$2 && shift 2 ;;
    --kill_process_timeout_seconds) KILL_PROCESS_TIMEOUT_SECONDS=$2 && shift 2 ;;
    --runtime_home_dir) RUNTIME_USER_HOME_DIR=$2 && shift 2 ;;
    --) shift && break ;;
    *) log_error "Invalid option: $1" && return 1 ;;
    esac
  done
}

function parse_arg_from_env() {
  if [ "X$YR_NODE_ID" != "X" ]; then
    NODE_ID=${YR_NODE_ID:0-20}
  fi
  if [ "X$YR_IP_ADDRESS" != "X" ]; then
    IP_ADDRESS=$YR_IP_ADDRESS
  fi
  if [ "X$YR_ETCD_ADDR_LIST" != "X" ]; then
    ETCD_ADDR_LIST=$YR_ETCD_ADDR_LIST
  fi
}

function check_greater_than_zero() {
  local param_name=$1
  local param_value=$2
  if [[ $param_value -le 0 ]]; then
    log_error "${param_name} should greater than zero, please check your input"
    exit 1
  fi
}

function check_greater_equal_zero() {
  local param_name=$1
  local param_value=$2
  if [[ $param_value -lt 0 ]]; then
    log_error "${param_name} should greater than or equal zero, please check your input"
    exit 1
  fi
}

function check_port_range() {
  local param_name=$1
  local param_value=$2
  if [[ $param_value -le 0 ]] || [[ $param_value -gt 65535 ]]; then
      log_error "${param_name} should in range [1-65535], please check your input"
      exit 1
  fi
}

function check_number_input() {
  check_greater_equal_zero "cpu_reserved" $CPU_RESERVED_FOR_DS_WORKER
  check_greater_equal_zero "min_instance_cpu_size" $MIN_INSTANCE_CPU_SIZE
  check_greater_equal_zero "max_instance_cpu_size" $MAX_INSTANCE_CPU_SIZE
  check_greater_equal_zero "min_instance_memory_size" $MIN_INSTANCE_MEMORY_SIZE
  check_greater_equal_zero "max_instance_memory_size" $MAX_INSTANCE_MEMORY_SIZE
  check_greater_equal_zero "fs_log_rolling_max_size" $FS_LOG_ROLLING_MAX_SIZE
  check_greater_equal_zero "fs_log_rolling_max_files" $FS_LOG_ROLLING_MAX_FILES
  check_greater_equal_zero "fs_log_rolling_retention_days" $FS_LOG_ROLLING_RETENTION_DAYS
  check_greater_equal_zero "fs_log_async_log_buf_secs" $FS_LOG_ASYNC_LOG_BUF_SECS
  check_greater_equal_zero "fs_log_async_log_max_queue_size" $FS_LOG_ASYNC_LOG_MAX_QUEUE_SIZE
  check_greater_equal_zero "fs_log_async_log_thread_count" $FS_LOG_ASYNC_LOG_THREAD_COUNT
  check_greater_equal_zero "ds_log_rolling_max_size" $DS_LOG_ROLLING_MAX_SIZE
  check_greater_equal_zero "ds_log_rolling_max_files" $DS_LOG_ROLLING_MAX_FILES
  check_greater_than_zero "runtime_heartbeat_timeout_ms" $RUNTIME_HEARTBEAT_TIMEOUT_MS
  check_greater_than_zero "runtime_log_rolling_max_size" $RUNTIME_LOG_ROLLING_MAX_SIZE
  check_greater_than_zero "runtime_log_rolling_max_files" $RUNTIME_LOG_ROLLING_MAX_FILES
  check_greater_than_zero "status_collect_interval" $STATUS_COLLECT_INTERVAL
  check_greater_than_zero "runtime_max_heartbeat_timeout_times" $RUNTIME_MAX_HEARTBEAT_TIMEOUT_TIMES
  check_greater_than_zero "runtime_init_call_timeout_seconds" $RUNTIME_INIT_CALL_TIMEOUT_SECONDS
  check_greater_than_zero "system_timeout" $SYSTEM_TIMEOUT
  check_greater_than_zero "function_agent_litebus_thread" $FUNCTION_AGENT_LITEBUS_THREAD
  check_greater_than_zero "function_proxy_litebus_thread" $FUNCTION_PROXY_LITEBUS_THREAD
  check_greater_than_zero "function_master_litebus_thread" $FUNCTION_MASTER_LITEBUS_THREAD
  check_greater_than_zero "ds_spill_size_limit" $DS_SPILL_SIZE_LIMIT
  check_greater_than_zero "ds_rpc_thread_num" $DS_RPC_THREAD_NUM
  check_greater_than_zero "ds_node_dead_timeout_s" $DS_NODE_DEAD_TIMEOUT_S
  check_greater_than_zero "ds_client_dead_timeout_s" $DS_CLIENT_DEAD_TIMEOUT_S
  check_greater_than_zero "ds_node_timeout_s" $DS_NODE_TIMEOUT_S
  check_greater_than_zero "ds_heartbeat_interval_ms" $DS_HEARTBEAT_INTERVAL_MS
  check_greater_than_zero "runtime_conn_timeout_s" $RUNTIME_CONN_TIMEOUT_S
  check_greater_than_zero "runtime_port_num" $RUNTIME_PORT_NUM
  check_greater_than_zero "ds_arena_per_tenant" $DS_ARENA_PER_TENANT
  check_greater_than_zero "fs_health_check_timeout" $FS_HEALTH_CHECK_TIMEOUT
  check_greater_than_zero "fs_health_check_retry_times" $FS_HEALTH_CHECK_RETRY_TIMES
  check_greater_equal_zero "fs_health_check_retry_interval" $FS_HEALTH_CHECK_RETRY_INTERVAL
  check_greater_than_zero "zmq_chunk_sz" $ZMQ_CHUNK_SZ
  check_greater_than_zero "meta_store_max_flush_concurrency" $META_STORE_MAX_FLUSH_CONCURRENCY
  check_greater_than_zero "meta_store_max_flush_batch_size" $META_STORE_MAX_FLUSH_BATCH_SIZE
  check_port_range "runtime_init_port" $RUNTIME_INIT_PORT
  check_port_range "control_plane_port_min" $CONTROL_PLANE_PORT_MIN
  check_port_range "control_plane_port_max" $CONTROL_PLANE_PORT_MAX
  check_port_range "data_plane_port_min" $DATA_PLANE_PORT_MIN
  check_port_range "data_plane_port_max" $DATA_PLANE_PORT_MAX
  check_port_range "function_agent_port" $FUNCTION_AGENT_PORT
  check_port_range "function_proxy_port" $FUNCTION_PROXY_PORT
  check_port_range "function_proxy_grpc_port" $FUNCTION_PROXY_GRPC_PORT
  check_port_range "global_scheduler_port" $GLOBAL_SCHEDULER_PORT
  check_port_range "ds_master_port" $DS_MASTER_PORT
  check_port_range "ds_worker_port" $DS_WORKER_PORT
  check_port_range "etcd_port" $ETCD_PORT
  check_port_range "etcd_peer_port" $ETCD_PEER_PORT
  check_port_range "etcd_proxy_port" $ETCD_PROXY_PORT
  if [[ $MAX_INSTANCE_CPU_SIZE -gt $MAX_VALUE_FOR_CPU_AND_MEM ]]; then
    MAX_INSTANCE_CPU_SIZE=$MAX_VALUE_FOR_CPU_AND_MEM
  fi
  if [[ $MAX_INSTANCE_MEMORY_SIZE -gt $MAX_VALUE_FOR_CPU_AND_MEM ]]; then
    MAX_INSTANCE_MEMORY_SIZE=$MAX_VALUE_FOR_CPU_AND_MEM
  fi
  if [[ $CONTROL_PLANE_PORT_MIN -gt $CONTROL_PLANE_PORT_MAX ]]; then
    log_error "control_plane_port_max should greater than control_plane_port_min"
    exit 1
  fi
  if [[ $DATA_PLANE_PORT_MIN -gt $DATA_PLANE_PORT_MAX ]]; then
    log_error "data_plane_port_max should greater than data_plane_port_min"
    exit 1
  fi
  if [[ $MIN_INSTANCE_CPU_SIZE -gt $MAX_INSTANCE_CPU_SIZE ]]; then
    log_error "min_instance_cpu_size should not greater than max_instance_cpu_size"
    exit 1
  fi
  if [[ $MIN_INSTANCE_MEMORY_SIZE -gt $MAX_INSTANCE_MEMORY_SIZE ]]; then
    log_error "min_instance_memory_sizE should not greater than max_instance_memory_size"
    exit 1
  fi
  if [ "X${ETCD_AUTH_TYPE}" == "XTLS" ]; then
    ENABLE_ETCD_AUTH="true"
  fi
}

function check_tls_config() {
  if [ "X${CURVE_KEY_PATH}" != "X" ]; then
    if ! [[ "${CURVE_KEY_PATH}" = /* ]]; then
      log_error "curve key path '${CURVE_KEY_PATH}' is not an absolute path, please check and retry"
      exit 1
    fi
    if [ ! -d "${CURVE_KEY_PATH}" ]; then
      log_error "curve key path '${CURVE_KEY_PATH}' is not exist"
      exit 1
    fi
    if [ ! -f "${CURVE_KEY_PATH}/worker.key" ]; then
      log_error "curve key path '${CURVE_KEY_PATH}/worker.key' is not exist"
      exit 1
    fi
    if [ ! -f "${CURVE_KEY_PATH}/worker.key_secret" ]; then
      log_error "curve key path '${CURVE_KEY_PATH}/worker.key_secret' is not exist"
      exit 1
    fi
    if [ ! -d "${CURVE_KEY_PATH}/worker_authorized_clients" ]; then
      log_error "curve key path '${CURVE_KEY_PATH}/worker_authorized_clients' is not exist"
      exit 1
    fi
    if [ ! -f "${CURVE_KEY_PATH}/worker_authorized_clients/worker.key" ]; then
      log_error "curve key path '${CURVE_KEY_PATH}/worker_authorized_clients/worker.key' is not exist"
      exit 1
    fi
    if [ ! -f "${CURVE_KEY_PATH}/worker_authorized_clients/client.key" ]; then
      log_error "curve key path '${CURVE_KEY_PATH}/worker_authorized_clients/client.key' is not exist"
      exit 1
    fi
  fi
  if [ "X${SSL_BASE_PATH}" != "X" ]; then
    if ! [[ "${SSL_BASE_PATH}" = /* ]]; then
      log_error "ssl base path '${SSL_BASE_PATH}' is not an absolute path, please check and retry"
      exit 1
    fi
    if [ ! -d "${SSL_BASE_PATH}" ]; then
      log_error "ssl base path '${SSL_BASE_PATH}' is not exist"
      exit 1
    fi
    if [ ! -f "${SSL_BASE_PATH}/${SSL_ROOT_FILE}" ]; then
      log_error "ssl cert '${SSL_BASE_PATH}/${SSL_ROOT_FILE}' is not exist"
      exit 1
    fi
    if [ ! -f "${SSL_BASE_PATH}/${SSL_CERT_FILE}" ]; then
      log_error "ssl cert '${SSL_BASE_PATH}/${SSL_CERT_FILE}' is not exist"
      exit 1
    fi
    if [ ! -f "${SSL_BASE_PATH}/${SSL_KEY_FILE}" ]; then
      log_error "ssl cert '${SSL_BASE_PATH}/${SSL_KEY_FILE}' is not exist"
      exit 1
    fi
  fi
  if [ "X${ETCD_SSL_BASE_PATH}" != "X" ]; then
    if ! [[ "${ETCD_SSL_BASE_PATH}" = /* ]]; then
      log_error "etcd ssl base path '${ETCD_SSL_BASE_PATH}' is not an absolute path, please check and retry"
      exit 1
    fi
    if [ ! -d "${ETCD_SSL_BASE_PATH}" ]; then
      log_error "etcd ssl base path '${ETCD_SSL_BASE_PATH}' is not exist"
      exit 1
    fi
    if [ ! -f "${ETCD_SSL_BASE_PATH}/${ETCD_CA_FILE}" ]; then
      log_error "etcd ssl cert '${ETCD_SSL_BASE_PATH}/${ETCD_CA_FILE}' is not exist"
      exit 1
    fi
    if [ ! -f "${ETCD_SSL_BASE_PATH}/${ETCD_CERT_FILE}" ]; then
      log_error "etcd ssl cert '${ETCD_SSL_BASE_PATH}/${ETCD_CERT_FILE}' is not exist"
      exit 1
    fi
    if [ ! -f "${ETCD_SSL_BASE_PATH}/${ETCD_KEY_FILE}" ]; then
      log_error "etcd ssl cert '${ETCD_SSL_BASE_PATH}/${ETCD_KEY_FILE}' is not exist"
      exit 1
    fi
    if [ ! -f "${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_CERT_FILE}" ]; then
      log_error "etcd ssl cert '${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_CERT_FILE}' is not exist"
      exit 1
    fi
    if [ ! -f "${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_KEY_FILE}" ]; then
      log_error "etcd ssl cert '${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_KEY_FILE}' is not exist"
      exit 1
    fi
  fi
}

function check_runtime_oom_kill_config() {
  if [ "X${OOM_KILL_ENABLE}" != "Xtrue" ] && [ "X${OOM_KILL_ENABLE}" != "Xfalse" ]; then
    log_error "oom_kill_enable can only be 'true' or 'false'"
    return 1
  fi
  if [ "X${OOM_KILL_ENABLE}" == "Xtrue" ]; then
    check_greater_than_zero "memory_detection_interval" $MEMORY_DETECTION_INTERVAL
    if [[ ${MEMORY_DETECTION_INTERVAL} -lt 10 ]]; then
        log_error "memory_detection_interval must be greater than or equal to 10 ms"
        return 1
    fi
    check_greater_than_zero "oom_consecutive_detection_count" $OOM_CONSECUTIVE_DETECTION_COUNT
  fi
}

function check_input() {
  check_number_input
  CPU4COMP=${CPU_ALL}
  MEM4COMP=$((MEM_ALL - MEM_RESERVED))
  if [ "X${ENABLE_MASTER}" = "Xtrue" ]; then
    if [ ${CPU4COMP} -le 0 ]; then
      CPU_ALL=1
      CPU4COMP=1
    fi
    if [ ${MEM4DATA} -le 0 ]; then
      MEM4DATA=2048
    fi
    if [ ${MEM4COMP} -le 0 ]; then
      ((MEM4COMP = MEM4DATA + 1))
      ((MEM_ALL = MEM4DATA + 1))
    fi
  fi
  if [ ${CPU4COMP} -le 0 ]; then
    log_error "cpu for function instances deployment is less then 0"
    return 1
  fi

  if [ ${MEM4COMP} -le 0 ]; then
    log_error "memory for function instances deployment is less then 0"
    return 1
  fi

  if [ ${MEM4DATA} -le 0 ]; then
    log_error "shared memory for data system is less then 0"
    return 1
  fi

  if [ $MEM4COMP -le $MEM4DATA ]; then
    log_error "overall memory should be larger than data system shared memory"
    return 1
  fi

  if [ -z "${INSTALL_DIR_PARENT}" ]; then
    log_error "deploy path is not specified"
    return 1
  fi

  if [ "${IP_ADDRESS}X" = "X" ]; then
    get_ip_auto
    ip_flag='x'$local_ip
    if [ "$ip_flag" == 'x0' ]; then
      echo "The IP address is not configured. This program is exiting"
      return 1
    fi
    if [ "${local_ip}X" = "X" ]; then
      log_error "ip_address is empty, and failed to select an ip automatically, please add argument -a IP_ADDRESS"
      return 1
    else
      IP_ADDRESS=$local_ip
    fi
  else
    if ! check_ip_exist ${IP_ADDRESS} ; then
      log_error "current ip ${IP_ADDRESS} is not belong to this machine, please change it"
      return 1
    fi
  fi
  echo $IP_ADDRESS
  HOST_IP=$IP_ADDRESS
  if [ "X${RUNTIME_HEARTBEAT_ENABLE}" != "Xtrue" ] && [ "X${RUNTIME_HEARTBEAT_ENABLE}" != "Xfalse" ]; then
    log_error "runtime_heartbeat_enable can only be 'true' or 'false'"
    return 1
  fi
  if [ "X${RUNTIME_RECOVER_ENABLE}" != "Xtrue" ] && [ "X${RUNTIME_RECOVER_ENABLE}" != "Xfalse" ]; then
    log_error "runtime_recover_enable can only be 'true' or 'false'"
    return 1
  fi
  if [ "X${RUNTIME_DIRECT_CONNECTION_ENABLE}" != "Xtrue" ] && [ "X${RUNTIME_DIRECT_CONNECTION_ENABLE}" != "Xfalse" ]; then
    log_error "runtime_direct_connection_enable can only be 'true' or 'false'"
    return 1
  fi
  if [ "X${RUNTIME_INSTANCE_DEBUG_ENABLE}" != "Xtrue" ] && [ "X${RUNTIME_INSTANCE_DEBUG_ENABLE}" != "Xfalse" ]; then
    log_error "runtime_instance_debug_enable can only be 'true' or 'false'"
    return 1
  fi
  case "X${NPU_COLLECTION_MODE}" in
        ("Xall" | "Xsfmd" | "Xtopo" | "Xhbm" | "Xcount" | "Xoff" )
          ;;
        (*)
          log_error "npu_collection_mode only support all, count, hbm, sfmd, topo, off"
          return 1
          ;;
  esac
  if [ "X${GPU_COLLECTION_ENABLE}" != "Xtrue" ] && [ "X${GPU_COLLECTION_ENABLE}" != "Xfalse" ]; then
      log_error "gpu_collection_enable can only be 'true' or 'false'"
      return 1
  fi
  if [ "X${DS_SPILL_ENABLE}" != "Xtrue" ] && [ "X${DS_SPILL_ENABLE}" != "Xfalse" ]; then
    log_error "ds_spill_enable can only be 'true' or 'false'"
    return 1
  fi
  if [ "X${DS_LOG_MONITOR_ENABLE}" != "Xtrue" ] && [ "X${DS_LOG_MONITOR_ENABLE}" != "Xfalse" ]; then
    log_error "ds_log_monitor_enable can only be 'true' or 'false'"
    return 1
  fi
  if [ "X${ENABLE_LOSSLESS_DATA_EXIT_MODE}" != "Xtrue" ] && [ "X${ENABLE_LOSSLESS_DATA_EXIT_MODE}" != "Xfalse" ]; then
    log_error "enable_lossless_data_exit_mode can only be 'true' or 'false'"
    return 1
  fi
  if [ "X${ETCD_NO_FSYNC}" != "Xtrue" ] && [ "X${ETCD_NO_FSYNC}" != "Xfalse" ]; then
    log_error "etcd_no_fsync can only be 'true' or 'false'"
    return 1
  fi
  if [ "X${ENABLE_TRACE}" != "Xtrue" ] && [ "X${ENABLE_TRACE}" != "Xfalse" ]; then
     log_error "enable_trace can only be 'true' or 'false'"
     return 1
  fi
  if [ "X${ENABLE_DIS_CONV_CALL_STACK}" != "Xtrue" ] && [ "X${ENABLE_DIS_CONV_CALL_STACK}" != "Xfalse" ]; then
     log_error "enable_dis_conv_call_stack can only be 'true' or 'false'"
     return 1
  fi
  if [ "X${ENABLE_METRICS}" != "Xtrue" ] && [ "X${ENABLE_METRICS}" != "Xfalse" ]; then
     log_error "enable_metrics can only be 'true' or 'false'"
     return 1
  fi
  if [ "X${IS_SCHEDULE_TOLERATE_ABNORMAL}" != "Xtrue" ] && [ "X${IS_SCHEDULE_TOLERATE_ABNORMAL}" != "Xfalse" ]; then
       log_error "is_schedule_tolerate_abnormal can only be 'true' or 'false'"
       return 1
    fi
  if [ "X${ETCD_PROXY_ENABLE}" = "Xtrue" ] ; then
    ETCD_PROXY_ENABLE="TRUE"
  fi
  if [ "X${STATUS_COLLECT_ENABLE}" = "Xtrue" ]; then
    STATUS_COLLECT_ENABLE="TRUE"
  fi
  if [[ ${DS_NODE_DEAD_TIMEOUT_S} -lt  ${DS_NODE_TIMEOUT_S} ]]; then
    log_error "ds_node_dead_timeout_s must be greater than or equal to the value of ds_node_timeout_s"
    return 1
  fi
  if [[ $SYSTEM_TIMEOUT -gt 12000 ]]; then
    log_info "system timeout : ${SYSTEM_TIMEOUT} ms"
  else
    SYSTEM_TIMEOUT=12000
  fi
  if [[ $ETCD_COMPACT_RETENTION -gt 1000 ]]; then
    log_info "etcd compact retention : ${ETCD_COMPACT_RETENTION}"
  else
    ETCD_COMPACT_RETENTION=10000
  fi
  case "X${NPU_COLLECTION_MODE}" in
      ("Xall" | "Xsfmd" | "Xtopo" | "Xhbm")
          command -v npu-smi &>/dev/null
          local ret=$?
          if [ ${ret} != 0 ]; then
            log_warning "npu collection is enabled, but npu-smi command is not found"
          fi
          ;;
  esac
  MEM4COMP=$((MEM4COMP - MEM4DATA))
  check_tls_config
  check_runtime_oom_kill_config
  if [ "X${ETCD_TABLE_PREFIX}" != "X" ]; then
    if ! [[ "X${ETCD_TABLE_PREFIX}" = /*  ]]; then
      UNIQUE_KEY_PREFIX="/${ETCD_TABLE_PREFIX}"
    else
      UNIQUE_KEY_PREFIX="${ETCD_TABLE_PREFIX}"
    fi
  fi
  return 0
}

function init_log_path() {
  if [ "${LOG_ROOT}X" = "defaultX" ]; then
    if [ "X${ENABLE_MASTER}" = "Xtrue" ]; then
      LOG_ROOT="${INSTALL_DIR_PARENT}/log"
    else
      LOG_ROOT="${INSTALL_DIR_PARENT}/${NODE_ID}/log"
    fi
  fi
  if ! [[ "${LOG_ROOT}" = /* ]]; then
    log_error "root log path '${LOG_ROOT}' is not an absolute path, please check and retry"
    return 1
  fi
  if [ "X${FS_LOG_PATH}" != "X" ] && [[ "${FS_LOG_PATH}" = *..* ]]; then
    log_error "function system log path '${FS_LOG_PATH}' cannot contains '..'"
    return 1
  fi
  if [ "X${DS_LOG_PATH}" != "X" ] && [[ "${DS_LOG_PATH}" = *..* ]]; then
    log_error "data system log path '${DS_LOG_PATH}' cannot contains '..'"
    return 1
  fi
  if [ "X${ETCD_LOG_PATH}" != "X" ] && [[ "${ETCD_LOG_PATH}" = *..* ]]; then
    log_error "etcd log path '${ETCD_LOG_PATH}' cannot contains '..'"
    return 1
  fi
  if [ "X${RUNTIME_LOG_PATH}" != "X" ] && [[ "${RUNTIME_LOG_PATH}" = *..* ]]; then
    log_error "runtime log path '${RUNTIME_LOG_PATH}' cannot contains '..'"
    return 1
  fi
  FS_LOG_PATH="${LOG_ROOT}/${FS_LOG_PATH}"
  DS_LOG_PATH="${LOG_ROOT}/${DS_LOG_PATH}"
  RUNTIME_LOG_PATH="${LOG_ROOT}/${RUNTIME_LOG_PATH}"
  ETCD_LOG_PATH="${LOG_ROOT}/${ETCD_LOG_PATH}"
  local exception_log_dir="${RUNTIME_LOG_PATH}"/exception
  # create log dir
  if  [ ! -d "${LOG_ROOT}" ] &&  ! mkdir -m 750 -p "${LOG_ROOT}" ; then
    log_error "root log path '${LOG_ROOT}' cannot be created, please check"
    return 1
  fi
  if  [ ! -d "${FS_LOG_PATH}" ] && ! mkdir -m 750 -p "${FS_LOG_PATH}" ; then
    log_error "function system log path '${FS_LOG_PATH}' cannot be created, please check"
    return 1
  fi
  if  [ ! -d "${DS_LOG_PATH}" ] && ! mkdir -m 750 -p "${DS_LOG_PATH}" ; then
    log_error "data system log path '${DS_LOG_PATH}' cannot be created, please check"
    return 1
  fi
  if  [ ! -d "${ETCD_LOG_PATH}" ] && ! mkdir -m 750 -p "${ETCD_LOG_PATH}" ; then
    log_error "etcd log path '${ETCD_LOG_PATH}' cannot be created, please check"
    return 1
  fi
  if  [ ! -d "${RUNTIME_LOG_PATH}" ] && ! mkdir -m 750 -p "${RUNTIME_LOG_PATH}" ; then
    log_error "runtime log path '${RUNTIME_LOG_PATH}' cannot be created, please check"
    return 1
  fi
  if  [ ! -d "${exception_log_dir}" ] && ! mkdir -m 750 -p "${exception_log_dir}" ; then
    log_error "runtime log exception path '${exception_log_dir}' cannot be created, please check"
    return 1
  fi
  # check log directory permission
  if echo "test" > ${LOG_ROOT}/test.log 2>/dev/null; then
    rm -f ${LOG_ROOT}/test.log
  else
    log_error "root log path '${LOG_ROOT}' has not permission to write, please check"
    return 1
  fi
  if echo "test" > ${FS_LOG_PATH}/test.log 2>/dev/null; then
    rm -f ${FS_LOG_PATH}/test.log
  else
    log_error "function system log path '${FS_LOG_PATH}' has not permission to write, please check"
    return 1
  fi
  if echo "test" > ${DS_LOG_PATH}/test.log 2>/dev/null; then
    rm -f ${DS_LOG_PATH}/test.log
  else
    log_error "data system log path '${DS_LOG_PATH}' has not permission to write, please check"
    return 1
  fi
  if echo "test" > ${ETCD_LOG_PATH}/test.log 2>/dev/null; then
    rm -f ${ETCD_LOG_PATH}/test.log
  else
    log_error "etcd log path '${ETCD_LOG_PATH}' has not permission to write, please check"
    return 1
  fi
  if echo "test" > ${RUNTIME_LOG_PATH}/test.log 2>/dev/null; then
    rm -f ${RUNTIME_LOG_PATH}/test.log
  else
    log_error "runtime log path '${RUNTIME_LOG_PATH}' has not permission to write, please check"
    return 1
  fi
  if echo "test" > ${exception_log_dir}/test.log 2>/dev/null; then
    rm -f ${exception_log_dir}/test.log
  else
    log_error "runtime exception path '${exception_log_dir}' has not permission to write, please check"
    return 1
  fi
  LOG_ROOT=$(readlink -m "${LOG_ROOT}")
  FS_LOG_PATH=$(readlink -m "${FS_LOG_PATH}")
  DS_LOG_PATH=$(readlink -m "${DS_LOG_PATH}")
  ETCD_LOG_PATH=$(readlink -m "${ETCD_LOG_PATH}")
  RUNTIME_LOG_PATH=$(readlink -m "${RUNTIME_LOG_PATH}")
}

function process_log_config() {
  DS_DEBUG_LOG_LEVEL=0
  if [ "$DS_LOG_LEVEL_STR" = "ERROR" ]; then
    DS_LOG_LEVEL=2
  elif [ "$DS_LOG_LEVEL_STR" = "WARN" ]; then
    DS_LOG_LEVEL=1
  elif [ "$DS_LOG_LEVEL_STR" = "INFO" ]; then
    DS_LOG_LEVEL=0
  elif [ "$DS_LOG_LEVEL_STR" = "DEBUG" ]; then
    DS_LOG_LEVEL=0
    DS_DEBUG_LOG_LEVEL=1
  fi


  DATA_PLANE_INSTALL_DIR="${INSTALL_DIR_PARENT}"/"${NODE_ID}"
  [ "${DS_SPILL_DIRECTORY}X" = "X" ] && DS_SPILL_DIRECTORY="${DATA_PLANE_INSTALL_DIR}/data_system/spill"
  FS_LOG_CONFIG="${FS_LOG_CONFIG//\{\{logLevel\}\}/$FS_LOG_LEVEL}"
  FS_LOG_CONFIG="${FS_LOG_CONFIG//\{\{logCompressEnable\}\}/$FS_LOG_COMPRESS_ENABLE}"
  FS_LOG_CONFIG="${FS_LOG_CONFIG//\{\{logRollingMaxSize\}\}/$FS_LOG_ROLLING_MAX_SIZE}"
  FS_LOG_CONFIG="${FS_LOG_CONFIG//\{\{logRollingMaxFiles\}\}/$FS_LOG_ROLLING_MAX_FILES}"
  FS_LOG_CONFIG="${FS_LOG_CONFIG//\{\{logRollingRetentionDays\}\}/$FS_LOG_ROLLING_RETENTION_DAYS}"
  FS_LOG_CONFIG="${FS_LOG_CONFIG//\{\{logAsyncBufSecs\}\}/$FS_LOG_ASYNC_LOG_BUF_SECS}"
  FS_LOG_CONFIG="${FS_LOG_CONFIG//\{\{logAsyncMaxQueueSize\}\}/$FS_LOG_ASYNC_LOG_MAX_QUEUE_SIZE}"
  FS_LOG_CONFIG="${FS_LOG_CONFIG//\{\{logAsyncThreadCount\}\}/$FS_LOG_ASYNC_LOG_THREAD_COUNT}"
  FS_LOG_CONFIG="${FS_LOG_CONFIG//\{\{logAlsologtostderr\}\}/$FS_ALSO_LOG_TO_STDERR}"
  FS_LOG_CONFIG="${FS_LOG_CONFIG//\{\{logConfigPath\}\}/$FS_LOG_PATH}"
  log_info "function system log config: ${FS_LOG_CONFIG}"

  FAAS_LOG_CONFIG="${FAAS_LOG_CONFIG//\{\{logLevel\}\}/$FS_LOG_LEVEL}"
  FAAS_LOG_CONFIG="${FAAS_LOG_CONFIG//\{\{logConfigPath\}\}/$FS_LOG_PATH}"
  log_info "faas log config: ${FAAS_LOG_CONFIG}"

  LOG_ROTATE_CONFIG="${LOG_ROTATE_CONFIG//\{\{logConfigPath\}\}/$RUNTIME_LOG_PATH}"
  LOG_ROTATE_CONFIG="${LOG_ROTATE_CONFIG//\{\{logRollingMaxSize\}\}/$RUNTIME_LOG_ROLLING_MAX_SIZE}"
  LOG_ROTATE_CONFIG="${LOG_ROTATE_CONFIG//\{\{logRollingMaxFiles\}\}/$RUNTIME_LOG_ROLLING_MAX_FILES}"
  LOG_ROTATE_CONFIG="${LOG_ROTATE_CONFIG//\{\{logRollingRetentionDays\}\}/$FS_LOG_ROLLING_RETENTION_DAYS}"

  local python_log_config="${BASE_DIR}/../../runtime/service/python/config/python-runtime-log.json"
  if [ -f "$python_log_config" ]; then
    sed -i "s|{{logLevel}}|${RUNTIME_LOG_LEVEL}|g" "$python_log_config"
  fi
}

function parse_master_info() {
  if [ "X${ENABLE_MASTER}" = "Xtrue" ]; then
    [ -z "${ETCD_IP}" ] && ETCD_IP=$IP_ADDRESS
    [ -z "${FUNCTION_MASTER_IP}" ] && FUNCTION_MASTER_IP=$IP_ADDRESS
    [ -z "${DS_MASTER_IP}" ] && DS_MASTER_IP=$IP_ADDRESS
    [ -n "${MASTER_INFO_OUT_FILE}" ] && [ -f "${MASTER_INFO_OUT_FILE}" ] && rm -f "${MASTER_INFO_OUT_FILE}"
    return 0
  fi
  if [ "X${MASTER_INFO}" != "X" ]; then
    FUNCTION_MASTER_IP=$( expr match "${MASTER_INFO}" '.*master_ip:\([0-9|\.]*\)' )
    ETCD_IP=$( expr match "${MASTER_INFO}" '.*etcd_ip:\([0-9|\.]*\)' )
    GLOBAL_SCHEDULER_PORT=$( expr match "${MASTER_INFO}" '.*global_scheduler_port:\([0-9|\.]*\)' )
    DS_MASTER_PORT=$( expr match "${MASTER_INFO}" '.*ds_master_port:\([0-9|\.]*\)' )
    DS_MASTER_IP=$FUNCTION_MASTER_IP
    ETCD_PORT=$( expr match "${MASTER_INFO}" '.*etcd_port:\([0-9|\.]*\)' )
    ETCD_PEER_PORT=$( expr match "${MASTER_INFO}" '.*etcd_peer_port:\([0-9|\.]*\)' )
    local ds_worker_port=$( expr match "${MASTER_INFO}" '.*ds-worker:\([0-9|\.]*\)' )
    MASTER_USED_DS_PORT="${DS_MASTER_PORT} ${ds_worker_port}"
    set_master_used_ds_port
  fi
  if [ "X${ENABLE_MULTI_MASTER}" = "Xtrue" ] && [ -z "${FUNCTION_MASTER_IP}" ]; then
    FUNCTION_MASTER_IP=$IP_ADDRESS
    return 0
  fi
  if  [ -z "${FUNCTION_MASTER_IP}" ]; then
    log_error "master_ip can not be empty"
    return 1
  fi
  if [ -n "${FUNCTION_MASTER_IP}" ] &&  [ -z "${ETCD_IP}" ]; then
    ETCD_IP=${FUNCTION_MASTER_IP}
  fi
  if  [ -z "${ETCD_IP}" ]; then
    log_error "etcd_ip can not be empty"
    return 1
  fi
  if  [ -z "${GLOBAL_SCHEDULER_PORT}" ]; then
    log_error "global_scheduler_port can not be empty"
    return 1
  fi
  if  [ "X${ENABLE_DISTRIBUTED_MASTER}" = "Xfalse" ] && [ -z "${DS_MASTER_PORT}" ]; then
    log_error "ds_master_port can not be empty when enable_distributed_master is false"
    return 1
  fi
  if  [ -z "${ETCD_PORT}" ]; then
    log_error "etcd_port can not be empty"
    return 1
  fi
  if  [ -z "${ETCD_PEER_PORT}" ]; then
    log_error "etcd_peer_port can not be empty"
    return 1
  fi
}

function parse_etcd_cluster()
{
  ETCD_CLUSTER_ADDRESS="${ETCD_IP}:${ETCD_PORT}"
  if [ "${ETCD_AUTH_TYPE}" == "TLS" ]; then
    ETCD_HTTP_PROTOCOL="https://"
  fi
  ETCD_ENDPOINTS_ADDR="${ETCD_HTTP_PROTOCOL}${ETCD_IP}:${ETCD_PORT}"
  if [ "X${ENABLE_MULTI_MASTER}" != "Xtrue" ]; then
    if [ "X${ETCD_MODE}" == "Xoutter" ]; then
      log_error "only multi-master mode supports outter etcd, please add '-e' enable multi master"
      return 1
    fi
    return 0
  fi
  # set ds-worker distributed master to true
  ENABLE_DISTRIBUTED_MASTER="true"
  ETCD_NO_FSYNC="false"
  ELECTION_MODE="etcd"
  if [ "X${ETCD_ADDR_LIST}" = "X" ]; then
    log_warning "etcd_addr_list is empty, set it to local ip ${IP_ADDRESS}"
    ETCD_ADDR_LIST=$IP_ADDRESS
  fi
  local cur_etcd_port="X"
  local cur_etcd_peer_port="X"
  local etcd_array=(${ETCD_ADDR_LIST//,/ })
  local index=0
  ETCD_CLUSTER_ADDRESS=""
  ETCD_ENDPOINTS_ADDR=""
  for etcd_item in ${etcd_array[@]}
  do
    local add_arr=(${etcd_item//:/ })
    local len=${#add_arr[@]}
    if [[ $len -eq 0 ]]; then
      log_error "etcd_addr_list format is error, ${etcd_item}"
      return 1
    fi
    local etcd_name="etcd${index}"
    ((index++));
    ((ETCD_NODE_NUM++))
    local etcd_ip=${add_arr[0]}
    local etcd_port=${ETCD_PORT}
    local etcd_peer_port=${ETCD_PEER_PORT}
    if [[ $len -ge 2 ]] && [ "X${add_arr[1]}" != "X" ]; then
      etcd_port=${add_arr[1]}
    fi
    if [[ $len -ge 3 ]] && [ "X${add_arr[2]}" != "X" ]; then
      etcd_peer_port=${add_arr[2]}
    fi
    ETCD_CLUSTER_ADDRESS="${ETCD_CLUSTER_ADDRESS},${etcd_ip}:${etcd_port}"
    ETCD_INITIAL_CLUSTER="${ETCD_INITIAL_CLUSTER},${etcd_name}=${ETCD_HTTP_PROTOCOL}${etcd_ip}:${etcd_peer_port}"
    ETCD_ENDPOINTS_ADDR="${ETCD_ENDPOINTS_ADDR},${ETCD_HTTP_PROTOCOL}${etcd_ip}:${etcd_port}"
    ETCD_ADDR_LIST_FOR_MASTER_INFO="${ETCD_ADDR_LIST_FOR_MASTER_INFO},etcd_addr_list:${etcd_ip}"
    if [ ${etcd_ip} = $IP_ADDRESS ]; then
      cur_etcd_port=${etcd_port}
      cur_etcd_peer_port=${etcd_peer_port}
      ETCD_NAME=$etcd_name
    fi
  done
  ETCD_INITIAL_CLUSTER=${ETCD_INITIAL_CLUSTER#*,}
  ETCD_CLUSTER_ADDRESS=${ETCD_CLUSTER_ADDRESS#*,}
  ETCD_ENDPOINTS_ADDR=${ETCD_ENDPOINTS_ADDR#*,}
  ETCD_ADDR_LIST_FOR_MASTER_INFO=${ETCD_ADDR_LIST_FOR_MASTER_INFO#*,}
  if [ "X${ETCD_MODE}" != "Xoutter" ] && [ "X${ENABLE_MASTER}" = "Xtrue" ] && [ "${cur_etcd_port}" = "X" ]; then
    log_error "current ip(${IP_ADDRESS}) not in etcd_addr_list"
    return 1
  fi
  if [ "$cur_etcd_port" != "X" ]; then
    ETCD_PORT="$cur_etcd_port"
  fi
  if [ "$cur_etcd_peer_port" != "X" ]; then
    ETCD_PEER_PORT="$cur_etcd_peer_port"
  fi
}

function parse_runtime_tls_config() {
  if [ "X${SSL_ENABLE}" = "Xtrue" ]; then
    ENABLE_MTLS=${SSL_ENABLE}
    if [ "X${SSL_BASE_PATH}" != "X" ]; then
      PRIVATE_KEY_PATH="${SSL_BASE_PATH}/module.key"
      CERTIFICATE_FILE_PATH="${SSL_BASE_PATH}/module.crt"
      VERIFY_FILE_PATH="${SSL_BASE_PATH}/ca.crt"
    fi
  fi
  if [ "X${RUNTIME_DS_ENCRYPT_ENABLE}" = "Xtrue" ]; then
    if [ "X${CURVE_KEY_PATH}" != "X" ]; then
      ENCRYPT_RUNTIME_PUBLIC_KEY_CONTEXT=$(cat "${CURVE_KEY_PATH}/client.key")
      ENCRYPT_RUNTIME_PRIVATE_KEY_CONTEXT=$(cat "${CURVE_KEY_PATH}/client.key_secret")
      ENCRYPT_DS_PUBLIC_KEY_CONTEXT=$(cat "${CURVE_KEY_PATH}/worker.key")
    fi
  fi
}

function export_config() {
  export FUNCTION_MASTER_IP NODE_ID FUNCTION_AGENT_ALIAS FS_LOG_PATH DS_LOG_LEVEL DS_DEBUG_LOG_LEVEL HOST_IP
  export ENABLE_MASTER IP_ADDRESS SERVICES_PATH DS_WORKER_UNIQUE_ENABLE INSTALL_DIR_PARENT DATA_PLANE_INSTALL_DIR
  export CONTROL_PLANE_PORT_MIN CONTROL_PLANE_PORT_MAX DATA_PLANE_PORT_MIN DATA_PLANE_PORT_MAX
  export MASTER_INFO_OUT_FILE CPU4COMP MEM4COMP MEM4DATA FS_LOG_LEVEL FS_LOG_COMPRESS_ENABLE CPU_ALL MEM_ALL
  export LOG_ROOT DS_LOG_PATH ETCD_LOG_PATH STD_LOG_SUFFIX CPU_RESERVED_FOR_DS_WORKER MAX_INSTANCE_CPU_SIZE MAX_INSTANCE_MEMORY_SIZE
  export DS_LOG_ROLLING_MAX_SIZE DS_LOG_ROLLING_MAX_FILES DS_RPC_THREAD_NUM DS_CLIENT_DEAD_TIMEOUT_S DS_NODE_DEAD_TIMEOUT_S
  export RUNTIME_LOG_PATH RUNTIME_LOG_LEVEL DS_LOG_LEVEL_STR MIN_INSTANCE_CPU_SIZE MIN_INSTANCE_MEMORY_SIZE
  export ACCESSOR_HTTP_PORT ACCESSOR_GRPC_PORT FUNCTION_AGENT_PORT FUNCTION_PROXY_PORT FUNCTION_PROXY_GRPC_PORT
  export RUNTIME_INIT_PORT DS_WORKER_PORT RUNTIME_CONN_TIMEOUT_S
  export RUNTIME_INIT_CALL_TIMEOUT_SECONDS IS_SCHEDULE_TOLERATE_ABNORMAL STATE_STORAGE_TYPE
  export MERGE_PROCESS_ENABLE DRIVER_GATEWAY_ENABLE
  export NPU_COLLECTION_MODE GPU_COLLECTION_ENABLE
  export GLOBAL_SCHEDULER_PORT METRICS_COLLECTOR_TYPE ETCD_PROXY_ENABLE
  export RUNTIME_METRICS_CONFIG
  export RUNTIME_HEARTBEAT_ENABLE RUNTIME_HEARTBEAT_TIMEOUT_MS RUNTIME_MAX_HEARTBEAT_TIMEOUT_TIMES RUNTIME_RECOVER_ENABLE RUNTIME_DIRECT_CONNECTION_ENABLE RUNTIME_INSTANCE_DEBUG_ENABLE
  # datasystem
  export RUNTIME_PORT_NUM RUNTIME_DEFAULT_CONFIG SYS_FUNC_RETRY_PERIOD DS_MASTER_IP DS_MASTER_PORT DS_SPILL_DIRECTORY DS_SPILL_ENABLE
  export DS_NODE_TIMEOUT_S DS_HEARTBEAT_INTERVAL_MS
  export DS_ARENA_PER_TENANT DS_ENABLE_FALLOCATE DS_ENABLE_HUGE_TLB DS_ENABLE_THP
  export DS_L2_CACHE_TYPE DS_SFS_PATH DS_LOG_MONITOR_ENABLE ZMQ_CHUNK_SZ ENABLE_LOSSLESS_DATA_EXIT_MODE
  export DS_SPILL_SIZE_LIMIT
  # etcd
  export ETCD_IP ETCD_PORT ETCD_PEER_PORT ETCD_PROXY_NUMS ETCD_PROXY_NUMS ETCD_PROXY_PORT ETCD_NO_FSYNC
  # trace and metrics
  export ENABLE_TRACE TRACE_CONFIG RUNTIME_TRACE_CONFIG ENABLE_METRICS METRICS_CONFIG METRICS_CONFIG_FILE STATUS_COLLECT_ENABLE STATUS_COLLECT_INTERVAL ENABLE_DIS_CONV_CALL_STACK
  export FUNCTION_AGENT_LITEBUS_THREAD FUNCTION_PROXY_LITEBUS_THREAD FUNCTION_MASTER_LITEBUS_THREAD
  export SYSTEM_TIMEOUT FUNCTION_PROXY_UNIQUE_ENABLE
  export ENABLE_META_STORE ENABLE_PERSISTENCE META_STORE_MODE META_STORE_EXCLUDED_KEYS
  export ENABLE_JEMALLOC JEMALLOC_LIB_PATH ENABLE_INHERIT_ENV IS_PARTIAL_WATCH_INSTANCES FUNCTION_PROXY_UNREGISTER_WHILE_STOP
  export DS_MAX_CLIENT_NUM DS_MEMORY_RECLAMATION_TIME_SECOND
  export IS_PROTOMSG_TO_RUNTIME MASSIF_ENABLE MASTER_USED_DS_PORT STS_CONFIG
  export BLOCK CUSTOM_RESOURCES LABELS SEPARATED_REDIRECT_RUNTIME_STD USER_LOG_EXPORT_MODE
  export ENABLE_DISTRIBUTED_MASTER DISABLE_NC_CHECK
  export SCHEDULE_RELAXED MAX_PRIORITY ENABLE_PREEMPTION KILL_PROCESS_TIMEOUT_SECONDS
  export RUNTIME_DS_CONNECT_TIMEOUT
  export MEMORY_DETECTION_INTERVAL OOM_KILL_ENABLE OOM_KILL_CONTROL_LIMIT OOM_CONSECUTIVE_DETECTION_COUNT
  export RUNTIME_USER_HOME_DIR CACHE_STORAGE_AUTH_TYPE CACHE_STORAGE_AUTH_ENABLE
  export PRIVATE_KEY_PATH CERTIFICATE_FILE_PATH VERIFY_FILE_PATH SSL_BASE_PATH

  # dashboard
  export ENABLE_DASHBOARD DASHBOARD_PORT DASHBOARD_GRPC_PORT
  export PROMETHEUS_ADDRESS PROMETHEUS_SSL_ENABLE PROMETHEUS_SSL_BASE_PATH PROMETHEUS_SSL_ROOT_FILE PROMETHEUS_SSL_CERT_FILE PROMETHEUS_SSL_KEY_FILE
  export DASHBOARD_SSL_ENABLE DASHBOARD_SSL_BASE_PATH DASHBOARD_SSL_CERT_FILE DASHBOARD_SSL_KEY_FILE
  # collector
  export ENABLE_COLLECTOR COLLECTOR_PORT
  # meta_service
  export ENABLE_META_SERVICE META_SERVICE_PORT
  # faas
  export ENABLE_FAAS_FRONTEND FAAS_FRONTEND_HTTP_PORT FAAS_FRONTEND_GRPC_PORT ENABLE_FUNCTION_SCHEDULER
}

function main() {
  parse_opt "$@"
  [ $? -ne 0 ] && help_msg && return 1
  parse_arg_from_env
  [ $? -ne 0 ] && help_msg && return 1
  check_input
  [ $? -ne 0 ] && help_msg && return 1
  parse_master_info
  [ $? -ne 0 ] && help_msg && return 1
  parse_etcd_cluster
  [ $? -ne 0 ] && help_msg && return 1
  init_log_path
  [ $? -ne 0 ] && return 1
  if [ "X${ONLY_CHECK_PARAM}" = "Xtrue" ]; then
    [ ! -d "${INSTALL_DIR_PARENT}" ] && mkdir -p "${INSTALL_DIR_PARENT}"
    return 0
  fi
  parse_runtime_tls_config
  [ $? -ne 0 ] && help_msg && return 1
  process_log_config
  export_config
}

main "$@"