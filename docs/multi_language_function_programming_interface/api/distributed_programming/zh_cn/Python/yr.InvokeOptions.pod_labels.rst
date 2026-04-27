.. _pod_labels_IO:

yr.InvokeOptions.pod_labels
--------------------------------

.. py:attribute:: InvokeOptions.pod_labels
   :type: Dict[str, str]

   Pod 标签仅在 Kubernetes 环境中使用。当创建函数实例时，`pod_labels` 可以接收用户提供的键值对并将其传递给函数系统。

   在 `ActorPattern` 函数实例完成特化（Running 状态）后，Scaler 会将传入的标签应用到 POD 上。
   
   当 `ActorPattern` 函数实例失败或被删除时，Scaler 会将 POD 的对应标签设置为空（移除它）。
   
   约束：
       - `pod_labels` 中可以存储的标签数量不能超过 5 个。
   
       `pod_labels` 中的键和值的约束条件：

         - 键（key）：支持大小写字母、数字以及中划线，允许长度为 1-63 个字符，不能以中划线开头或结尾，不允许空字符串。
         - 值（value）：支持大小写字母、数字以及中划线，长度为 1-63 个字符，不能以中划线开头或结尾，允许空字符串。
   
   异常：
       当用户传入的 `pod_labels` 不满足约束条件时，会抛出对应的异常及错误信息。