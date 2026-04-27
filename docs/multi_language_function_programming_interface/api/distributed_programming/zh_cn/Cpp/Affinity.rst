Affinity
==========

.. cpp:class:: Affinity

    这是亲和操作符的基类，它定义了亲和操作符的通用接口。

    由以下类继承：YR::InstancePreferredAffinity、YR::InstancePreferredAntiAffinity、YR::InstanceRequiredAffinity、YR::InstanceRequiredAntiAffinity、YR::ResourcePreferredAffinity、YR::ResourcePreferredAntiAffinity、YR::ResourceRequiredAffinity、YR::ResourceRequiredAntiAffinity。

    **公共函数**
 
    .. cpp:function:: Affinity() = delete
       
        不支持默认构造函数。

    .. cpp:function:: Affinity(const std::string &kind, const std::string &type, const std::list<LabelOperator> &operators)
       
        构造一个亲和操作对象。

        参数：
            - **kind** - 亲和的标签种类，主要包括 RESOURCE 和 INSTANCE。其中 Resource 为预置的资源标签亲和，Instance 为动态的实例标签亲和。
            - **type** - 亲和具体类型，主要包括 PREFERRED、PREFERRED_ANTI、REQUIRED 和 REQUIRED_ANTI，分别表示弱亲和、弱反亲和、强亲和和强反亲和。
            - **operators** - 标签操作列表，详细信息请参见 LabelOperator 及其子类。
    
    .. cpp:function:: Affinity(const std::string &kind, const std::string &type, const std::list<LabelOperator> &operators, const std::string &affinityScope)
       
        构造一个亲和操作对象。
    
        参数：
            - **kind** - 亲和的标签种类，主要包括 RESOURCE 和 INSTANCE。RESOURCE 表示预定义的资源标签亲和，而 INSTANCE 表示动态实例标签亲和。
            - **type** - 亲和具体类型，主要包括 PREFERRED、PREFERRED_ANTI、REQUIRED 和 REQUIRED_ANTI，分别表示弱亲和、弱反亲和、强亲和和强反亲和。
            - **operators** - 标签操作列表，详细信息请参见 LabelOperator 及其子类。
            - **affinityScope** - 实例亲和范围主要包括两种类型：AFFINITYSCOPE_POD 和 AFFINITYSCOPE_NODE。

    .. cpp:function:: ~Affinity() = default
       
        默认析构函数。

    .. cpp:function:: std::string GetAffinityKind() const
       
        获取亲和的标签种类。

        返回：
            亲和的种类。
    
    .. cpp:function:: std::string GetAffinityType() const

        获取亲和类型。

        返回：
            亲和的类型。
    
    .. cpp:function:: std::list<LabelOperator> GetLabelOperators() const

        获取带有亲和的标签操作对象列表。
    
        返回：
            带有亲和的标签操作对象列表。

    .. cpp:function:: std::string GetAffinityScope() const

        获取实例亲和的范围。
        
        返回：
            实例亲和的范围。

    .. cpp:function:: void SetAffinityScope(const std::string &affinityScope)

        设置实例亲和的范围。

.. cpp:class:: LabelOperator
    
    这是标签操作符的基类，它定义了标签操作符的通用接口。

    由以下类继承：YR::LabelDoesNotExistOperator、YR::LabelExistsOperator、YR::LabelInOperator、YR::LabelNotInOperator。

    **公共函数**
 
    .. cpp:function:: LabelOperator() = delete
       
        不支持默认构造函数。
  
    .. cpp:function:: LabelOperator(const std::string &type, const std::string &key, const std::liststd::string &values)
       
        构造一个标签操作对象。

        参数：
            - **type**：标签操作类型，支持以下类别：LABEL_IN、LABEL_NOT_IN、LABEL_EXISTS、LABEL_DOES_NOT_EXIST。
            - **key**：标签的键。
            - **values**：标签的值。
  
    .. cpp:function:: LabelOperator(const std::string &type, const std::string &key)
       
        构造一个标签操作对象。
    
        参数：
            - **type**：标签操作类型，支持以下类别：LABEL_IN、LABEL_NOT_IN、LABEL_EXISTS、LABEL_DOES_NOT_EXIST。
            - **key**：标签的键。

    .. cpp:function:: ~LabelOperator() = default
       
        默认构造函数。
    
    .. cpp:function:: std::string GetOperatorType() const
       
        获取标签操作对象的类型。

        返回：
            标签操作对象的类型。
    
    .. cpp:function:: std::string GetKey() const
       
        获取标签的键。
    
        返回：
            标签的键。
    
    .. cpp:function:: std::list<std::string> GetValues() const
       
        获取标签的值。
        
        返回：
            标签的值。

