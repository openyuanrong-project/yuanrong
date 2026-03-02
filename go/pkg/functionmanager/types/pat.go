/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
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

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Pat define k8s obj
type Pat struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              PatSpec   `json:"spec"`
	Status            PatStatus `json:"status"`
}

// PatSpec define k8s spec
type PatSpec struct {
	KeepAliveTimeout int64  `json:"keep_alive_timeout"`
	DelegateRole     string `json:"delegate_role"`
	RequireCount     int64  `json:"require_count"`
	EnvironmentId    string `json:"environment_id"`
	Vpc              VPC    `json:"vpc"`
}

// VPC define k8s spec vpc
type VPC struct {
	DomainId       string   `json:"domain_id"`
	ProjectId      string   `json:"project_id"`
	VpcId          string   `json:"vpc_id"`
	SubnetId       string   `json:"subnet_id"`
	TenantCidr     string   `json:"tenant_cidr"`
	HostVmCidr     string   `json:"host_vm_cidr"`
	Gateway        string   `json:"gateway"`
	SecurityGroups []string `json:"security_groups"`
	Ipv6Enable     bool     `json:"ipv6_enable"`
}

// PatStatus define k8s status
type PatStatus struct {
	PatPods []PatPodStatus `json:"pat_pods"`
}

// PatPodStatus define k8s status pod info
type PatPodStatus struct {
	PatContainerIp    string    `json:"pat_container_ip"`
	PatVmIp           string    `json:"pat_vm_ip"`
	PatPortIp         string    `json:"pat_port_ip"`
	PatMacAddr        string    `json:"pat_mac_addr"`
	PatGateway        string    `json:"pat_gateway"`
	PatPodName        string    `json:"pat_pod_name"`
	TenantCidr        string    `json:"tenant_cidr"`
	NatExcludeSubnets []string  `json:"nat_exclude_subnets"`
	SubMetaDigest     string    `json:"sub_meta_digest"`
	Status            string    `json:"status"`
	LastUpdateTime    time.Time `json:"last_update_time"`
}
