load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "admission.go",
        "config.go",
        "doc.go",
        "domain.go",
        "plugin.go",
    ],
    importpath = "k8s.io/kubernetes/plugin/pkg/admission/alwaysqualifyimages",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/core:go_default_library",
        "//pkg/util/parsers:go_default_library",
        "//plugin/pkg/admission/alwaysqualifyimages/apis/alwaysqualifyimages:go_default_library",
        "//plugin/pkg/admission/alwaysqualifyimages/apis/alwaysqualifyimages/install:go_default_library",
        "//plugin/pkg/admission/alwaysqualifyimages/apis/alwaysqualifyimages/v1alpha1:go_default_library",
        "//plugin/pkg/admission/alwaysqualifyimages/apis/alwaysqualifyimages/validation:go_default_library",
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/api/errors:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apimachinery/announced:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apimachinery/registered:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime/serializer:go_default_library",
        "//vendor/k8s.io/apiserver/pkg/admission:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "admission_test.go",
        "domain_test.go",
    ],
    importpath = "k8s.io/kubernetes/plugin/pkg/admission/alwaysqualifyimages",
    library = ":go_default_library",
    deps = [
        "//pkg/apis/core:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime:go_default_library",
        "//vendor/k8s.io/apiserver/pkg/admission:go_default_library",
    ],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [
        ":package-srcs",
        "//plugin/pkg/admission/alwaysqualifyimages/apis/alwaysqualifyimages:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
