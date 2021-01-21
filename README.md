#### minio 配置和golang api集成文档
> minio说明
```text
MinIO is a High Performance Object Storage released under Apache License v2.0.
It is API compatible with Amazon S3 cloud storage service. 
Use MinIO to build high performance infrastructure for machine learning, analytics and application data workloads.
更多的minio信息可以查阅官网：
https://min.io/
```
> minioPerfect
```text
minioPerfect是一个自建minio server的客户端，根据官方golang api选择基本功能组合而来。
只提供常用的客户端基本操作功能。
官方完整客户端mc下载链接如下：
https://dl.min.io/client/mc/release/linux-amd64/mc
minio server是个人所设想的服务器初始化系统化重要的一环，它具有部署便捷，高性能，文档丰富等优点，
最重要的是不仅可以实现资源集中管理，又可以解决初始化时重复拷贝的问题。
```
> minio启动
```shell script
#下载服务器端
wget https://dl.min.io/server/minio/release/darwin-amd64/minio

#设置MINIO_ROOT_USER 和 MINIO_ROOT_PASSWORD
vim /etc/profile
export MINIO_ROOT_USER=minio
export MINIO_ROOT_PASSWORD=fzrjocr7xJAcrxdovq1_
#启动
minio server --address ip:port data_path
```
> 自定义api功能
```text
下载文件: minioPerfect dl bucketname objectname path
上传文件：minioPerfect put bucketname objectname
创建桶  ：minioPerfect mb bucketname
桶列表  ：minioPerfect ls
桶内文件：minioPerfect lsobj bucketname
删除文件: minioPerfect rm bucketname prefix 【前缀匹配批量删除】
```
> 分布式集群功能
```text
可以部署多server 通过nginx代理实现集群
```
> 开启生成prometheus配置
```shell script
[root@game3 export]# ./mc admin prometheus generate myminio
scrape_configs:
- job_name: minio-job
  bearer_token: eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjQ3NjQ3MDk3NTIsImlzcyI6InByb21ldGhldXMiLCJzdWIiOiJtaW5pbyJ9.55716_-gsxe_-ntofIUe8fBAhGrG6XUCQQk8xjw5nJSrMjnw_JRohhWdEeTzbt7xLpgt-PSJoJEcJAGmbofa8w
  metrics_path: /minio/prometheus/metrics
  scheme: http
  static_configs:
  - targets: ['192.168.137.132:9088']
```
> 增加redis消息通知
```shell script
#mc导出config配置
[root@game3 export]# ./mc admin config export myminio > /tmp/my-serverconfig 
#修改配置
notify_redis enable=true format=namespace address=127.0.0.1:6379 key=bucketevents password=doumaoxin queue_dir= queue_limit=0
#导回配置
[root@game3 export]# ./mc admin config import myminio < /tmp/my-serverconfig --insecure
#重启server，成功的话启动时会输出：SQS ARNs: arn:minio:sqs::_:redis
#使用minio 客户端启用bucket通知
[root@game3 export]# ./mc event add myminio/mhxzx arn:minio:sqs::_:redis --suffix .sh
Successfully added arn:minio:sqs::_:redis
# 可以使用--prefix --suffix 匹配

#查看事件匹配
[root@game3 export]# ./mc event list myminio/mhxzx
arn:minio:sqs::_:redis   s3:ObjectCreated:*,s3:ObjectRemoved:*,s3:ObjectAccessed:*   Filter: suffix=".sh"
#添加测试文件test.sh
[root@game3 ~]# minioPerfect put mhxzx test.sh
{
        "Bucket": "mhxzx",
        "Key": "test.sh",
        "ETag": "d41d8cd98f00b204e9800998ecf8427e",
        "Size": 0,
        "LastModified": "0001-01-01T00:00:00Z",
        "Location": "",
        "VersionID": "",
        "Expiration": "0001-01-01T00:00:00Z",
        "ExpirationRuleID": ""
}

#查看redis
127.0.0.1:6379> HGET bucketevents mhxzx/test.sh
"{\"Records\":[{\"eventVersion\":\"2.0\",\"eventSource\":\"minio:s3\",\"awsRegion\":\"\",\"eventTime\":\"2021-01-20T04:44:17.444Z\",\"eventName\":\"s3:ObjectCreated:Put\",\"userIdentity\":{\"principalId\":\"minio\"},\"requestParameters\":{\"accessKey\":\"minio\",\"region\":\"\",\"sourceIPAddress\":\"192.168.137.132\"},\"responseElements\":{\"content-length\":\"0\",\"x-amz-request-id\":\"165BD727C0233802\",\"x-minio-deployment-id\":\"82f56252-cdf5-4108-b463-432107e3c717\",\"x-minio-origin-endpoint\":\"http://192.168.137.132:9088\"},\"s3\":{\"s3SchemaVersion\":\"1.0\",\"configurationId\":\"Config\",\"bucket\":{\"name\":\"mhxzx\",\"ownerIdentity\":{\"principalId\":\"minio\"},\"arn\":\"arn:aws:s3:::mhxzx\"},\"object\":{\"key\":\"test.sh\",\"eTag\":\"d41d8cd98f00b204e9800998ecf8427e\",\"contentType\":\"application/octet-stream\",\"userMetadata\":{\"content-type\":\"application/octet-stream\"},\"sequencer\":\"165BD727C0495E7B\"}},\"source\":{\"host\":\"192.168.137.132\",\"port\":\"\",\"userAgent\":\"MinIO (linux; amd64) minio-go/v7.0.7\"}}]}"
#后续可接入elasticsearch等
```