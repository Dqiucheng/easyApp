使用文档
===============
#  1、下载地址
~~~
https://github.com/Dqiucheng/easyApp/
~~~
下载完成后放入src目录

# 2、安装部署
2.1、如需改包名，需同时修改引入地址，全局将`easyApp/`替换为`包名/`

2.2、进入项目下的config目录，将`tomlConfigDemo`更名为`tomlConfig`

2.3、加载依赖包
~~~
go mod init

go mod tidy
~~~

2.4、运行
~~~
go run main.go
~~~

# 3、安装air平滑重启包（生产环境不建议使用）
~~~
go get github.com/cosmtrek/air
~~~
> 进入项目目录

3.1、普通启动
~~~
air
~~~
3.2、后台启动
~~~
nohup air &
~~~

## 4、查看性能指标
> 注：需要先安装 graphviz 下载地址：https://www.graphviz.org/download/
* web访问，127.0.0.1:8080/debug/pprof
* 控制台访问，go tool pprof http://127.0.0.1:8080/debug/pprof/profile

4.1、查看图形界面
> 注：需要将FlameGraph工具安装目录配置到path环境变量里，下载地址：github.com/brendangregg/FlameGraph

4.2、查看火焰图
~~~
# 使用示例
go-torch -u http://127.0.0.1:8080/debug/pprof/profile -p > profile.svg
~~~