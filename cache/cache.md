*** Used

cache.Init(file) file 为配置文件 例如：cache.yml

```bash
redis:
  cluster: sentinel  # 单机模式 留空或standalone  哨兵模式 sentinel
  master:
    protocol:
    host: mastername
    port:
    password: masterpassword
    db: 0
  sentinel:
    - 192.168.0.100:26379
    - 192.168.0.101:26379
    - 192.168.0.102:26379
```

cache.RCache.Set("test_key", "test_value", 0)
cache.RCache.Get("test_key")