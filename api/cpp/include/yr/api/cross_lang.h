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

#pragma once
#include "yr/api/invoke_options.h"

template <typename...>
using void_t = void;

namespace YR {
namespace internal {

class CrossLangBaseType {
public:
    bool IsCrossLang(void)
    {
        return true;
    }
};

class CrossLangClass : public CrossLangBaseType {
public:
    CrossLangClass(const FunctionLanguage &lang, const std::string &moduleName, const std::string &className,
                   const std::string &funcName)
        : lang(lang), moduleName(moduleName), className(className), funcName(funcName)
    {
    }

    CrossLangClass(const CrossLangClass &other)
        : lang(other.GetLangType()),
          moduleName(other.GetModuleName()),
          className(other.GetClassName()),
          funcName(other.GetFuncName())
    {
    }

    CrossLangClass operator=(const CrossLangClass &other)
    {
        if (&other != this) {
            this->lang = other.GetLangType();
            this->moduleName = other.GetModuleName();
            this->className = other.GetClassName();
            this->funcName = other.GetFuncName();
        }
        return *this;
    }
    CrossLangClass() = default;
    virtual ~CrossLangClass() = default;

    virtual FunctionLanguage GetLangType(void) const
    {
        return lang;
    }

    virtual std::string GetModuleName(void) const
    {
        return moduleName;
    }

    virtual std::string GetClassName(void) const
    {
        return className;
    }

    virtual std::string GetFuncName(void) const
    {
        return funcName;
    }

private:
    FunctionLanguage lang;
    std::string moduleName;
    std::string className;
    std::string funcName;
};

template <class, class = void>
struct IsCrossLang : std::false_type {};

template <class T>
struct IsCrossLang<T, void_t<decltype(std::declval<T>().IsCrossLang())>> : std::true_type {};
}  // namespace internal

class CppInstanceClass : public internal::CrossLangClass {
public:
    CppInstanceClass(const std::string &creatorName)
        : internal::CrossLangClass(internal::FunctionLanguage::FUNC_LANG_CPP, "", "", creatorName)
    {
    }

    CppInstanceClass() {}

    void operator()() {}

    /*!
        @brief Create a CppInstanceClass object for invoking C++ functions.

        @code
        // C++ class definition
        class Counter {
        public:
            int count;

            Counter()
            {}

            Counter(int init)
            {
                count = init;
            }

            static Counter* FactoryCreate()
            {
                return new Counter(10);
            }

            std::string RemoteVersion()
            {
                return "RemoteActor v0";
            }
        };

        YR_INVOKE(Counter::FactoryCreate, &Counter::RemoteVersion);

        // C++ code to create and invoke the instance
        int main(void) {
            YR::Config conf;
            YR::Init(conf);

            auto cppCls = YR::CppInstanceClass::FactoryCreate("Counter::FactoryCreate");
            auto cppIns = YR::Instance(cppCls).Invoke();
            auto obj = cppIns.CppFunction<std::string>("&Counter::RemoteVersion").Invoke();
            auto res = *YR::Get(obj);
            std::cout << "add one result is " << res << std::endl;
        }
        @endcode

        @param creatorName The name of the C++ class's constructor or factory function.
        @return A CppInstanceClass object that carries the necessary information to create a C++ function class
       instance. This object can be passed to YR::Instance as a parameter.
    */
    static CppInstanceClass FactoryCreate(const std::string &creatorName)
    {
        return CppInstanceClass(creatorName);
    }
};

class PyInstanceClass : public internal::CrossLangClass {
public:
    PyInstanceClass(const std::string &moduleName, const std::string &className)
        : internal::CrossLangClass(internal::FunctionLanguage::FUNC_LANG_PYTHON, moduleName, className, "__init__")
    {
    }

    PyInstanceClass() {}

    void operator()() {}
    /*!
       @brief Creates a `PyInstanceClass` object, which can be passed as a parameter to `YR::Instance` to successfully
       create a Python function class instance.
       @param moduleName The name of the Python module.
       @param className The name of the Python class.
       @return Returns a `PyInstanceClass` object carrying function information, which can be passed as a parameter to
       `YR::Instance`.

       @snippet{trimleft} py_instance_example.cpp py instance function
     */
    static PyInstanceClass FactoryCreate(const std::string &moduleName, const std::string &className)
    {
        return PyInstanceClass(moduleName, className);
    }
};

class JavaInstanceClass : public internal::CrossLangClass {
public:
    explicit JavaInstanceClass(std::string className)
        : internal::CrossLangClass(internal::FunctionLanguage::FUNC_LANG_JAVA, "", className, "<init>")
    {
    }

    JavaInstanceClass() {}

    void operator()() {}

    /*!
        @brief Create a JavaInstanceClass object for invoking Java functions
        @code
        // Java code:
        package io.yuanrong.demo;

        // A regular Java class.
        public class Counter {
            private int value = 0;

            public int increment() {
                    this.value += 1;
                    return this.value;
                }
        }

        // C++ code:
        int main(void) {
            YR::Config conf;
            YR::Init(conf);
            auto javaInstance = YR::JavaInstanceClass::FactoryCreate("io.yuanrong.demo.Counter");
            auto r1 = YR::Instance(javaInstance).Invoke();
            auto r2 = r1.JavaFunction<int>("increment").Invoke(1);
            auto res = *YR::Get(r2);

            std::cout << "PlusOneWithJavaClass with result=" << res << std::endl;
            return res;
        }
        @endcode
        @param className The fully qualified class name of the Java class, including package name. If the class is an
       inner static class, use '$' to connect the outer class and inner class.
        @return A JavaInstanceClass object that carries the necessary information to create a Java function class
       instance. This object can be passed to YR::Instance as a parameter.
    */
    static JavaInstanceClass FactoryCreate(const std::string &className)
    {
        return JavaInstanceClass(className);
    }
};

template <typename R>
class CppClassMethod : public internal::CrossLangBaseType {
public:
    R operator()()
    {
        return {};
    }
};

template <typename R>
class PyClassMethod : public internal::CrossLangBaseType {
public:
    R operator()()
    {
        return {};
    }
};

template <typename R>
class JavaClassMethod : public internal::CrossLangBaseType {
public:
    R operator()()
    {
        return {};
    }
};

template <typename R>
class CppFunctionHandler : public internal::CrossLangBaseType {
public:
    R operator()()
    {
        return {};
    }
};

/**
 * @class PyFunctionHandler
 * @tparam R return value type
 */
template <typename R>
class PyFunctionHandler : public internal::CrossLangBaseType {
public:
    R operator()()
    {
        return {};
    }
};

template <typename R>
class JavaFunctionHandler : public internal::CrossLangBaseType {
public:
    R operator()()
    {
        return {};
    }
};

template <>
class CppClassMethod<void> : public internal::CrossLangBaseType {
public:
    void operator()() {}
};

template <>
class PyClassMethod<void> : public internal::CrossLangBaseType {
public:
    void operator()() {}
};

template <>
class JavaClassMethod<void> : public internal::CrossLangBaseType {
public:
    void operator()() {}
};

template <>
class CppFunctionHandler<void> : public internal::CrossLangBaseType {
public:
    void operator()() {}
};

template <>
class PyFunctionHandler<void> : public internal::CrossLangBaseType {
public:
    void operator()() {}
};

template <>
class JavaFunctionHandler<void> : public internal::CrossLangBaseType {
public:
    void operator()() {}
};
}  // namespace YR