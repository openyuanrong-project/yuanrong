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
# usage bash mock_network_problem.sh [tcp_port] [recover_time_second].
# This script will cause the data packets sent to tcp_port to be lost for a duration of recover_time_second.

set -e
tcp_port=$1
recover_time_second=$2

# check whether tcp_port and recover_time is valid. if invalid, print error and quit.
if [ -z "$tcp_port" ] || [ -z "$recover_time_second" ]
then
   echo "Usage: bash mock_network_problem.sh tcp_port recover_time_second."
   exit 1
fi

echo "Start mock bad network in port ${tcp_port} for ${recover_time_second}s."
sudo modprobe ifb
sudo ip link set dev ifb0 up
# redirect packet to ifb0
sudo tc qdisc add dev lo ingress
sudo tc filter add dev lo parent ffff: protocol ip u32 match ip dport ${tcp_port} 0xffff flowid 1:1 action mirred egress redirect dev ifb0
# add delay
# sudo tc qdisc add dev ifb0 root netem delay 190s
# add packet loss
sudo tc qdisc add dev ifb0 root netem loss 100%

sleep ${recover_time_second}

# remove rules
sudo tc qdisc del dev lo ingress
sudo tc qdisc del dev ifb0 root

echo "Mock bad network finished."