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

#include "yr/yr.h"

int AfterSleepSec(int x)
{
    std::this_thread::sleep_for(std::chrono::seconds(x));
    return x + 1;
}
YR_INVOKE(AfterSleepSec)

int main(int argc, char **argv)
{
    // check args num
    if (argc != 3) {
        std::cerr << "Usage: " << argv[0] << " <string> <int>" << std::endl;
        return 1;
    }
    std::string testMode = argv[1];  // CREATE|INVOKE
    int mockBadNetworkDurationSec = std::stoi(argv[2]);

    YR::Config config;
    config.mode = YR::Config::Mode::CLUSTER_MODE;
    auto info = YR::Init(config);
    std::cout << "job id: " << info.jobId << std::endl;

    // RPC retry test

    pid_t pid = getpid();
    std::cout << "Process ID: " << pid << std::endl;
    std::system("netstat -nap |grep -E \"main|function_pro\" |grep tcp6");
    std::cout << "Input the function_proxy SRC port to runtime:" << std::endl;

    std::string port;
    std::cin >> port;
    std::cout << "RUN. will mock bad network at port " << port << " for " << mockBadNetworkDurationSec <<" s."<< std::endl;

    if(testMode == "INVOKE") {
        std::cout << "Is INVOKE test. immediately invoke to Create a Instance."<<std::endl;
        auto obj1 = YR::Function(AfterSleepSec).Invoke(1);
        int val1 = *YR::Get(obj1);
        std::cout << "Immediately invoke received val is: " << val1 << std::endl;
    } else {
        std::cout << "Is CREATE test."<<std::endl;
    }

    std::thread enablePortThread([port, mockBadNetworkDurationSec]() {
        if (mockBadNetworkDurationSec > 0) {
            std::system(
                ("bash " + std::string(std::getenv("BASE_DIR")) + "/mock_network_problem.sh " + port + " " + std::to_string(mockBadNetworkDurationSec)).c_str());
        }
    });

    // sleep 100ms
    std::this_thread::sleep_for(std::chrono::milliseconds(500));

    auto obj2 = YR::Function(AfterSleepSec).Invoke(1);
    try {
        int val = *YR::Get(obj2);  // ==2
        std::cout << "received! val is: " << val << std::endl;
    } catch (YR::Exception e) {
        std::cout << e.Msg() << std::endl;
    }

    enablePortThread.join();
    std::cout << "Finished, you should check log file." << std::endl;

    YR::Finalize();
    return 0;
}
