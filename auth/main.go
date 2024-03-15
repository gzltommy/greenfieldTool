package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	permTypes "github.com/bnb-chain/greenfield/x/permission/types"
	cosmosSdk "github.com/cosmos/cosmos-sdk/types"
	exterrors "github.com/pkg/errors"
)

func main() {
	var (
		privateKey    string
		bucketName    string
		authedAddress string
	)

	// StringVar用指定的名称、控制台参数项目、默认值、使用信息注册一个string类型flag，并将flag的值保存到p指向的变量
	flag.StringVar(&privateKey, "p", "", "主账号私钥")
	flag.StringVar(&bucketName, "b", "", "存储桶名字")
	flag.StringVar(&authedAddress, "a", "", "被授权的钱包地址")

	// 从arguments中解析注册的flag。必须在所有flag都注册好而未访问其值时执行。未注册却使用flag -help时，会返回ErrHelp。
	flag.Parse()

	// 打印
	fmt.Printf("privateKey=%v bucketName=%v authedAddress=%v \n", privateKey, bucketName, authedAddress)

	if privateKey == "" || bucketName == "" || authedAddress == "" {
		fmt.Println("无效的参数！")
		return
	}

	_, err := BucketAuth(privateKey, bucketName, authedAddress, []permTypes.ActionType{
		permTypes.ACTION_TYPE_ALL,
	})
	if err != nil {
		fmt.Printf("authorization failed! error:%v", err)
		return
	}

	fmt.Println("authorization ok!! ")
}

const (
	// 正式网
	rpcAddr = "https://greenfield-chain.bnbchain.org:443"
	chainId = "greenfield_1017-1"

	// 测试网
	//rpcAddr = "https://gnfd-testnet-fullnode-tendermint-us.bnbchain.org:443"
	//chainId = "greenfield_5600-1"
)

// BucketAuth 将桶的操作权限授予某个地址
// 注意，被授予权限的地址必须先在链上存在（发生过交易？也建一个桶？）
func BucketAuth(privateKey, bucketName, principal string, permission []permTypes.ActionType) (*permTypes.Policy, error) {
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
