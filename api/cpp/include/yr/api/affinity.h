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

#include "yr/api/constant.h"

namespace YR {
/*! @class LabelOperator affinity.h "include/yr/api/affinity.h"
 *  @brief This is the base class for tag operators, which defines the common interface for tag operators.
 */
class LabelOperator {
public:
    /*!
    @brief Does not support default constructor.
    */
    LabelOperator() = delete;

    /*!
     * @brief Construct a label operation object.
     *
     * @param type  Label operation type, supporting the following categories:
     *              `LABEL_IN`, `LABEL_NOT_IN`, `LABEL_EXISTS`, `LABEL_DOES_NOT_EXIST`.
     * @param key   The key of the label.
     * @param values The value of the label.
     */
    LabelOperator(const std::string &type, const std::string &key, const std::list<std::string> &values)
        : operatorType(type), key(key), values(values)
    {
    }

    /*!
     * @brief Construct a label operation object.
     *
     * @param type  Label operation type, supporting the following categories:
     *              `LABEL_IN`, `LABEL_NOT_IN`, `LABEL_EXISTS`, `LABEL_DOES_NOT_EXIST`.
     * @param key   The key of the label.
     */
    LabelOperator(const std::string &type, const std::string &key) : operatorType(type), key(key) {}

    /*!
     * @brief Default Destructor.
     */
    ~LabelOperator() = default;

    /*!
     * @brief Get the type of the label operation object.
     *
     * @return The type of the label operation object.
     */
    std::string GetOperatorType() const
    {
        return this->operatorType;
    }

    /*!
     * @brief Get the key of the label.
     *
     * @return The key of the label.
     */
    std::string GetKey() const
    {
        return this->key;
    }

    /*!
     * @brief Get the values of the label.
     *
     * @return The values of the label.
     */
    std::list<std::string> GetValues() const
    {
        return this->values;
    }

private:
    std::string operatorType;
    std::string key;
    std::list<std::string> values;
};

/*! @class LabelInOperator affinity.h "include/yr/api/affinity.h"
 *  @brief The label operation class with the operation type `LABEL_IN`.
 */
class LabelInOperator : public LabelOperator {
public:
    /*!
     * @brief Construct a label operation object of type `LABEL_IN`.
     *
     * @param key   The key of the label.
     * @param values The values of the label.
     */
    LabelInOperator(const std::string &key, const std::list<std::string> &values) : LabelOperator(LABEL_IN, key, values)
    {
    }
};

/*! @class LabelNotInOperator affinity.h "include/yr/api/affinity.h"
 *  @brief The label operation class with the operation type `LABEL_NOT_IN`.
 */
class LabelNotInOperator : public LabelOperator {
public:
    /*!
     * @brief Construct a label operation object of type `LABEL_NOT_IN`.
     *
     * @param key   The key of the label.
     * @param values The values of the label.
     */
    LabelNotInOperator(const std::string &key, const std::list<std::string> &values)
        : LabelOperator(LABEL_NOT_IN, key, values)
    {
    }
};

/*! @class LabelExistsOperator affinity.h "include/yr/api/affinity.h"
 *  @brief The label operation class with the operation type `LABEL_EXISTS`.
 */
class LabelExistsOperator : public LabelOperator {
public:
    /*!
     * @brief Construct a label operation object of type `LABEL_EXISTS`.
     *
     * @param key The key of the label.
     */
    LabelExistsOperator(const std::string &key) : LabelOperator(LABEL_EXISTS, key) {}
};

/*! @class LabelDoesNotExistOperator affinity.h "include/yr/api/affinity.h"
 *  @brief The label operation class with the operation type `LABEL_DOES_NOT_EXIST`.
 */
class LabelDoesNotExistOperator : public LabelOperator {
public:
    /*!
     * @brief Construct a label operation object of type `LABEL_DOES_NOT_EXIST`.
     *
     * @param key The key of the label.
     */
    LabelDoesNotExistOperator(const std::string &key) : LabelOperator(LABEL_DOES_NOT_EXIST, key) {}
};

/*! @class Affinity affinity.h "include/yr/api/affinity.h"
 *  @brief This is the base class for affinity operators, which defines the common interface for affinity operators.
 */
class Affinity {
public:
    /*!
    @brief Does not support default constructor.
    */
    Affinity() = delete;

    /*!
     * @brief Construct an affinity operation object.
     *
     * @param kind  The affinity kind mainly include `RESOURCE` and `INSTANCE`. `RESOURCE` refers to predefined resource
     *              label affinity, while `INSTANCE` refers to dynamic instance label affinity.
     * @param type  The affinity type mainly include `PREFERRED`, `PREFERRED_ANTI`, `REQUIRED` and `REQUIRED_ANTI`,
     *              which represent weak affinity, weak anti-affinity, strong affinity, and strong anti-affinity,
     *              respectively.
     * @param operators Label operation list, see LabelOperator and its subclasses for details.
     */
    Affinity(const std::string &kind, const std::string &type, const std::list<LabelOperator> &operators)
        : affinityKind(kind), affinityType(type), labelOperators(operators)
    {
    }

