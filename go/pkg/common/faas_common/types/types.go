/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package types -
package types

// HTTPResponse is general http response
type HTTPResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// InnerInstanceData is the function instance data stored in ETCD
type InnerInstanceData struct {
	IP              string           `json:"ip"`
	Port            string           `json:"port"`
	Status          string           `json:"status"`
	P2pPort         string           `json:"p2pPort"`
	GrpcPort        string           `json:"grpcPort,omitempty"`
	NodeIP          string           `json:"nodeIP,omitempty"`
	NodePort        string           `json:"nodePort,omitempty"`
	NodeName        string           `json:"nodeName,omitempty"`
	NodeID          string           `json:"nodeID,omitempty"`
	Applier         string           `json:"applier,omitempty"` // silimar to OwnerIP
	OwnerIP         string           `json:"ownerIP,omitempty"`
	FuncSig         string           `json:"functionSignature,omitempty"`
	Reserved        bool             `json:"reserved,omitempty"`
	CPU             int64            `json:"cpu,omitempty"`
	Memory          int64            `json:"memory,omitempty"`
	GroupID         string           `json:"groupID,omitempty"`
	StackID         string           `json:"stackID,omitempty"`
	CustomResources map[string]int64 `json:"customResources,omitempty" valid:"optional"`
}

// LogTankService -
type LogTankService struct {
	GroupID  string `json:"logGroupId" valid:",optional"`
	StreamID string `json:"logStreamId" valid:",optional"`
}

// TraceService -
type TraceService struct {
	TraceAK     string `json:"tracing_ak" valid:",optional"`
	TraceSK     string `json:"tracing_sk" valid:",optional"`
	ProjectName string `json:"project_name" valid:",optional"`
}

// Initializer include initializer handler and timeout
type Initializer struct {
	Handler string `json:"initializer_handler" valid:",optional"`
	Timeout int64  `json:"initializer_timeout" valid:",optional"`
}

// FuncMountConfig function mount config
type FuncMountConfig struct {
	FuncMountUser FuncMountUser `json:"mount_user" valid:",optional"`
	FuncMounts    []FuncMount   `json:"func_mounts" valid:",optional"`
}

// FuncMountUser function mount user
type FuncMountUser struct {
	UserID  int `json:"user_id" valid:",optional"`
	GroupID int `json:"user_group_id" valid:",optional"`
}

// FuncMount function mount
type FuncMount struct {
	MountType      string `json:"mount_type" valid:",optional"`
	MountResource  string `json:"mount_resource" valid:",optional"`
	MountSharePath string `json:"mount_share_path" valid:",optional"`
	LocalMountPath string `json:"local_mount_path" valid:",optional"`
	Status         string `json:"status" valid:",optional"`
}

// Role include x_role and app_x_role
type Role struct {
	XRole    string `json:"xrole" valid:",optional"`
	AppXRole string `json:"app_xrole" valid:",optional"`
}

// FunctionDeploymentSpec define function deployment spec
type FunctionDeploymentSpec struct {
	BucketID  string `json:"bucket_id"`
	ObjectID  string `json:"object_id"`
	Layers    string `json:"layers"`
	DeployDir string `json:"deploydir"`
}

// InstanceResource describes the cpu and memory info of an instance
type InstanceResource struct {
	CPU             string           `json:"cpu"`
	Memory          string           `json:"memory"`
	CustomResources map[string]int64 `json:"customresources"`
}

// Worker define a worker
type Worker struct {
	Instances       []*Instance `json:"instances"`
	FunctionName    string      `json:"functionname"`
	FunctionVersion string      `json:"functionversion"`
	Tenant          string      `json:"tenant"`
	Business        string      `json:"business"`
}

// Instance define a instance
type Instance struct {
	IP             string `json:"ip"`
	Port           string `json:"port"`
	GrpcPort       string `json:"grpcPort"`
	InstanceID     string `json:"instanceID,omitempty"`
	DeployedIP     string `json:"deployed_ip"`
	DeployedNode   string `json:"deployed_node"`
	DeployedNodeID string `json:"deployed_node_id"`
	TenantID       string `json:"tenant_id"`
}

// InstanceCreationRequest is used to create instance
type InstanceCreationRequest struct {
	LogicInstanceID string           `json:"logicInstanceID"`
	FuncName        string           `json:"functionName"`
	Applier         string           `json:"applier"`
	DeployNode      string           `json:"deployNode"`
	Business        string           `json:"business"`
	TenantID        string           `json:"tenantID"`
	Version         string           `json:"version"`
	OwnerIP         string           `json:"ownerIP"`
	TraceID         string           `json:"traceID"`
	TriggerFlag     string           `json:"triggerFlag"`
	VersionUrn      string           `json:"versionUrn"`
	CPU             int64            `json:"cpu"`
	Memory          int64            `json:"memory"`
	GroupID         string           `json:"groupID"`
	StackID         string           `json:"stackID"`
	CustomResources map[string]int64 `json:"customResources,omitempty" valid:"optional"`
}

