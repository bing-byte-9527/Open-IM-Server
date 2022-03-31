package register

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	http2 "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	rpc "Open_IM/pkg/proto/friend"
	"Open_IM/pkg/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ParamsLogin struct {
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	Password    string `json:"password"`
	Platform    int32  `json:"platform"`
	OperationID string `json:"operationID" binding:"required"`
}

func Login(c *gin.Context) {
	params := ParamsLogin{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}
	var account string
	if params.Email != "" {
		account = params.Email
	} else {
		account = params.PhoneNumber
	}

	r, err := im_mysql_model.GetRegister(account)
	if err != nil {
		log.NewError(params.OperationID, "user have not register", params.Password, account, err.Error())
		c.JSON(http.StatusOK, gin.H{"errCode": constant.NotRegistered, "errMsg": "Mobile phone number is not registered"})
		return
	}
	if r.Password != params.Password {
		log.NewError(params.OperationID, "password err", params.Password, account, r.Password, r.Account)
		c.JSON(http.StatusOK, gin.H{"errCode": constant.PasswordErr, "errMsg": "password err"})
		return
	}
	//登录成功 获取userid
	userAccountid := r.Account
	//---------添加默认好友开始--------------------------------------------
	fmt.Println("-------------------------userid:" + userAccountid)
	params1 := api.ImportFriendReq{}
	params1.FriendUserIDList[0] = userAccountid
	params1.FromUserID = "88888888" //默认好友id
	params1.OperationID = params.OperationID
	reqf := &rpc.ImportFriendReq{}
	utils.CopyStructFields(reqf, &params1)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImFriendName)
	client := rpc.NewFriendClient(etcdConn)

	RpcResp, err := client.ImportFriend(context.Background(), reqf)
	if err != nil {
		log.NewError(reqf.OperationID, "ImportFriend failed ", err.Error(), reqf.String())
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": "ImportFriend failed "})
		return
	}
	//--------------------------添加默认好友结束---------------------------------------------------

	//-------------------添加默认群开始---------------------------11

	paramsInviteUser := api.InviteUserToGroupReq{}
	paramsInviteUser.GroupID = "cef3553ebf11886f3b07bc42c683b044"
	paramsInviteUser.InvitedUserIDList[0] = userAccountid
	paramsInviteUser.OperationID = params.OperationID
	paramsInviteUser.Reason = "没有原因"

	reqToGroup := &rpc.InviteUserToGroupReq{}
	utils.CopyStructFields(reqToGroup, &params)
	log.NewInfo(reqToGroup.OperationID, "InviteUserToGroup args ", reqToGroup.String())

	client = rpc.NewGroupClient(etcdConn)
	RpcResp, err = client.InviteUserToGroup(context.Background(), reqToGroup)
	if err != nil {
		log.NewError(reqToGroup.OperationID, "InviteUserToGroup failed ", err.Error(), reqToGroup.String())
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": err.Error()})
		return
	}
	//-------------------添加默认群结束---------------------------
	url := fmt.Sprintf("http://%s:10000/auth/user_token", utils.ServerIP)
	openIMGetUserToken := api.UserTokenReq{}
	openIMGetUserToken.OperationID = params.OperationID
	openIMGetUserToken.Platform = params.Platform
	openIMGetUserToken.Secret = config.Config.Secret
	openIMGetUserToken.UserID = account
	openIMGetUserTokenResp := api.UserTokenResp{}
	bMsg, err := http2.Post(url, openIMGetUserToken, 2)
	if err != nil {
		log.NewError(params.OperationID, "request openIM get user token error", account, "err", err.Error())
		c.JSON(http.StatusOK, gin.H{"errCode": constant.GetIMTokenErr, "errMsg": err.Error()})
		return
	}
	err = json.Unmarshal(bMsg, &openIMGetUserTokenResp)
	if err != nil || openIMGetUserTokenResp.ErrCode != 0 {
		log.NewError(params.OperationID, "request get user token", account, "err", "")
		c.JSON(http.StatusOK, gin.H{"errCode": constant.GetIMTokenErr, "errMsg": ""})
		return
	}
	c.JSON(http.StatusOK, gin.H{"errCode": constant.NoError, "errMsg": "", "data": openIMGetUserTokenResp.UserToken})

}
