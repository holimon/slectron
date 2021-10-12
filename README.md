# slectron
## 说明
### 目的
* 忽略golang和web之间的IPC耦合逻辑，互操作性逻辑完全由用户自定义
* 提供在线或离线方式用于静态资源或一般性网页的渲染环境部署
### 特性
* 基于electron为golang提供基础可用的UI(静态资源和一般网页等)渲染环境
* 在实例初始化过程中即检查vendor或系统是否存在electron环境
* 默认基于taobao镜像源提供快速的electron部署方式；并支持镜像源修改和自定义离线模式镜像源
* 配置运行参数可为一般性URL、静态资源入口文件或asar包等
* 支持并赞同UI静态资源内嵌，实际运行之前完成静态资源解包
### 其他
* 感谢go-astlectron项目和其作者
## 示例
### 示例1
> HTTP URL演示
```
package main

import (
	"github.com/holimon/slectron"
)

func main() {
	if s, e := slectron.New(slectron.Options{ElectronVersion: "14.0.1", ElectronParam: "https://www.baidu.com"}); e == nil {
		s.Start()
	}
}

```
### 示例2
> 自定义electron源和静态资源内嵌(go.rice)的简单使用演示，其他内嵌库亦适用

> Tips: go.rice embed-go 需注意变量、init函数和main函数初始化顺序
```
package main

import (
	rice "github.com/GeertJohan/go.rice"
	"github.com/holimon/slectron"
)

func main() {
	assets := rice.MustFindBox("assets")
	cache := rice.MustFindBox("cache")
	s, _ := slectron.New(slectron.Options{ElectronVersion: "14.0.1", CustomVendorer: func() (content []byte, err error) {
		return cache.Bytes("electron.zip")
	}})
	s.AssetsWrite(func() (name string, content []byte, err error) {
		name = "app.asar"
		content, err = assets.Bytes("app.asar")
		return
	})
	param, _ := s.AssetsQuote("app.asar")
	s.SetExecuteArgs(param)
	go s.Start()
	s.Wait()
}


```