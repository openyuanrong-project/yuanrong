/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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

int PlusTen(int x)
{
    return x + 10;  // add 10
}

YR_INVOKE(PlusTen)

int main()
{
    YR::Config conf;
    YR::Init(conf);

    // support 8 affinities：ResourcePreferredAffinity、InstancePreferredAffinity、ResourcePreferredAntiAffinity、
    // InstancePreferredAntiAffinity、ResourceRequiredAffinity、InstanceRequiredAffinity、ResourceRequiredAntiAffinity、
    // InstanceRequiredAntiAffinity
    // support 4 labelOperators: LabelInOperator、LabelNotInOperator、LabelExistsOperator、LabelDoesNotExistOperator
    YR::InvokeOptions opts;
    auto op1 = YR::LabelExistsOperator("key1");
    auto aff1 = YR::ResourcePreferredAffinity(op1);
    auto aff2 = YR::ResourceRequiredAffinity(YR::LabelInOperator("key2", {"value1, value2"}));
    auto aff3 = YR::ResourcePreferredAffinity({YR::LabelDoesNotExistOperator("key3")});
    auto aff4 =
        YR::ResourcePreferredAffinity({YR::LabelExistsOperator("key4"), YR::LabelNotInOperator("key5", {"value3"})});
    // only valid while preferred affinities, when false the preferred list order is invalid, default true.
    opts.preferredPriority = false;
    opts.AddAffinity(aff1).AddAffinity(aff2).AddAffinity({aff3, aff4});
    auto r2 = YR::Function(PlusTen).Options(opts).Invoke(2);
    auto integer = *YR::Get(r2);
    printf("res :%d\n", integer);

    YR::Finalize();
    return 0;
}