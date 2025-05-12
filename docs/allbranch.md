# 目录结构

如下所示，第一级目录是Github用户名（或者说命名空间），第二级目录则是C库自定义名称，第三级目录是该库的原版本号，第四级目录则是该版本号下的 `llcppg` 转换生成的 go bindings module

```
.
├── 7bitcoder/
│   └── 7bitconf/
│       └── 1.0.0/
│            └── go.mod/
│       └── 1.1.0/
└── abbyssoul/
    └── solace/
        └── 0.3.9/
```

其中，C库版本号应该是不携带v前缀的的，如果原版本号有v前缀，则删掉以满足`module path`的要求。


# 名称与版本
## 自定义名称

在all分支中，其库名称不必与`cppkg`名称保持一致，以下是自动化自定义名称策略

1. 删除所有`lib`前缀
2. 将名称中所有`-`替换成`_`



## 版本分发

为了能让Go兼容，其版本分发规则如下

```
namespace/custom_clib_name/clib_version/v0.1.0
```

`v0.1.0` 意思是我们默认所有库都不是稳定状态的，因为目前 `llcppg` 依然处于开发状态，


## 版本管理

一般来说，如果需要使用 `all` 分支，那么在 `import` 中需要包含 C库版本号，但在Go中，`module path`包含版本号可能不是最佳实践。

为了避免这个问题，需要通过在`go.mod`中进行`replace`操作

如下

```
replace github.com/goplus/llpkg/davegamble/cjson/1.7.18 v0.1.0 => github.com/goplus/llpkg/davegamble/cjson all
```

这个操作可以是通过`llgo get`自动完成的

例如，当用户拉取`all`分支的时候，

```
llgo get github.com/goplus/llpkg/davegamble/cjson/1.7.18@all
```

自动插入 `replace`


# 转换

TODO


# 测试
由于 `all` 分支 是批量转换而来，因此没有方法可以编写相关测试用例

为了验证其是否可用，暂时通过 `llgo build` 方式进行检查