.. cpp:class:: ResourcePreferredAffinity : public YR::Affinity
    
    亲和种类为 RESOURCE，亲和类型为 PREFERRED。

    **公共函数**

    .. cpp:function:: ResourcePreferredAffinity(const LabelOperator &labelOperator)
   
        构造一个新的 ResourcePreferred 亲和对象。

        参数：
            - **labelOperator**：标签操作。

    .. cpp:function:: ResourcePreferredAffinity(const std::list<LabelOperator> &operators)
   
        构造一个新的 ResourcePreferred 亲和对象。
    
        参数：
            - **operators**：标签操作列表。
  
.. cpp:class:: InstancePreferredAffinity : public YR::Affinity
    
    亲和种类为 INSTANCE，亲和类型为 PREFERRED。

    **公共函数**

    .. cpp:function:: InstancePreferredAffinity(const LabelOperator &labelOperator)

        构造一个新的 InstancePreferred 亲和对象。

        参数：
            - **labelOperator**：标签操作。

    .. cpp:function:: InstancePreferredAffinity(const std::list<LabelOperator> &operators)

        构造一个新的 InstancePreferred 亲和对象。

        参数：
            - **operators**：标签操作列表。

.. cpp:class:: ResourceRequiredAffinity : public YR::Affinity
    
    亲和类型为 RESOURCE，亲和种类为 REQUIRED。

    **公共函数**

    .. cpp:function:: ResourceRequiredAffinity(const LabelOperator &labelOperator)

        构造一个新的 ResourceRequired 亲和对象。

        参数：
            - **labelOperator**：标签操作。

    .. cpp:function:: ResourceRequiredAffinity(const std::list<LabelOperator> &operators)

        构造一个新的 ResourceRequired 亲和对象。

        参数：
            - **operators**：标签操作列表。

.. cpp:class:: InstanceRequiredAffinity : public YR::Affinity
    
    亲和种类为 INSTANCE，亲和类型为 REQUIRED。

    **公共函数**

    .. cpp:function:: InstanceRequiredAffinity(const LabelOperator &labelOperator)

        构造一个新的 InstanceRequired 亲和对象。

        参数：
            - **labelOperator**：标签操作。

    .. cpp:function:: InstanceRequiredAffinity(const std::list<LabelOperator> &operators)

        构造一个新的 InstanceRequired 亲和对象。

        参数：
            - **operators**：标签操作列表。

.. cpp:class:: ResourcePreferredAntiAffinity : public YR::Affinity
    
    亲和种类为 RESOURCE，亲和类型为 PREFERRED_ANTI。

    **公共函数**

    .. cpp:function:: ResourcePreferredAntiAffinity(const LabelOperator &labelOperator)

        构造一个新的 ResourcePreferred 反亲和对象。

        参数：
            - **labelOperator**：标签操作。

    .. cpp:function:: ResourcePreferredAntiAffinity(const std::list<LabelOperator> &operators)

        构造一个新的 ResourcePreferred 反亲和对象。

        参数：
            - **operators**：标签操作列表。

.. cpp:class:: InstancePreferredAntiAffinity : public YR::Affinity
    
    亲和种类为 INSTANCE，亲和类型为 PREFERRED_ANTI。

    **公共函数**

    .. cpp:function:: InstancePreferredAntiAffinity(const LabelOperator &labelOperator)

        构造一个新的 InstancePreferred 反亲和对象。

        参数：
            - **labelOperator**：标签操作。

    .. cpp:function:: InstancePreferredAntiAffinity(const std::list<LabelOperator> &operators)

        构造一个新的 InstancePreferred 反亲和对象。

        参数：
            - **operators**：标签操作列表。
  