// InstanceCreationSuccessResponse is the struct returned by workermanager upon successful instance creation
type InstanceCreationSuccessResponse struct {
	HTTPResponse
	Worker   *Worker   `json:"worker"`
	Instance *Instance `json:"instance"`
}

// InstanceDeletionRequest is used to delete instance
type InstanceDeletionRequest struct {
	InstanceID  string `json:"instanceID"`
	FuncName    string `json:"functionName"`
	FuncVersion string `json:"functionVersion"`
	TenantID    string `json:"tenantID"`
	BusinessID  string `json:"businessID"`
	Applier     string `json:"applier"`
	Force       bool   `json:"force"`
}

// InstanceDeletionResponse is the struct returned by workermanager upon successful instance deletion
type InstanceDeletionResponse struct {
	HTTPResponse
	Reserved bool `json:"reserved"`
}

// HookArgs keeps args of hook
type HookArgs struct {
	FuncArgs        []byte // Call() request in worker
	SrcTenant       string
	DstTenant       string
	StateID         string
	LogType         string
	StateKey        string // for trigger state call
	FunctionVersion string // for trigger state call
	ExternalRequest bool   // for trigger state call
	ServiceID       string
	TraceID         string
	InvokeType      string
}

// ResourceStack stores properties of resource stack
type ResourceStack struct {
	StackID         string           `json:"id" valid:"required"`
	CPU             int64            `json:"cpu" valid:"required"`
	Mem             int64            `json:"mem" valid:"required"`
	CustomResources map[string]int64 `json:"customResources,omitempty" valid:"optional"`
}

// ResourceGroup stores properties of resource group
type ResourceGroup struct {
	GroupID         string                     `json:"id" valid:"required"`
	DeployOption    string                     `json:"deployOption" valid:"required"`
	GroupState      string                     `json:"groupState" valid:"required"`
	ResourceStacks  []ResourceStack            `json:"resourceStacks" valid:"required"`
	ScheduledStacks map[string][]ResourceStack `json:"scheduledStacks,omitempty" valid:"optional"`
}

// AffinityInfo is data affinity information
type AffinityInfo struct {
	AffinityRequest AffinityRequest
	AffinityNode    string // if AffinityNode is not empty, the affinity node has been calculated
	NeedToForward   bool
}

// AffinityRequest is affinity request parameter
type AffinityRequest struct {
	Strategy  string   `json:"strategy"`
	ObjectIDs []string `json:"object_ids"`
}

// GroupInfo stores groupID and stackID
type GroupInfo struct {
	GroupID string `json:"groupID"`
	StackID string `json:"stackID"`
}

// InvokeOption contains invoke options
type InvokeOption struct {
	AffinityRequest  AffinityRequest
	GroupInfo        GroupInfo
	ResourceMetaData map[string]float32
}

// ScheduleConfig defines schedule config
type ScheduleConfig struct {
	Policy                         int     `json:"policy" valid:"optional"`
	ForwardScheduleFirst           bool    `json:"forwardScheduleResourceNotEnough" valid:"optional"`
	SleepingMemThreshold           float32 `json:"sleepingMemoryThreshold" valid:"optional"`
	SelectInstanceToSleepingPolicy string  `json:"selectInstanceToSleepingPolicy" valid:"optional"`
}

// MetricsData shows the quantities of a specific resource
type MetricsData struct {
	TotalResource float32 `json:"totalResource"`
	InUseResource float32 `json:"inUseResource"`
}

// ResourceMetrics contains several resources' MetricsData
type ResourceMetrics map[string]MetricsData

// WorkerMetrics stores metrics used for scheduler
type WorkerMetrics struct {
	SystemResources ResourceMetrics
	// key levels: functionUrn instanceID
	FunctionResources map[string]map[string]ResourceMetrics
}

// InnerWorkerData is the worker data stored in ETCD
type InnerWorkerData struct {
	IP                        string           `json:"ip"`
	Port                      string           `json:"port"`
	NodeIP                    string           `json:"nodeIP"`
	P2pPort                   string           `json:"p2pPort"`
	NodeName                  string           `json:"nodeName"`
	NodeID                    string           `json:"nodeID"`
	WorkerAgentID             string           `json:"workerAgentID"`
	AllocatableCPU            int64            `json:"allocatableCPU"`
	AllocatableMemory         int64            `json:"allocatableMemory"`
	AllocatableCustomResource map[string]int64 `json:"allocatableCustomResource"`
}

