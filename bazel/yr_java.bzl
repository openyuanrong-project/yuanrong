load("@rules_java//java:defs.bzl", "java_binary", "java_library", "java_test")

def group_of_java_tests(test_suite_name, test_classes, **kwargs):
    generated_tests = []

    for test_class in test_classes:
        name = test_suite_name + "_" + test_class.replace('.', '_')
        java_test(
            name = name,
            test_class = test_class,
            **kwargs
        )
        generated_tests.append(name)

    native.test_suite(
        name = test_suite_name,
        tests = generated_tests,
    )