    /*!
     * @brief Construct an affinity operation object.
     *
     * @param kind  The affinity kind mainly include `RESOURCE` and `INSTANCE`. `RESOURCE` refers to predefined resource
     *              label affinity, while `INSTANCE` refers to dynamic instance label affinity.
     * @param type  The affinity type mainly include `PREFERRED`, `PREFERRED_ANTI`, `REQUIRED` and `REQUIRED_ANTI`,
     *              which represent weak affinity, weak anti-affinity, strong affinity, and strong anti-affinity,
     *              respectively.
     * @param operators Label operation list, see LabelOperator and its subclasses for details.
     * @param affinityScope The instance affinity scope mainly includes two types: `AFFINITYSCOPE_POD`
     *                      and `AFFINITYSCOPE_NODE`.
     */
    Affinity(const std::string &kind, const std::string &type, const std::list<LabelOperator> &operators,
             const std::string &affinityScope)
        : affinityKind(kind), affinityType(type), labelOperators(operators), affinityScope(affinityScope) {
    }

    /*!
     * @brief Default Destructor.
     */
    ~Affinity() = default;

    /*!
    * @brief Get Affinity kind.
    *
    * @return The kind of affinity.
    */
    std::string GetAffinityKind() const
    {
        return this->affinityKind;
    }

    /*!
    * @brief Get Affinity type.
    *
    * @return The type of affinity.
    */
    std::string GetAffinityType() const
    {
        return this->affinityType;
    }

    /*!
    * @brief Get the list of label operation objects with affinity.
    *
    * @return The list of label operation objects with affinity.
    */
    std::list<LabelOperator> GetLabelOperators() const
    {
        return this->labelOperators;
    }

    /*!
     * @brief Get the scope of instance affinity.
     *
     * @return The scope of instance affinity.
     */
    std::string GetAffinityScope() const
    {
        return this->affinityScope;
    }

    /*!
     * @brief Set the scope of instance affinity.
     */
    void SetAffinityScope(const std::string &affinityScope)
    {
        this->affinityScope = affinityScope;
    }

private:
    std::string affinityKind;
    std::string affinityType;
    std::list<LabelOperator> labelOperators;
    std::string affinityScope;
};

/*! @class ResourcePreferredAffinity affinity.h "include/yr/api/affinity.h"
 *  @brief The affinity kind is `RESOURCE`, and the affinity type is `PREFERRED`.
 */
class ResourcePreferredAffinity : public Affinity {
public:
    /*!
     * @brief Construct a new ResourcePreferred Affinity object.
     *
     * @param labelOperator Label operation.
     */
    ResourcePreferredAffinity(const LabelOperator &labelOperator) : Affinity(RESOURCE, PREFERRED, {labelOperator}) {}

    /*!
     * @brief Construct a new ResourcePreferred Affinity object.
     *
     * @param operators Label operation list.
     */
    ResourcePreferredAffinity(const std::list<LabelOperator> &operators) : Affinity(RESOURCE, PREFERRED, operators) {}
};

/*! @class InstancePreferredAffinity affinity.h "include/yr/api/affinity.h"
 *  @brief The affinity kind is `INSTANCE`, and the affinity type is `PREFERRED`.
 */
class InstancePreferredAffinity : public Affinity {
public:
    /*!
     * @brief Construct a new InstancePreferred Affinity object.
     *
     * @param labelOperator Label operation.
     */
    InstancePreferredAffinity(const LabelOperator &labelOperator) : Affinity(INSTANCE, PREFERRED, {labelOperator}) {}

    /*!
     * @brief Construct a new InstancePreferred Affinity object.
     *
     * @param operators Label operation list.
     */
    InstancePreferredAffinity(const std::list<LabelOperator> &operators) : Affinity(INSTANCE, PREFERRED, operators) {}
};

/*! @class ResourcePreferredAntiAffinity affinity.h "include/yr/api/affinity.h"
 *  @brief The affinity kind is `RESOURCE`, and the affinity type is `PREFERRED_ANTI`.
 */
class ResourcePreferredAntiAffinity : public Affinity {
public:
    /*!
     * @brief Construct a new ResourcePreferred AntiAffinity object.
     *
     * @param labelOperator Label operation.
     */
    ResourcePreferredAntiAffinity(const LabelOperator &labelOperator)
        : Affinity(RESOURCE, PREFERRED_ANTI, {labelOperator})
    {
    }

