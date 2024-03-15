/*
replace (
	cosmossdk.io/api => github.com/bnb-chain/greenfield-cosmos-sdk/api v0.0.0-20230816082903-b48770f5e210
	cosmossdk.io/math => github.com/bnb-chain/greenfield-cosmos-sdk/math v0.0.0-20230816082903-b48770f5e210
	github.com/cometbft/cometbft => github.com/bnb-chain/greenfield-cometbft v1.1.0
	github.com/cometbft/cometbft-db => github.com/bnb-chain/greenfield-cometbft-db v0.8.1-alpha.1
	github.com/consensys/gnark-crypto => github.com/consensys/gnark-crypto v0.7.0
	github.com/cosmos/cosmos-sdk => github.com/bnb-chain/greenfield-cosmos-sdk v1.1.0
	github.com/cosmos/iavl => github.com/bnb-chain/greenfield-iavl v0.20.1
	github.com/ferranbt/fastssz => github.com/ferranbt/fastssz v0.0.0-20210905181407-59cf6761a7d5
	github.com/syndtr/goleveldb => github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
	golang.org/x/exp => golang.org/x/exp v0.0.0-20230515195305-f3d0a9c9a5cc
	google.golang.org/protobuf => google.golang.org/protobuf v1.31.0
)

go mod edit -require github.com/consensys/gnark-crypto@v0.7.0
go mod edit -require golang.org/x/exp@v0.0.0-20230515195305-f3d0a9c9a5cc

*/

// BucketName:"defaution"
// BucketSPEndpoint:https://gnfd-testnet-sp1.bnbchain.org
// 浏览器的访问路径：BucketSPEndpoint +  "/view" + "/objectpath"；如
// https://gnfd-testnet-sp1.bnbchain.org/view/defution1%2Fnft%2Fbb%2Fb4c25916c1210a7d660b949b75ce37ad.png

package greenfield

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	permTypes "github.com/bnb-chain/greenfield/x/permission/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	cosmosSdk "github.com/cosmos/cosmos-sdk/types"
	exterrors "github.com/pkg/errors"
	"strings"
	"time"
)

type Client struct {
	chainId    string
	rpcAddr    string
	privateKey string
	bucketName string
	cli        client.IClient
}

func NewClient(chainId, rpcAddr, privateKey, bucketName string) (*Client, error) {
	account, err := types.NewAccountFromPrivateKey("test", privateKey)
	if err != nil {
		return nil, exterrors.WithStack(err)
	}
	cli, err := client.New(chainId, rpcAddr, client.Option{DefaultAccount: account})
	if err != nil {
		return nil, exterrors.WithStack(err)
	}

	return &Client{
		chainId:    chainId,
		rpcAddr:    rpcAddr,
		privateKey: privateKey,
		bucketName: bucketName,
		cli:        cli,
	}, nil
}

func (c *Client) GetBalance(address string) (string, error) {
	balance, err := c.cli.GetAccountBalance(context.Background(), address)
	if err != nil {
		return "", exterrors.WithStack(err)
	}
	return balance.Amount.String(), nil
}

func (c *Client) Storage(objectName, contentType string, objectContent []byte) (*types.ObjectDetail, error) {
	var (
		ctx      = context.Background()
		obDetail *types.ObjectDetail
	)

	// 1.create object
	txnHash, err := c.cli.CreateObject(ctx,
		c.bucketName,
		objectName,
		bytes.NewReader(objectContent),
		types.CreateObjectOptions{
			Visibility:  storageTypes.VISIBILITY_TYPE_PUBLIC_READ,
			ContentType: contentType,
		})
	if err != nil {
		// 已经存在，获取头
		if strings.Contains(err.Error(), "repeated object") {
			obDetail, err = c.cli.HeadObject(ctx, c.bucketName, objectName)
			if err == nil {
				return obDetail, nil
			}
		}
		return nil, exterrors.WithStack(err)
	}

	// 2.put object
	err = c.cli.PutObject(ctx,
		c.bucketName,
		objectName,
		int64(len(objectContent)),
		bytes.NewReader(objectContent), types.PutObjectOptions{
			TxnHash:     txnHash,
			ContentType: contentType,
		})
	if err != nil {
		return nil, exterrors.WithStack(err)
	}

	t1 := time.Now().Unix()

	// 3.等待 Seal
	obDetail, err = c.waitObjectSeal(objectName)
	if err != nil {
		return nil, exterrors.WithStack(err)
	}
	fmt.Println("-----------", time.Now().Unix()-t1)

	return obDetail, nil
}

// waitObjectSeal wait for the object to be sealed
func (c *Client) waitObjectSeal(objectName string) (*types.ObjectDetail, error) {
	ctx := context.Background()
	timeout := time.After(20 * time.Second)
	ticker := time.NewTicker(2 * time.Second)

	for {
		select {
		case <-timeout:
			err := exterrors.New("object not sealed after 20 seconds")
			if err != nil {
				return nil, exterrors.WithStack(err)
			}
		case <-ticker.C:
			objDetail, err := c.cli.HeadObject(ctx, c.bucketName, objectName)
			if err != nil {
				return nil, exterrors.WithStack(err)
			}
			if objDetail.ObjectInfo.GetObjectStatus() == storageTypes.OBJECT_STATUS_SEALED {
				ticker.Stop()
				return objDetail, nil
			}
		}
	}
}

// BucketAuth 将桶的操作权限授予某个地址
// 注意，被授予权限的地址必须先在链上存在（发生过交易？也建一个桶？）
func BucketAuth(chainId, rpcAddr, privateKey, bucketName, principal string, permission []permTypes.ActionType) (*permTypes.Policy, error) {
	// you need to set the principal address in config.go to run this examples
	if len(principal) < 42 {
		return nil, exterrors.New("please set principal if you need run permission test")
	}
	granteeAddr, err := cosmosSdk.AccAddressFromHexUnsafe(principal)
	if err != nil {
		return nil, exterrors.Wrapf(err, "principal addr invalid")
	}

	principalStr, err := utils.NewPrincipalWithAccount(granteeAddr)
	if err != nil {
		return nil, exterrors.WithStack(err)
	}

	account, err := types.NewAccountFromPrivateKey("test", privateKey)
	if err != nil {
		return nil, exterrors.WithStack(err)
	}
	cli, err := client.New(chainId, rpcAddr, client.Option{DefaultAccount: account})
	if err != nil {
		return nil, exterrors.WithStack(err)
	}

	// put bucket policy
	ctx := context.Background()
	statements := utils.NewStatement(permission, permTypes.EFFECT_ALLOW, nil, types.NewStatementOptions{})
	policyTx, err := cli.PutBucketPolicy(ctx, bucketName, principalStr, []*permTypes.Statement{&statements},
		types.PutPolicyOption{})
	if err != nil {
		return nil, exterrors.WithStack(err)
	}
	_, err = cli.WaitForTx(ctx, policyTx)
	if err != nil {
		return nil, exterrors.WithStack(err)
	}

	// get bucket policy
	policyInfo, err := cli.GetBucketPolicy(ctx, bucketName, principal)
	if err != nil {
		return nil, exterrors.WithStack(err)
	}
	return policyInfo, nil
}
