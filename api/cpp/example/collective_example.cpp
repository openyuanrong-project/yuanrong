/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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
//! [cpp instance function]
class CollectiveActor {
public:
    int count;

    CollectiveActor() = default;

    ~CollectiveActor() = default;

    static CollectiveActor *FactoryCreate()
    {
        return new CollectiveActor();
    }

    int Compute(std::vector<int> in, std::string &groupName, uint8_t op)
    {
        int *output = new int[in.size()];

        // allreduce
        YR::Collective::AllReduce(in.data(), output, in.size(), YR::DataType::INT, YR::ReduceOp(op), groupName);
        YR::Collective::Barrier(groupName);

        // send and recv
        if (YR::Collective::GetRank(groupName) == 0) {
            YR::Collective::Send(in.data(), in.size(), YR::DataType::INT, 1, 1234, groupName);
        } else if (YR::Collective::GetRank(groupName) == 1) {
            output = new int[in.size()];
            YR::Collective::Recv(output, in.size(), YR::DataType::INT, 0, 1234, groupName);
        }
        YR::Collective::Barrier(groupName);

        int result = 0;
        for (int i = 0; i < in.size(); ++i) {
            result += *output++;
        }
        YR::Collective::DestroyCollectiveGroup(groupName);
        return result;
    }
};

YR_INVOKE(CollectiveActor::FactoryCreate, &CollectiveActor::Compute)

int main(void)
{
    YR::Config conf;
    YR::Init(conf);

    std::vector<YR::NamedInstance<CollectiveActor>> instances;
    std::vector<std::string> instanceIDs;
    for (int i = 0; i < 4; ++i) {
        auto ins = YR::Instance(CollectiveActor::FactoryCreate).Invoke();
        instances.push_back(ins);
        instanceIDs.push_back(ins.GetInstanceId());
    }

    std::string groupName = "test-group";
    YR::Collective::CreateCollectiveGroup(instanceIDs, 4, {0, 1, 2, 3}, groupName);

    std::vector<int> input = {1, 2, 3, 4};
    std::vector<YR::ObjectRef<int>> res;
    for (int i = 0; i < 4; ++i) {
        res.push_back(instances[i]
                          .Function(&CollectiveActor::Compute)
                          .Invoke(input, groupName, static_cast<uint8_t>(YR::ReduceOp::SUM)));
    }
    auto res0 = *YR::Get(res[0]);  // allreduce result
    auto res1 = *YR::Get(res[1]);  // send recv result(input)

    std::cout << "allreduce result: " << res0 << ", recv result: " << res1 << std::endl;
    return 0;
}
//! [cpp instance function]