
## 关于本项目

本项目提供了一个grpc+etcd实现的一个负载demo，可以用于生产，但是最好是自己实现一个负载，而不是用默认的轮询，因为不同的服务可能全职不一样。

1. 下载 protoc

[点击跳转](https://github.com/protocolbuffers/protobuf/releases) ,注意要下载系统对应的，并且设置PATH


2. 下载 proto-gen-go

由于etcd和grpc的故事很长，我们使用旧版本的proto-gen-go

```shell script
go get github.com/golang/protobuf/protoc-gen-go@v1.3.2
```

3. 生成 pb文件

```shell script
protoc -I thirty_part --go_out=plugins=grpc:api/v1/pb/  -I api/v1/pb api/v1/pb/ping.proto
```


