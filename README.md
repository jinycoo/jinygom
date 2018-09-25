# 使用文档

## 版本
1.0.0

#### 使用

```bash

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinygo"
)

func main() {
	jg := jinygo.New()
	v1 := jg.RGroup("v1")
	{
		v1.Get("index", Index)
	}
	jg.Run()
}

func Index(c *gin.Context) {
	c.JSON(200, gin.H{"jinygo: "good job"})
}

```

#### 2. 设置配置文件

默认配置路径为 编译后文件所在文件夹下 conf文件夹下
配置app.yml 及 db cache等配置文件

如不适用默认配置路径，操作如下：
```bash
$ export CONFIG_PATH=/etc/jinygo/conf

// 添加ENV prefix  （推荐）
jg := jinygo.New()
jg.SetEnvPrefix("jg")
...

or

$ export JG_CONFIG_PATH=/etc/jinygo/conf

```

conf文件夹下必须有名为app.yml的配置文件