// TerminateRequest sent from worker manager to worker to delete function instance
type TerminateRequest struct {
	RuntimeID   string `json:"runtime_id"`
	FuncName    string `json:"function_name"`
	FuncVersion string `json:"function_version"`
	TenantID    string `json:"tenant_id"`
	BusinessID  string `json:"business_id" valid:"optional"`
}

// UserAgency define AK/SK of user's agency
type UserAgency struct {
	AccessKey     string `json:"accessKey"`
	SecretKey     string `json:"secretKey"`
	Token         string `json:"token"`
	SecurityAk    string `json:"securityAk"`
	SecuritySk    string `json:"securitySk"`
	SecurityToken string `json:"securityToken"`
}

// CustomHealthCheck custom health check
type CustomHealthCheck struct {
	TimeoutSeconds   int `json:"timeoutSeconds" valid:",optional"`
	PeriodSeconds    int `json:"periodSeconds" valid:",optional"`
	FailureThreshold int `json:"failureThreshold" valid:",optional"`
}

// FuncCode include function code file and link info
type FuncCode struct {
	File string `json:"file" valid:",optional"`
	Link string `json:"link" valid:",optional"`
}

// StrategyConfig -
type StrategyConfig struct {
	Concurrency int `json:"concurrency" valid:",optional"`
}

// FuncSpec contains specifications of a function
type FuncSpec struct {
	ETCDType          string           `json:"-"`
	FunctionKey       string           `json:"-"`
	FuncMetaSignature string           `json:"-"`
	FuncMetaData      FuncMetaData     `json:"funcMetaData" valid:",optional"`
	S3MetaData        S3MetaData       `json:"s3MetaData" valid:",optional"`
	CodeMetaData      CodeMetaData     `json:"codeMetaData" valid:",optional"`
	EnvMetaData       EnvMetaData      `json:"envMetaData" valid:",optional"`
	StsMetaData       StsMetaData      `json:"stsMetaData" valid:",optional"`
	ResourceMetaData  ResourceMetaData `json:"resourceMetaData" valid:",optional"`
	InstanceMetaData  InstanceMetaData `json:"instanceMetaData" valid:",optional"`
	ExtendedMetaData  ExtendedMetaData `json:"extendedMetaData" valid:",optional"`
}

// FunctionMetaInfo define function meta info for FunctionGraph
type FunctionMetaInfo struct {
	FuncMetaData     FuncMetaData     `json:"funcMetaData" valid:",optional"`
	S3MetaData       S3MetaData       `json:"s3MetaData" valid:",optional"`
	CodeMetaData     CodeMetaData     `json:"codeMetaData" valid:",optional"`
	EnvMetaData      EnvMetaData      `json:"envMetaData" valid:",optional"`
	StsMetaData      StsMetaData      `json:"stsMetaData" valid:",optional"`
	ResourceMetaData ResourceMetaData `json:"resourceMetaData" valid:",optional"`
	InstanceMetaData InstanceMetaData `json:"instanceMetaData" valid:",optional"`
	ExtendedMetaData ExtendedMetaData `json:"extendedMetaData" valid:",optional"`
	RootfsSpecMeta   RootfsSpecMeta   `json:"rootfs" valid:",optional"`
}

// FuncMetaData define meta data of functions
type FuncMetaData struct {
	Layers              []*Layer          `json:"layers" valid:",optional"`
	Name                string            `json:"name"`
	FunctionDescription string            `json:"description" valid:"stringlength(1|1024)"`
	FunctionURN         string            `json:"functionUrn"`
	TenantID            string            `json:"tenantId"`
	AgentID             string            `json:"agentId" valid:",optional"`
	Tags                map[string]string `json:"tags" valid:",optional"`
	FunctionUpdateTime  string            `json:"functionUpdateTime" valid:",optional"`
	FunctionVersionURN  string            `json:"functionVersionUrn"`
	RevisionID          string            `json:"revisionId" valid:"stringlength(1|20),optional"`
	CodeSize            int               `json:"codeSize" valid:"int"`
	CodeSha512          string            `json:"codeSha512" valid:"stringlength(1|128),optional"`
	Handler             string            `json:"handler" valid:"stringlength(1|255)"`
	Runtime             string            `json:"runtime" valid:"stringlength(1|63)"`
	Timeout             int64             `json:"timeout" valid:"required"`
	Version             string            `json:"version" valid:"stringlength(1|32)"`
	DeadLetterConfig    string            `json:"deadLetterConfig" valid:"stringlength(1|255)"`
	BusinessID          string            `json:"businessId"  valid:"stringlength(1|32)"`
	FunctionType        string            `json:"functionType" valid:",optional"`
	FuncID              string            `json:"func_id" valid:",optional"`
	FuncName            string            `json:"func_name" valid:",optional"`
	DomainID            string            `json:"domain_id" valid:",optional"`
	ProjectName         string            `json:"project_name" valid:",optional"`
	Service             string            `json:"service" valid:",optional"`
	Dependencies        string            `json:"dependencies" valid:",optional"`
	EnableCloudDebug    string            `json:"enable_cloud_debug" valid:",optional"`
	IsStatefulFunction  bool              `json:"isStatefulFunction" valid:"optional"`
	IsBridgeFunction    bool              `json:"isBridgeFunction" valid:"optional"`
	IsStreamEnable      bool              `json:"isStreamEnable" valid:"optional"`
	Type                string            `json:"type" valid:"optional"`
	EnableAuthInHeader  bool              `json:"enable_auth_in_header" valid:"optional"`
	DNSDomainCfg        []DNSDomainInfo   `json:"dns_domain_cfg" valid:",optional"`
	VPCTriggerImage     string            `json:"vpcTriggerImage" valid:",optional"`
	StateConfig         StateConfig       `json:"stateConfig" valid:",optional"`
	BusinessType        string            `json:"businessType" valid:"optional"`
	AutoScaleConfig     AutoScaleConfig   `json:"autoScaleConfig" valid:"optional"`
}

