/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

#pragma once

#include <list>
#include <memory>
#include <sstream>

#include "src/dto/constant.h"
#include "src/libruntime/fsclient/protobuf/common.grpc.pb.h"
#include "src/libruntime/fsclient/protobuf/core_service.grpc.pb.h"

namespace YR {
namespace Libruntime {
using PBAffinity = ::common::Affinity;
using LabelExpression = ::common::LabelExpression;
using SubCondition = ::common::SubCondition;
using Condition = ::common::Condition;
using Selector = ::common::Selector;
using PBInstanceAffinity = ::common::InstanceAffinity;
using PBAffinityScope = ::common::AffinityScope;

class LabelOperator {
public:
    LabelOperator(const std::string &type) : operatorType(type) {}

    LabelOperator() = default;
    ~LabelOperator() = default;

    std::string GetOperatorType() const
    {
        return this->operatorType;
    }

    std::string GetKey() const
    {
        return this->key;
    }

    std::list<std::string> GetValues() const
    {
        return this->values;
    }

    std::size_t GetLabelOperatorHash() const
    {
        std::size_t h1 = std::hash<std::string>()(GetOperatorType());
        std::size_t h2 = std::hash<std::string>()(GetKey());
        std::size_t res = h1 ^ h2;
        auto values = GetValues();
        for (auto &value : values) {
            std::size_t h3 = std::hash<std::string>()(value);
            res = res ^ h3;
        }
        return res;
    }

    void SetKey(std::string key)
    {
        this->key = key;
    }

    void SetValues(std::list<std::string> values)
    {
        this->values = values;
    }

    virtual LabelExpression GetLabelMatchExpression()
    {
        return LabelExpression();
    }

    std::string GetString()
    {
        std::stringstream ss;
        ss << operatorType << " " << key << " ";
        for (const auto &v : this->values) {
            ss << v << " ";
        }
        return ss.str();
    }

private:
    std::string operatorType;
    std::string key;
    std::list<std::string> values;
};

class LabelInOperator : public LabelOperator {
public:
    LabelInOperator() : LabelOperator(LABEL_IN) {}

    LabelExpression GetLabelMatchExpression()
    {
        auto expression = LabelExpression();
        expression.set_key(GetKey());
        auto values = GetValues();
        for (auto &value : values) {
            expression.mutable_op()->mutable_in()->add_values(value);
        }
        return expression;
    }
};

class LabelNotInOperator : public LabelOperator {
public:
    LabelNotInOperator() : LabelOperator(LABEL_NOT_IN) {}

    LabelExpression GetLabelMatchExpression()
    {
        auto expression = LabelExpression();
        expression.set_key(GetKey());
        auto values = GetValues();
        for (auto &value : values) {
            expression.mutable_op()->mutable_notin()->add_values(value);
        }
        return expression;
    }
};

class LabelExistsOperator : public LabelOperator {
public:
    LabelExistsOperator() : LabelOperator(LABEL_EXISTS) {}

    LabelExpression GetLabelMatchExpression()
    {
        auto expression = LabelExpression();
        expression.set_key(GetKey());
        expression.mutable_op()->mutable_exists();
        return expression;
    }
};

class LabelDoesNotExistOperator : public LabelOperator {
public:
    LabelDoesNotExistOperator() : LabelOperator(LABEL_DOES_NOT_EXIST) {}

    LabelExpression GetLabelMatchExpression()
    {
        auto expression = LabelExpression();
        expression.set_key(GetKey());
        expression.mutable_op()->mutable_notexist();
        return expression;
    }
};

class Affinity {
public:
    Affinity(const std::string &kind, const std::string &type) : affinityKind(kind), affinityType(type) {}

    Affinity() = default;
    ~Affinity() = default;

    std::string GetAffinityKind() const
    {
        return this->affinityKind;
    }

    std::string GetAffinityType() const
    {
        return this->affinityType;
    }

    std::list<std::shared_ptr<LabelOperator>> GetLabelOperators() const
    {
        return this->labelOperators;
    }

    void SetLabelOperators(std::list<std::shared_ptr<LabelOperator>> operators)
    {
        this->labelOperators = std::move(operators);
    }

    void SetPreferredPriority(bool preferredPriority)
    {
        this->preferredPriority = preferredPriority;
    }

    void SetRequiredPriority(bool requiredPriority)
    {
        this->requiredPriority = requiredPriority;
    }

    void SetPreferredAntiOtherLabels(bool preferredAntiOtherLabels)
    {
        this->preferredAntiOtherLabels = preferredAntiOtherLabels;
    }

    bool GetPreferredAntiOtherLabels()
    {
        return this->preferredAntiOtherLabels;
    }

    std::string GetAffinityScope() const
    {
        return this->affinityScope;
    }

    void SetAffinityScope(const std::string &affinityScope)
    {
        this->affinityScope = affinityScope;
    }

    void UpdateAffinityScope(PBInstanceAffinity *pbInstanceAffinity)
    {
        if (affinityScope == AFFINITYSCOPE_NODE) {
            pbInstanceAffinity->set_scope(common::AffinityScope::NODE);
        } else if (affinityScope == AFFINITYSCOPE_POD) {
            pbInstanceAffinity->set_scope(common::AffinityScope::POD);
        }
    }

    virtual void UpdatePbAffinity(PBAffinity *pbAffinity)
    {
        auto *condition = pbAffinity->mutable_resource()->mutable_preferredaffinity()->mutable_condition();
        UpdateCondition(condition);
    }

