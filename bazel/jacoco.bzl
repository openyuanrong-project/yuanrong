package(default_visibility = ["//visibility:public"])

java_import(
    name = "jacoco_agent",
    jars = ["lib/jacocoagent.jar"]
)