.. cpp:class:: ResourceRequiredAntiAffinity : public YR::Affinity
        
    亲和种类为 RESOURCE，亲和类型为 REQUIRED_ANTI。

    **公共函数**

    .. cpp:function:: ResourceRequiredAntiAffinity(const LabelOperator &labelOperator)

        构造一个新的 ResourceRequired 反亲和对象。

        参数：
            - **labelOperator**：标签操作。

    .. cpp:function:: ResourceRequiredAntiAffinity(const std::list<LabelOperator> &operators)

        构造一个新的 ResourceRequired 反亲和对象。

        参数：
            - **operators**：标签操作列表。

.. cpp:class:: InstanceRequiredAntiAffinity : public YR::Affinity
        
    亲和种类为 INSTANCE，亲和类型为 REQUIRED_ANTI。

    **公共函数**

    .. cpp:function:: InstanceRequiredAntiAffinity(const LabelOperator &labelOperator)

        构造一个新的 InstanceRequired 反亲和对象。

        参数：
            - **labelOperator**：标签操作。

    .. cpp:function:: InstanceRequiredAntiAffinity(const std::list<LabelOperator> &operators)

        构造一个新的 InstanceRequired 反亲和对象。

        参数：
            - **operators**：标签操作列表。
  
.. cpp:class:: LabelInOperator : public YR::LabelOperator
        
    标签操作类，操作类型为 LABEL_IN。

    **公共函数**

    .. cpp:function:: LabelInOperator(const std::string &key, const std::liststd::string &values)

        构造一个类型为 LABEL_IN 的标签操作对象。

        参数：
            - **key**：标签操作。
            - **values**：标签的值列表。

.. cpp:class:: LabelNotInOperator : public YR::LabelOperator

    标签操作类，操作类型为 LABEL_NOT_IN。

    **公共函数**

    .. cpp:function:: LabelNotInOperator(const std::string &key, const std::liststd::string &values)

        构造一个类型为 LABEL_NOT_IN 的标签操作对象。

        参数：
            - **key**：标签的键。
            - **values**：标签的值列表。

.. cpp:class:: LabelExistsOperator : public YR::LabelOperator

    标签操作类，操作类型为 LABEL_EXISTS。

    **公共函数**

    .. cpp:function:: LabelExistsOperator(const std::string &key)

        构造一个类型为 LABEL_EXISTS 的标签操作对象。

        参数：
            - **key**：标签的键。

.. cpp:class:: LabelDoesNotExistOperator : public YR::LabelOperator

    标签操作类，操作类型为 LABEL_DOES_NOT_EXIST。

    **公共函数**

    .. cpp:function:: LabelDoesNotExistOperator(const std::string &key)

        构造一个类型为 LABEL_DOES_NOT_EXIST 的标签操作对象。

        参数：
            - **key**：标签的键。

参数补充说明如下：

.. cpp:var:: const std::string YR::RESOURCE = "Resource"

    亲和种类，表示资源标签亲和。

.. cpp:var:: const std::string YR::INSTANCE = "Instance"
    
    亲和种类，表示动态实例标签亲和。

.. cpp:var:: const std::string YR::PREFERRED = "PreferredAffinity"

    亲和类型，表示弱亲和（推荐亲和）。

.. cpp:var:: const std::string YR::PREFERRED_ANTI = "PreferredAntiAffinity"

    亲和类型，表示弱反亲和（推荐反亲和）。

.. cpp:var:: const std::string YR::REQUIRED = "RequiredAffinity"

    亲和类型，表示强亲和（必需亲和）。

.. cpp:var:: const std::string YR::REQUIRED_ANTI = "RequiredAntiAffinity"
    
    亲和类型，表示强反亲和（必需反亲和）。

.. cpp:var:: const std::string YR::AFFINITYSCOPE_POD = "POD"

    实例亲和范围，表示 Pod 级别的亲和。

.. cpp:var:: const std::string YR::AFFINITYSCOPE_NODE = "NODE"
    
    实例亲和范围，表示 Node 级别的亲和。

.. cpp:var:: const std::string YR::LABEL_IN = "LabelIn"

    标签操作类型，表示标签具有对应的值。

.. cpp:var:: const std::string YR::LABEL_NOT_IN = "LabelNotIn"

    标签操作类型，表示标签不具有对应的值。

.. cpp:var:: const std::string YR::LABEL_EXISTS = "LabelExists"
    
    标签操作类型，表示存在对应的标签。

.. cpp:var:: const std::string YR::LABEL_DOES_NOT_EXIST = "LabelDoesNotExist"
    
    标签操作类型，表示不存在对应的标签。