type RootfsStorageInfo struct {
	Endpoint  string `json:"endpoint" valid:",optional"`
	Bucket    string `json:"bucket" valid:",optional"`
	Object    string `json:"object" valid:",optional"`
	AccessKey string `json:"accessKey" valid:",optional"`
	SecretKey string `json:"secretKey" valid:",optional"`
}

type RootfsSpecMeta struct {
	Runtime     string            `json:"runtime" valid:",optional"`
	Type        string            `json:"type" valid:",optional"`
	ImageURL    string            `json:"imageurl" valid:",optional"`
	Path        string            `json:"path" valid:",optional"`
	ReadOnly    bool              `json:"readonly" valid:",optional"`
	StorageInfo RootfsStorageInfo `json:"storageInfo" valid:",optional"`
	MountPoint  string            `json:"mountpoint" valid:",optional"`
}

// StateConfig ConsistentWithInstance- The lifecycle is consistent with that of the instance.
// Independent - The lifecycle is independent of instances.
type StateConfig struct {
	LifeCycle string `json:"lifeCycle"`
}

// S3MetaData define meta function info for OBS
type S3MetaData struct {
	AppID        string   `json:"appId" valid:"stringlength(1|128),optional"`
	BucketID     string   `json:"bucketId" valid:"stringlength(1|255),optional"`
	ObjectID     string   `json:"objectId" valid:"stringlength(1|255),optional"`
	BucketURL    string   `json:"bucketUrl" valid:"url,optional"`
	CodeType     string   `json:"code_type" valid:",optional"`
	CodeURL      string   `json:"code_url" valid:",optional"`
	CodeFileName string   `json:"code_filename" valid:",optional"`
	FuncCode     FuncCode `json:"func_code" valid:",optional"`
}

// LocalMetaData -
type LocalMetaData struct {
	StorageType string `json:"storage_type" valid:",optional"`
	CodePath    string `json:"code_path" valid:",optional"`
}

// CodeMetaData -
type CodeMetaData struct {
	Sha512 string `json:"sha512" valid:",optional"`
	LocalMetaData
	S3MetaData
}

// EnvMetaData -
type EnvMetaData struct {
	Environment       string `json:"environment"`
	EncryptedUserData string `json:"encrypted_user_data"`
	EnvKey            string `json:"envKey" valid:",optional"`
	CryptoAlgorithm   string `json:"cryptoAlgorithm" valid:",optional"`
}

// StsMetaData define sts info of functions
type StsMetaData struct {
	EnableSts        bool              `json:"enableSts"`
	ServiceName      string            `json:"serviceName,omitempty"`
	MicroService     string            `json:"microService,omitempty"`
	SensitiveConfigs map[string]string `json:"sensitiveConfigs,omitempty"`
	StsCertConfig    map[string]string `json:"stsCertConfig,omitempty"`
}

// ResourceMetaData include resource data such as cpu and memory
type ResourceMetaData struct {
	CPU                 int64  `json:"cpu"`
	Memory              int64  `json:"memory"`
	GpuMemory           int64  `json:"gpu_memory"`
	EnableDynamicMemory bool   `json:"enable_dynamic_memory" valid:",optional"`
	CustomResources     string `json:"customResources" valid:",optional"`
	EnableTmpExpansion  bool   `json:"enable_tmp_expansion" valid:",optional"`
	EphemeralStorage    int    `json:"ephemeral_storage" valid:"int,optional"`
	CustomResourcesSpec string `json:"CustomResourcesSpec" valid:",optional"`
}

