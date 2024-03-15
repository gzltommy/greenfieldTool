package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/prysmaticlabs/prysm/time"
	"github.com/zorrotommy/greenfieldTool/greenfield"
	"golang.org/x/sync/errgroup"
	"os"
	"strings"
)

type StringSlice []string

func (ss *StringSlice) String() string {
	return fmt.Sprintf("%v", []string(*ss))
}

func (ss *StringSlice) Set(value string) error {
	*ss = append(*ss, value)
	return nil
}

func main() {
	var (
		privateKey StringSlice
		bucketName string
		pathName   string
		testNet    int
	)

	flag.Var(&privateKey, "pk", "有写入通权限的私钥,如: -pk=0x1.. -pk=0x2..")
	flag.StringVar(&bucketName, "bn", "", "存储桶名字")
	flag.StringVar(&pathName, "pn", "", "文件路径")
	flag.IntVar(&testNet, "test", 1, "默认是测试环境，1测试环境（默认）,0正式环境")
	flag.Parse()

	// 打印
	fmt.Printf("传参：privateKey=%v bucketName=%v pathName=%v test=%v\n\n", privateKey, bucketName, pathName, testNet)

	// 是否有效
	if len(privateKey) == 0 || bucketName == "" || pathName == "" {
		fmt.Println("无效的参数！")
		return
	}

	err := greenfield.Init(privateKey, bucketName, testNet == 1)
	if err != nil {
		fmt.Printf("init fail.err:%v\n", err)
		return
	}

	//greenfield.UploadResource()
	allFile, err := GetAllFile(pathName)
	if err != nil {
		fmt.Printf("GetAllFile fail.err:%v\n", err)
		return
	}

	errGroup, _ := errgroup.WithContext(context.Background())
	errGroup.SetLimit(2)
	t1 := time.Now().Unix()
	fmt.Println("开始")
	for _, f := range allFile {
		// 更新当前最新区块号
		errGroup.Go(func() error {
			UploadFile(f)
			return nil
		})
	}
	_ = errGroup.Wait()
	fmt.Printf("总耗时：%v s \n", time.Now().Unix()-t1)
}

func UploadFile(fileName string) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		// 没有完成
		fmt.Printf("失败：ReadFile(%v) fail.err:%v \n", fileName, err)
		return
	}

	fileName = strings.TrimLeft(fileName, `./`)
	fileName = strings.TrimLeft(fileName, `.\`)
	fileName = strings.TrimLeft(fileName, `.`)

	path, err := greenfield.UploadResourceWithFilePath(data, fileName)
	if err != nil {
		fmt.Printf("失败：UploadResourceWithFilePath(%v) fail.err:%v \n", fileName, err)
		return
	}
	fmt.Printf("%s upload succeed! \n", path)
}

func GetAllFile(pathname string) ([]string, error) {
	result := make([]string, 0, 10)
	dirEntry, err := os.ReadDir(pathname)
	if err != nil {
		fmt.Printf("读取文件目录失败，pathname=%v, err=%v \n", pathname, err)
		return result, err
	}

	// 所有文件/文件夹
	for _, fi := range dirEntry {
		fullname := pathname + "/" + fi.Name()
		// 是文件夹则递归进入获取;是文件，则压入数组
		if fi.IsDir() {
			temp, err := GetAllFile(fullname)
			if err != nil {
				fmt.Printf("读取文件目录失败,fullname=%v, err=%v", fullname, err)
				return result, err
			}
			result = append(result, temp...)
		} else {
			result = append(result, fullname)
		}
	}

	return result, nil
}
