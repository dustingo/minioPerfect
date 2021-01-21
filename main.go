package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"os"
	"strings"
)

type perfectMinio struct {
	minioClient *minio.Client
	bucketname  string
	objectname  string
	path        string
}

// object struct
type obj struct {
	ObjName string `json:objname`
	ObjSize int64  `json:objsize`
	ObjEtag string `json:objetag`
}

// 操作结果
type result struct {
	Action     string
	Status     string
	StatusInfo string
}

var res result

// 桶子metadata
var bucketInfo minio.BucketInfo

//帮助信息
const help = `
    下载文件: minioPerfect dl bucketname objectname
    上传文件：minioPerfect put bucketname objectname
    创建桶  ：minioPerfect mb bucketname
    桶列表  ：minioPerfect ls
    桶内文件：minioPerfect lsobj bucketname
    删除文件: minioPerfect rm bucketname prefix 【前缀匹配批量删除】
`

// json序列化
func MarshToJson(s *result, bucketname string, err error) string {
	s.Action = os.Args[1]
	if err != nil {
		s.Status = "failed"
		s.StatusInfo = err.Error()
	} else {
		s.Status = "success"
	}
	outStr, _ := json.MarshalIndent(s, "", "\t")
	return string(outStr)
}

// 下载文件
func (p *perfectMinio) downloadFile(client *minio.Client, bucketname string, objectname string, path string) {
	err := client.FGetObject(context.Background(), bucketname, objectname, path, minio.GetObjectOptions{})
	fmt.Println(MarshToJson(&res, bucketname, err))

}

// 上传文件
func (p *perfectMinio) uploadFile(client *minio.Client, bucketname string, objectname string) {
	file, err := os.Open(objectname)
	if err != nil {
		fmt.Println(MarshToJson(&res, bucketname, err))
		return
	}
	defer file.Close()
	fileStat, err := file.Stat()
	if err != nil {
		fmt.Println(&res, bucketname, err)
		return
	}
	uploadInfo, err := client.PutObject(context.Background(), bucketname, objectname, file, fileStat.Size(), minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		fmt.Println(MarshToJson(&res, bucketname, err))
		return
	}
	outStr, _ := json.MarshalIndent(uploadInfo, "", "\t")
	fmt.Println(string(outStr))
}

// 删除文件
// 官方文档有误。RemoveObjects的参数objectsCh是一个ObjectInfo类型的chan，但是官方文档却将objectsCh初始化成string类型的chanobjectsCh := make(chan string)
func (p *perfectMinio) removeFile(client *minio.Client, bucketname string, objectname string) {
	objectsCh := make(chan minio.ObjectInfo)
	go func() {
		defer close(objectsCh)
		for object := range client.ListObjects(context.Background(), bucketname, minio.ListObjectsOptions{Prefix: objectname, Recursive: true}) {
			objectsCh <- object
		}
	}()

	opts := minio.RemoveObjectsOptions{GovernanceBypass: true}
	for rErr := range client.RemoveObjects(context.Background(), bucketname, objectsCh, opts) {
		fmt.Println("begin to delete")
		fmt.Println(MarshToJson(&res, bucketname, fmt.Errorf("ObjectName:%s,VersionID:%s,err:%s", rErr.ObjectName, rErr.Err)))
	}
}

// 列出桶子里的文件对象信息
func (p *perfectMinio) listObject(client *minio.Client, bucketname string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	objects := client.ListObjects(ctx, bucketname, minio.ListObjectsOptions{
		Recursive: true,
	})
	for object := range objects {
		var file obj
		file.ObjEtag = object.ETag
		file.ObjName = object.Key
		file.ObjSize = object.Size
		outStr, err := json.MarshalIndent(&file, "", "\t")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(outStr))
	}
}

// 创建筒子
func (p *perfectMinio) makeBucket(client *minio.Client, bucketname string) {
	found, err := client.BucketExists(context.Background(), bucketname)
	if err != nil {
		fmt.Println(err)
		return
	}
	if found {
		fmt.Println(MarshToJson(&res, bucketname, fmt.Errorf("%s already exists", bucketname)))
	} else {
		err = client.MakeBucket(context.Background(), bucketname, minio.MakeBucketOptions{Region: "cn-north-1", ObjectLocking: false})
		fmt.Println(MarshToJson(&res, bucketname, err))
	}
}

// 列出筒子列表
func (p *perfectMinio) ls(client *minio.Client) {
	buckets, err := client.ListBuckets(context.Background())
	if err != nil {
		fmt.Println(MarshToJson(&res, "", err))
		return
	}
	for _, bucket := range buckets {
		bucketInfo.Name = bucket.Name
		bucketInfo.CreationDate = bucket.CreationDate
		outStr, _ := json.MarshalIndent(&bucketInfo, "", "\t")
		fmt.Println(string(outStr))
	}
}

//mc alias set myminio http://192.168.137.132:9000 minio fzrjocr7xJAcrxdovq1_
func main() {

	// init client
	minioClient, err := minio.New("10.133.39.14:9876", &minio.Options{
		Creds:  credentials.NewStaticV4("minio", "fzrjocr7xJAcrxdovq1_", ""),
		Secure: false,
	})
	if err != nil {
		println("found err", err)
		return
	}
	var myminio perfectMinio
	if os.Args[1] == "dl" {
		myminio.minioClient = minioClient
		myminio.bucketname = os.Args[2]
		myminio.objectname = os.Args[3]
		myminio.path = strings.Join([]string{os.Args[4], os.Args[3]}, "/")
		myminio.downloadFile(myminio.minioClient, myminio.bucketname, myminio.objectname, myminio.path)
	} else if os.Args[1] == "lsobj" {
		myminio.minioClient = minioClient
		myminio.bucketname = os.Args[2]
		myminio.listObject(myminio.minioClient, myminio.bucketname)
	} else if os.Args[1] == "mb" {
		myminio.minioClient = minioClient
		myminio.bucketname = os.Args[2]
		myminio.makeBucket(myminio.minioClient, myminio.bucketname)
	} else if os.Args[1] == "ls" {
		myminio.minioClient = minioClient
		myminio.ls(myminio.minioClient)
	} else if os.Args[1] == "put" {
		myminio.minioClient = minioClient
		myminio.bucketname = os.Args[2]
		myminio.objectname = os.Args[3]
		myminio.uploadFile(myminio.minioClient, myminio.bucketname, myminio.objectname)
	} else if os.Args[1] == "rm" {
		myminio.minioClient = minioClient
		myminio.bucketname = os.Args[2]
		myminio.objectname = os.Args[3]
		myminio.removeFile(myminio.minioClient, myminio.bucketname, myminio.objectname)
	} else if os.Args[1] == "--help" || os.Args[1] == "-h" {
		fmt.Println(help)
	} else {
		res.Action = os.Args[1]
		res.Status = "failed"
		res.StatusInfo = "invalid action"
		outStr, _ := json.MarshalIndent(res, "", "\t")
		fmt.Println(string(outStr))
	}
}