// InstanceMetaData define instance meta data of FG functions
type InstanceMetaData struct {
	MaxInstance    int64  `json:"maxInstance" valid:",optional"`
	MinInstance    int64  `json:"minInstance" valid:",optional"`
	ConcurrentNum  int    `json:"concurrentNum" valid:",optional"`
	DiskLimit      int64  `json:"diskLimit"   valid:",optional"`
	InstanceType   string `json:"instanceType" valid:",optional"`
	SchedulePolicy string `json:"schedulePolicy" valid:",optional"`
	ScalePolicy    string `json:"scalePolicy" valid:",optional"`
	IdleMode       bool   `json:"idleMode" valid:",optional"`
	PoolLabel      string `json:"poolLabel"`
	PoolID         string `json:"poolId" valid:",optional"`
}

// ExtendedMetaData define external meta data of functions
type ExtendedMetaData struct {
	ImageName              string                 `json:"image_name" valid:",optional"`
	Role                   Role                   `json:"role" valid:",optional"`
	VpcConfig              *VpcConfig             `json:"func_vpc" valid:",optional"`
	EndpointTenantVpc      *VpcConfig             `json:"endpoint_tenant_vpc" valid:",optional"`
	FuncMountConfig        *FuncMountConfig       `json:"mount_config" valid:",optional"`
	StrategyConfig         StrategyConfig         `json:"strategy_config" valid:",optional"`
	ExtendConfig           string                 `json:"extend_config" valid:",optional"`
	Initializer            Initializer            `json:"initializer" valid:",optional"`
	Heartbeat              Heartbeat              `json:"heartbeat" valid:",optional"`
	EnterpriseProjectID    string                 `json:"enterprise_project_id" valid:",optional"`
	LogTankService         LogTankService         `json:"log_tank_service" valid:",optional"`
	TraceService           TraceService           `json:"tracing_config" valid:",optional"`
	CustomContainerConfig  CustomContainerConfig  `json:"custom_container_config" valid:",optional"`
	AsyncConfigLoaded      bool                   `json:"async_config_loaded" valid:",optional"`
	RestoreHook            RestoreHook            `json:"restore_hook,omitempty" valid:",optional"`
	NetworkController      NetworkController      `json:"network_controller" valid:",optional"`
	UserAgency             UserAgency             `json:"user_agency" valid:",optional"`
	CustomFilebeatConfig   CustomFilebeatConfig   `json:"custom_filebeat_config"`
	CustomHealthCheck      CustomHealthCheck      `json:"custom_health_check" valid:",optional"`
	DynamicConfig          DynamicConfigEvent     `json:"dynamic_config" valid:",optional"`
	CustomGracefulShutdown CustomGracefulShutdown `json:"runtime_graceful_shutdown"`
	PreStop                PreStop                `json:"pre_stop"`
	RaspConfig             RaspConfig             `json:"rasp_config"`
	ServeDeploySchema      ServeDeploySchema      `json:"serveDeploySchema" valid:"optional"`
	ImagePullConfig        ImagePullConfig        `json:"imagePullConfig,omitempty"`
	UserOtelConfig         UserOtelConfig         `json:"userOtelConfig,omitempty"`
	EnableAgentSession     bool                   `json:"enable_agent_session,omitempty"`
}

// UserOtelConfig -
type UserOtelConfig struct {
	Enable        bool              `json:"enable" valid:"optional"`
	InitContainer OtelInitContainer `json:"initContainer"`
	OtelEnv       map[string]string `json:"otelEnv"`
}

// OtelInitContainer -
type OtelInitContainer struct {
	Image           string          `json:"image"`
	Command         []string        `json:"command"`
	ShardDir        string          `json:"shardDir"`
	ResourceRequest ResourceRequire `json:"ResourceRequest"`
	ResourceLimit   ResourceRequire `json:"ResourceLimit"`
}

// ResourceRequire -
type ResourceRequire struct {
	Cpu    int `json:"cpu"`
	Memory int `json:"memory"`
}

// ImagePullConfig image pull config
type ImagePullConfig struct {
	Secrets []string `json:"secrets,omitempty"`
}

// CustomGracefulShutdown define the option of custom container's runtime graceful shutdown
type CustomGracefulShutdown struct {
	MaxShutdownTimeout int `json:"maxShutdownTimeout"`
}

// PreStop include pre_stop handler and timeout
type PreStop struct {
	Handler string `json:"pre_stop_handler" valid:",optional"`
	Timeout int    `json:"pre_stop_timeout" valid:",optional"`
}

// DynamicConfigEvent  dynamic config etcd event
type DynamicConfigEvent struct {
	Enabled       bool   `json:"enabled"` // use for signature
	UpdateTime    string `json:"update_time"`
	ConfigContent []KV   `json:"config_content"`
}

// KV config key and value
type KV struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Heartbeat define user custom heartbeat function config
type Heartbeat struct {
	// Handler define heartbeat function entry
	Handler string `json:"heartbeat_handler" valid:",optional"`
}

