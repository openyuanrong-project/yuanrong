/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
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

#include <iostream>

#include "yr/yr.h"

int main()
{
    YR::Config conf;
    YR::Init(conf);
    //! [create producer]
    std::string streamName = "streamName";
    // create stream producer
    YR::ProducerConf producerConf{};
    std::shared_ptr<YR::Producer> producer = YR::CreateProducer(streamName, producerConf);
    //! [create producer]
    // create stream consumer
    //! [create consumer]
    YR::SubscriptionConfig config("subName", YR::SubscriptionType::STREAM);
    std::shared_ptr<YR::Consumer> consumer = YR::Subscribe(streamName, config);
    //! [create consumer]
    //! [producer send]
    // producer send data
    std::string str = "hello";
    YR::Element element((uint8_t *)(str.c_str()), str.size());
    producer->Send(element);
    //! [producer send]
    //! [consumer recv]
    // consumer receive data
    std::vector<YR::Element> elements;
    consumer->Receive(1, 6000, elements);  // timeout 6s
    consumer->Ack(elements[0].id);
    std::string actualData0(reinterpret_cast<char *>(elements[0].ptr), elements[0].size);
    std::cout << "receive: " << actualData0 << std::endl;
    //! [consumer recv]
    //! [close producer]
    producer->Close();
    //! [close producer]
    //! [close consumer]
    consumer->Close();
    //! [close consumer]
    //! [delete stream]
    // delete stream
    YR::DeleteStream(streamName);
    //! [delete stream]
    return 0;
}