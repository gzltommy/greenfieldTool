package greenfield

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	exterrors "github.com/pkg/errors"
	"github.com/zorrotommy/greenfieldTool/file_type"
	"sync"
)

const (
	// 正式网
	ChainId = "greenfield_1017-1"
	RpcAddr = "https://greenfield-chain.bnbchain.org:443"

	// 测试网
	TestChainId = "greenfield_5600-1"
	TestRpcAddr = "https://gnfd-testnet-fullnode-tendermint-us.bnbchain.org:443"
)

var (
	clientPool *ClientPool
	once       sync.Once
)

type ClientPool struct {
	clis chan *Client
}

func Init(privateKeys []string, bucketName string, testGreenfield bool) error {
	if len(privateKeys) == 0 {
		return exterrors.New("privateKeys is empty")
	}
	chainId, rpcAddr := ChainId, RpcAddr
	if testGreenfield {
		chainId, rpcAddr = TestChainId, TestRpcAddr
	}

	var err error
	once.Do(func() {
		clientPool = &ClientPool{
			clis: make(chan *Client, len(privateKeys)+1),
		}
		for _, pk := range privateKeys {
			var cli *Client
			cli, err = NewClient(chainId, rpcAddr, pk, bucketName)
			if err != nil {
				return
			}
			clientPool.clis <- cli
		}
	})
	return err
}

// Get 从池子获取一个
func (cp *ClientPool) Get() *Client {
	cli := <-cp.clis
	return cli
}

// Put 放回客户端池子
func (cp *ClientPool) Put(cli *Client) {
	cp.clis <- cli
}

func UploadResource(fileData []byte, ext, dir string) (string, error) {
	cli := clientPool.Get()
	defer clientPool.Put(cli)
	sh := md5.New()
	sh.Write(fileData)
	fileHash := hex.EncodeToString(sh.Sum([]byte("")))
	if dir == "" {
		dir = "resource/common"
	}
	objectName := fmt.Sprintf("%s/%s%s", dir, fileHash, ext)
	contentType := file_type.DetectFileType(fileData)
	_, err := cli.Storage(objectName, contentType, fileData)
	if err != nil {
		return "", err
	}
	return objectName, nil
}

func UploadResourceWithFilePath(fileData []byte, filePath string) (string, error) {
	cli := clientPool.Get()
	defer clientPool.Put(cli)
	contentType := file_type.DetectFileType(fileData)
	_, err := cli.Storage(filePath, contentType, fileData)
	if err != nil {
		return "", err
	}
	return filePath, nil
}