// CustomContainerConfig contains the metadata for custom container
type CustomContainerConfig struct {
	ControlPath string   `json:"control_path" valid:",optional"`
	Image       string   `json:"image" valid:",optional"`
	Command     []string `json:"command" valid:",optional"`
	Args        []string `json:"args" valid:",optional"`
	WorkingDir  string   `json:"working_dir" valid:",optional"`
	UID         int      `json:"uid" valid:",optional"`
	GID         int      `json:"gid" valid:",optional"`
}

// CustomFilebeatConfig custom filebeat config
type CustomFilebeatConfig struct {
	SidecarConfigInfo *SidecarConfigInfo `json:"sidecarConfigInfo"`
	CPU               int64              `json:"cpu"`
	Memory            int64              `json:"memory"`
	Version           string             `json:"version"`
	ImageAddress      string             `json:"imageAddress"`
}

// RaspConfig rasp config key and value
type RaspConfig struct {
	InitImage      string `json:"init-image"`
	RaspImage      string `json:"rasp-image"`
	RaspServerIP   string `json:"rasp-server-ip"`
	RaspServerPort string `json:"rasp-server-port"`
	Envs           []KV   `json:"envs"`
}

// SidecarConfigInfo sidecat config info
type SidecarConfigInfo struct {
	ConfigFiles     []CustomLogConfigFile `json:"configFiles"`
	LiveNessShell   string                `json:"livenessShell"`
	ReadNessShell   string                `json:"readnessShell"`
	PreStopCommands string                `json:"preStopCommands"`
}

// CustomLogConfigFile custom log config file
type CustomLogConfigFile struct {
	Path   string `json:"path"`
	Data   string `json:"data"`
	Secret bool   `json:"secret"`
}

// RestoreHook include restorehook handler and timeout
type RestoreHook struct {
	Handler string `json:"restore_hook_handler,omitempty" valid:",optional"`
	Timeout int64  `json:"restore_hook_timeout,omitempty" valid:",optional"`
}

// NetworkController contains some special network settings
type NetworkController struct {
	DisablePublicNetwork bool      `json:"disable_public_network" valid:",optional"`
	TriggerAccessVpcs    []VpcInfo `json:"trigger_access_vpcs" valid:",optional"`
}

// PATServiceRequest -
type PATServiceRequest struct {
	ID             string   `json:"id,omitempty"`
	Namespace      string   `json:"namespace"`
	DomainID       string   `json:"domainID,omitempty"`
	ProjectID      string   `json:"projectID,omitempty"`
	VpcID          string   `json:"vpcID,omitempty"`
	SubnetID       string   `json:"subnetID,omitempty"`
	TenantCidr     string   `json:"tenantCidr,omitempty"`
	HostVMCidr     string   `json:"hostVMCidr,omitempty"`
	Gateway        string   `json:"gateway,omitempty"`
	SecurityGroups []string `json:"securityGroups,omitempty"`
	Xrole          string   `json:"xrole,omitempty"`
}

// PatResponseInfo create pat response
type PatResponseInfo struct {
	PatPods []NATConfigure `json:"patPods"`
}

// NATConfigure pat service config
type NATConfigure struct {
	Namespace      string `json:"namespace"`
	PatContainerIP string `json:"patContainerIP"`
	PatVMIP        string `json:"patVmIP"`
	PatPortIP      string `json:"patPortIP"`
	PatMacAddr     string `json:"patMacAddr"`
	PatGateway     string `json:"patGateway"`
	PatPodName     string `json:"patPodName"`
	TenantCidr     string `json:"tenantCidr"`
	SubMetaDigest  string `json:"subMetaDigest"`
}

// VpcInfo contains the information of VPC access restriction
type VpcInfo struct {
	VpcName string `json:"vpc_name,omitempty"`
	VpcID   string `json:"vpc_id,omitempty"`
}

// VpcConfig include info of function vpc
type VpcConfig struct {
	ID             string   `json:"id,omitempty"`
	DomainID       string   `json:"domain_id,omitempty"`
	Namespace      string   `json:"namespace,omitempty"`
	VpcName        string   `json:"vpc_name,omitempty"`
	VpcID          string   `json:"vpc_id,omitempty"`
	SubnetName     string   `json:"subnet_name,omitempty"`
	SubnetID       string   `json:"subnet_id,omitempty"`
	TenantCidr     string   `json:"tenant_cidr,omitempty"`
	HostVMCidr     string   `json:"host_vm_cidr,omitempty"`
	Gateway        string   `json:"gateway,omitempty"`
	Xrole          string   `json:"xrole,omitempty"`
	SecurityGroups []string `json:"security_groups,omitempty"`
}

