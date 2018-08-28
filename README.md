# JinyGo 使用文档

JinyGo API Web Framework

## codis版本
0.0.1

#### 使用

```bash

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinygo"
)

func main() {
	jiny := jinygo.New()
	v1 := jiny.RGroup("v1")
	{
		v1.Get("index", Index)
	}
	jiny.Run()
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
$ export CONFIG_PATH=/data/app/jiny

jiny := jinygo.New()
jiny.SetEnvPrefix("jiny")
...

or

$ export JINY_CONFIG_PATH=/data/app/jiny

```

jiny文件夹下必须有名为app.yml的配置文件，具体配置参照example/conf/app.yml