    /*!
     * @brief Construct a new ResourcePreferred AntiAffinity object.
     *
     * @param operators Label operation list.
     */
    ResourcePreferredAntiAffinity(const std::list<LabelOperator> &operators)
        : Affinity(RESOURCE, PREFERRED_ANTI, operators)
    {
    }
};

/*! @class InstancePreferredAntiAffinity affinity.h "include/yr/api/affinity.h"
 *  @brief The affinity kind is `INSTANCE`, and the affinity type is `PREFERRED_ANTI`.
 */
class InstancePreferredAntiAffinity : public Affinity {
public:
    /*!
     * @brief Construct a new InstancePreferred AntiAffinity object.
     *
     * @param labelOperator Label operation.
     */
    InstancePreferredAntiAffinity(const LabelOperator &labelOperator)
        : Affinity(INSTANCE, PREFERRED_ANTI, {labelOperator})
    {
    }

    /*!
     * @brief Construct a new InstancePreferred AntiAffinity object.
     *
     * @param operators Label operation list.
     */
    InstancePreferredAntiAffinity(const std::list<LabelOperator> &operators)
        : Affinity(INSTANCE, PREFERRED_ANTI, operators)
    {
    }
};

/*! @class ResourceRequiredAffinity affinity.h "include/yr/api/affinity.h"
 *  @brief The affinity kind is `RESOURCE`, and the affinity type is `REQUIRED`.
 */
class ResourceRequiredAffinity : public Affinity {
public:
    /*!
     * @brief Construct a new ResourceRequired Affinity object.
     *
     * @param labelOperator Label operation.
     */
    ResourceRequiredAffinity(const LabelOperator &labelOperator) : Affinity(RESOURCE, REQUIRED, {labelOperator}) {}

    /*!
     * @brief Construct a new ResourceRequired Affinity object.
     *
     * @param operators Label operation list.
     */
    ResourceRequiredAffinity(const std::list<LabelOperator> &operators) : Affinity(RESOURCE, REQUIRED, operators) {}
};

/*! @class InstanceRequiredAffinity affinity.h "include/yr/api/affinity.h"
 *  @brief The affinity kind is `INSTANCE`, and the affinity type is `REQUIRED`.
 */
class InstanceRequiredAffinity : public Affinity {
public:
    /*!
     * @brief Construct a new InstanceRequired Affinity object.
     *
     * @param labelOperator Label operation.
     */
    InstanceRequiredAffinity(const LabelOperator &labelOperator) : Affinity(INSTANCE, REQUIRED, {labelOperator}) {}

    /*!
     * @brief Construct a new InstanceRequired Affinity object.
     *
     * @param operators Label operation list.
     */
    InstanceRequiredAffinity(const std::list<LabelOperator> &operators) : Affinity(INSTANCE, REQUIRED, operators) {}
};

/*! @class ResourceRequiredAntiAffinity affinity.h "include/yr/api/affinity.h"
 *  @brief The affinity kind is `RESOURCE`, and the affinity type is `REQUIRED_ANTI`.
 */
class ResourceRequiredAntiAffinity : public Affinity {
public:
    /*!
     * @brief Construct a new ResourceRequired AntiAffinity object.
     *
     * @param labelOperator Label operation.
     */
    ResourceRequiredAntiAffinity(const LabelOperator &labelOperator)
        : Affinity(RESOURCE, REQUIRED_ANTI, {labelOperator})
    {
    }

    /*!
     * @brief Construct a new ResourceRequired AntiAffinity object.
     *
     * @param operators Label operation list.
     */
    ResourceRequiredAntiAffinity(const std::list<LabelOperator> &operators)
        : Affinity(RESOURCE, REQUIRED_ANTI, operators)
    {
    }
};

/*! @class InstanceRequiredAntiAffinity affinity.h "include/yr/api/affinity.h"
 *  @brief The affinity kind is `INSTANCE`, and the affinity type is `REQUIRED_ANTI`.
 */
class InstanceRequiredAntiAffinity : public Affinity {
public:
    /*!
     * @brief Construct a new InstanceRequired AntiAffinity object.
     *
     * @param labelOperator Label operation.
     */
    InstanceRequiredAntiAffinity(const LabelOperator &labelOperator)
        : Affinity(INSTANCE, REQUIRED_ANTI, {labelOperator})
    {
    }

    /*!
     * @brief Construct a new InstanceRequired AntiAffinity object.
     *
     * @param operators Label operation list.
     */
    InstanceRequiredAntiAffinity(const std::list<LabelOperator> &operators)
        : Affinity(INSTANCE, REQUIRED_ANTI, operators)
    {
    }
};
}  // namespace YR