// Layer define layer info
type Layer struct {
	BucketURL      string `json:"bucketUrl" valid:"url,optional"`
	ObjectID       string `json:"objectId" valid:"stringlength(1|255),optional"`
	BucketID       string `json:"bucketId" valid:"stringlength(1|255),optional"`
	AppID          string `json:"appId" valid:"stringlength(1|128),optional"`
	ETag           string `json:"etag" valid:"optional"`
	Link           string `json:"link" valid:"optional"`
	Name           string `json:"name" valid:",optional"`
	Sha256         string `json:"sha256" valid:"optional"`
	DependencyType string `json:"dependencyType" valid:",optional"`
}

// DNSDomainInfo dns domain info
type DNSDomainInfo struct {
	ID         string `json:"id"`
	DomainName string `json:"domain_name"`
	Type       string `json:"type" valid:",optional"`
	ZoneType   string `json:"zone_type" valid:",optional"`
}

// DataSystemConfig data system client config
type DataSystemConfig struct {
	TimeoutMs int      `json:"timeoutMs" validate:"required"`
	Clusters  []string `json:"clusters"`
}

// XiangYunFourConfig -
type XiangYunFourConfig struct {
	Site          string `json:"site"`
	TenantID      string `json:"tenantID"`
	ApplicationID string `json:"applicationID"`
	ServiceID     string `json:"serviceID"`
}

// MemoryControlConfig Memory use control config
type MemoryControlConfig struct {
	LowerMemoryPercent     float64 `json:"lowerMemoryPercent" valid:",optional"`
	HighMemoryPercent      float64 `json:"highMemoryPercent" valid:",optional"`
	StatefulHighMemPercent float64 `json:"statefulHighMemoryPercent" valid:",optional"`
	BodyThreshold          uint64  `json:"bodyThreshold" valid:",optional"`
	MemDetectIntervalMs    int     `json:"memDetectIntervalMs" valid:",optional"`
}

// InstanceStatus Instance status, controlled by the kernel
type InstanceStatus struct {
	Code      int32  `json:"code" validate:"required"`
	Msg       string `json:"msg" validate:"required"`
	Type      int32  `json:"type" validate:"optional"`
	ExitCode  int32  `json:"exitCode" validate:"optional"`
	ErrorCode int32  `json:"errCode" validate:"optional"`
}

// PodResourceInfo describe actual resource info of pod
type PodResourceInfo struct {
	Worker  ResourceConfig `json:"worker,omitempty"`
	Runtime ResourceConfig `json:"runtime,omitempty"`
}

// ResourceConfig sub-struct of FuncInstanceInfo
type ResourceConfig struct {
	CPULimit      int64 `json:"cpuLimit" valid:",optional"` // unit: milli-cores(m)
	CPURequest    int64 `json:"cpuRequest" valid:",optional"`
	MemoryLimit   int64 `json:"memoryLimit" valid:",optional"` // unit: byte
	MemoryRequest int64 `json:"memoryRequest" valid:",optional"`
}

// Extensions -
type Extensions struct {
	Source            string `json:"source"`
	CreateTimestamp   string `json:"createTimestamp"`
	UpdateTimestamp   string `json:"updateTimestamp"`
	PID               string `json:"pid"`
	PodName           string `json:"podName"`
	PodNamespace      string `json:"podNamespace"`
	PodDeploymentName string `json:"podDeploymentName"`
}

// InstanceSpecification contains specification of a instance in etcd
type InstanceSpecification struct {
	InstanceID      string            `json:"instanceID" validate:"required"`
	DataSystemHost  string            `json:"dataSystemHost" validate:"required"`
	RequestID       string            `json:"requestID" valid:",optional"`
	RuntimeID       string            `json:"runtimeID" valid:",optional"`
	RuntimeAddress  string            `json:"runtimeAddress" valid:",optional"`
	FunctionAgentID string            `json:"functionAgentID" valid:",optional"`
	FunctionProxyID string            `json:"functionProxyID" valid:",optional"`
	Function        string            `json:"function"`
	RestartPolicy   string            `json:"restartPolicy" valid:",optional"`
	Resources       Resources         `json:"resources"`
	ActualUse       Resources         `json:"actualUse" valid:",optional"`
	ScheduleOption  ScheduleOption    `json:"scheduleOption"`
	CreateOptions   map[string]string `json:"createOptions"`
	Labels          []string          `json:"labels"`
	StartTime       string            `json:"startTime"`
	InstanceStatus  InstanceStatus    `json:"instanceStatus"`
	JobID           string            `json:"jobID"`
	SchedulerChain  []string          `json:"schedulerChain" valid:",optional"`
	ParentID        string            `json:"parentID"`
	DeployTimes     int32             `json:"deployTimes"`
	Extensions      Extensions        `json:"extensions" valid:",optional"`
}