    std::size_t GetAffinityHash()
    {
        std::size_t h1 = std::hash<std::string>()(GetAffinityKind());
        std::size_t h2 = std::hash<std::string>()(GetAffinityType());
        std::size_t res = h1 ^ h2;
        auto labelOperators = GetLabelOperators();
        for (auto it = labelOperators.begin(); it != labelOperators.end(); it++) {
            std::size_t h3 = (*it)->GetLabelOperatorHash();
            res = res ^ h3;
        }
        return res;
    }

    std::vector<LabelExpression> GetLabels()
    {
        std::vector<LabelExpression> res;
        for (auto labelOperator : labelOperators) {
            res.push_back(labelOperator->GetLabelMatchExpression());
        }
        return res;
    }

    void UpdateCondition(Condition *condition)
    {
        condition->set_orderpriority(requiredPriority || preferredPriority);
        auto labels = GetLabels();
        auto subConditions = condition->add_subconditions();
        for (auto express : labels) {
            *subConditions->add_expressions() = express;
        }
    }

    std::string GetString()
    {
        std::stringstream ss;
        ss << affinityKind << " " << affinityType << " " << preferredPriority << " " << preferredAntiOtherLabels << " ";
        for (const auto &it : this->labelOperators) {
            ss << it->GetString() << " ";
        }
        return ss.str();
    }

protected:
    std::string affinityKind;
    std::string affinityType;
    bool preferredPriority = true;
    bool requiredPriority = false;
    bool preferredAntiOtherLabels = false;
    std::list<std::shared_ptr<LabelOperator>> labelOperators;
    std::string affinityScope;
};

class ResourcePreferredAffinity : public Affinity {
public:
    ResourcePreferredAffinity() : Affinity(RESOURCE, PREFERRED) {}

    void UpdatePbAffinity(PBAffinity *pbAffinity)
    {
        if (requiredPriority || (preferredAntiOtherLabels && preferredPriority)) {
            auto *condition = pbAffinity->mutable_resource()->mutable_requiredaffinity()->mutable_condition();
            UpdateCondition(condition);
        } else {
            auto *condition = pbAffinity->mutable_resource()->mutable_preferredaffinity()->mutable_condition();
            UpdateCondition(condition);
        }
    }
};

class InstancePreferredAffinity : public Affinity {
public:
    InstancePreferredAffinity() : Affinity(INSTANCE, PREFERRED) {}

    void UpdatePbAffinity(PBAffinity *pbAffinity)
    {
        auto *condition = pbAffinity->mutable_instance()->mutable_preferredaffinity()->mutable_condition();
        UpdateCondition(condition);
        auto *pbInstanceAffinity = pbAffinity->mutable_instance();
        UpdateAffinityScope(pbInstanceAffinity);
    }
};

class ResourcePreferredAntiAffinity : public Affinity {
public:
    ResourcePreferredAntiAffinity() : Affinity(RESOURCE, PREFERRED_ANTI) {}

    void UpdatePbAffinity(PBAffinity *pbAffinity)
    {
        if (requiredPriority || (preferredPriority && preferredAntiOtherLabels)) {
            auto *condition = pbAffinity->mutable_resource()->mutable_requiredantiaffinity()->mutable_condition();
            UpdateCondition(condition);
        } else {
            auto *condition = pbAffinity->mutable_resource()->mutable_preferredantiaffinity()->mutable_condition();
            UpdateCondition(condition);
        }
    }
};

class InstancePreferredAntiAffinity : public Affinity {
public:
    InstancePreferredAntiAffinity() : Affinity(INSTANCE, PREFERRED_ANTI) {}

    void UpdatePbAffinity(PBAffinity *pbAffinity)
    {
        auto *condition = pbAffinity->mutable_instance()->mutable_preferredantiaffinity()->mutable_condition();
        UpdateCondition(condition);
        auto *pbInstanceAffinity = pbAffinity->mutable_instance();
        UpdateAffinityScope(pbInstanceAffinity);
    }
};

class ResourceRequiredAffinity : public Affinity {
public:
    ResourceRequiredAffinity() : Affinity(RESOURCE, REQUIRED) {}

    void UpdatePbAffinity(PBAffinity *pbAffinity)
    {
        auto *condition = pbAffinity->mutable_resource()->mutable_requiredaffinity()->mutable_condition();
        UpdateCondition(condition);
    }
};

class InstanceRequiredAffinity : public Affinity {
public:
    InstanceRequiredAffinity() : Affinity(INSTANCE, REQUIRED) {}

    void UpdatePbAffinity(PBAffinity *pbAffinity)
    {
        auto *condition = pbAffinity->mutable_instance()->mutable_requiredaffinity()->mutable_condition();
        UpdateCondition(condition);
        auto *pbInstanceAffinity = pbAffinity->mutable_instance();
        UpdateAffinityScope(pbInstanceAffinity);
    }
};

class ResourceRequiredAntiAffinity : public Affinity {
public:
    ResourceRequiredAntiAffinity() : Affinity(RESOURCE, REQUIRED_ANTI) {}

    void UpdatePbAffinity(PBAffinity *pbAffinity)
    {
        auto *condition = pbAffinity->mutable_resource()->mutable_requiredantiaffinity()->mutable_condition();
        UpdateCondition(condition);
    }
};

class InstanceRequiredAntiAffinity : public Affinity {
public:
    InstanceRequiredAntiAffinity() : Affinity(INSTANCE, REQUIRED_ANTI) {}

    void UpdatePbAffinity(PBAffinity *pbAffinity)
    {
        auto *condition = pbAffinity->mutable_instance()->mutable_requiredantiaffinity()->mutable_condition();
        UpdateCondition(condition);
        auto *pbInstanceAffinity = pbAffinity->mutable_instance();
        UpdateAffinityScope(pbInstanceAffinity);
    }
};
}  // namespace Libruntime
}  // namespace YR