// InstanceSpecificationFG contains specification of instance in etcd for functionGraph
type InstanceSpecificationFG struct {
	OwnerIP      string          `json:"ownerIP"`
	CreationTime int             `json:"creationTime"`
	Applier      string          `json:"applier"`
	NodeIP       string          `json:"nodeIP"`
	NodePort     string          `json:"nodePort"`
	InstanceIP   string          `json:"ip"`
	InstancePort string          `json:"port"`
	CPU          int             `json:"cpu"`
	Memory       int             `json:"memory"`
	BusinessType string          `json:"businessType"`
	Resource     PodResourceInfo `json:"resource,omitempty"`
}

// Resources -
type Resources struct {
	Resources map[string]Resource `json:"resources"`
}

// Resource -
type Resource struct {
	Name    string      `json:"name"`
	Type    ValueType   `json:"type"`
	Scalar  ValueScalar `json:"scalar"`
	Ranges  ValueRanges `json:"ranges"`
	Set     ValueSet    `json:"set"`
	Runtime string      `json:"runtime"`
	Driver  string      `json:"driver"`
	Disk    DiskInfo    `json:"disk"`
}

// ValueType -
type ValueType int32

// ValueScalar -
type ValueScalar struct {
	Value float64 `json:"value"`
	Limit float64 `json:"limit"`
}

// ValueRanges -
type ValueRanges struct {
	Range []ValueRange `protobuf:"bytes,1,rep,name=range,proto3" json:"range,omitempty"`
}

// ValueSet -
type ValueSet struct {
	Items string `json:"items"`
}

// ValueRange -
type ValueRange struct {
	Begin uint64 `json:"begin"`
	End   uint64 `json:"end"`
}

// DiskInfo -
type DiskInfo struct {
	Volume    Volume `json:"volume"`
	Type      string `json:"type"`
	DevPath   string `json:"devPath"`
	MountPath string `json:"mountPath"`
}

// Volume -
type Volume struct {
	Mode          int32  `json:"mode"`
	SourceType    int32  `json:"sourceType"`
	HostPaths     string `json:"hostPaths"`
	ContainerPath string `json:"containerPath"`
	ConfigMapPath string `json:"configMapPath"`
	EmptyDir      string `json:"emptyDir"`
	ElaraPath     string `json:"elaraPath"`
}

// ScheduleOption -
type ScheduleOption struct {
	SchedPolicyName string   `json:"schedPolicyName"`
	Priority        int32    `json:"priority"`
	Affinity        Affinity `json:"affinity"`
}

// Affinity -
type Affinity struct {
	NodeAffinity         NodeAffinity     `json:"nodeAffinity"`
	InstanceAffinity     InstanceAffinity `json:"instanceAffinity"`
	InstanceAntiAffinity InstanceAffinity `json:"instanceAntiAffinity"`
}

// NodeAffinity -
type NodeAffinity struct {
	Affinity map[string]string `json:"affinity"`
}

// InstanceAffinity -
type InstanceAffinity struct {
	Affinity map[string]string `json:"affinity"`
}

// InstanceInfo the instance info which can be parsed from the etcd path, instanceName is used to hold a place in the
// hash ring while instanceID is used to invoke this instance
type InstanceInfo struct {
	TenantID     string
	FunctionName string
	Version      string
	InstanceName string `json:"instanceName"`
	InstanceID   string `json:"instanceId"`
	Exclusivity  string
	Address      string
}

// RolloutResponse -
type RolloutResponse struct {
	AllocRecord  map[string][]string `json:"allocRecord"`
	RegisterKey  string              `json:"registerKey"`
	ErrorCode    int                 `json:"errorCode"`
	ErrorMessage string              `json:"errorMessage"`
}

// NuwaRuntimeInfo contains ers workload info for function
type NuwaRuntimeInfo struct {
	WisecloudRuntimeId     string `json:"wisecloudRuntimeId"`
	WisecloudSite          string `json:"wisecloudSite"`
	WisecloudTenantId      string `json:"wisecloudTenantId"`
	WisecloudApplicationId string `json:"wisecloudApplicationId"`
	WisecloudServiceId     string `json:"wisecloudServiceId"`
	WisecloudEnvironmentId string `json:"wisecloudEnvironmentId"`
	EnvLabel               string `json:"envLabel"`
}

// CallHandlerResponse is the response returned by faas manager's CallHandler
type CallHandlerResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ResponseWriter -
type ResponseWriter interface {
	SSEWrite([]byte) (int, error)
	ClientDisconnectChan() <-chan struct{}
}

// AutoScaleConfig -
type AutoScaleConfig struct {
	SLAQuota      int `json:"slaQuota" valid:"optional" default:"-1"`
	ScaleDownTime int `json:"scaleDownTime" valid:"optional" default:"-1"`
	BurstScaleNum int `json:"burstScaleNum" valid:"optional" default:"-1"`
}
