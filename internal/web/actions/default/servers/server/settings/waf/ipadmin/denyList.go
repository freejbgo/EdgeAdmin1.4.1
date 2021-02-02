package ipadmin

import (
	"github.com/TeaOSLab/EdgeAdmin/internal/web/actions/actionutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/dao"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/maps"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

type DenyListAction struct {
	actionutils.ParentAction
}

func (this *DenyListAction) Init() {
	this.Nav("", "setting", "denyList")
	this.SecondMenu("waf")
}

func (this *DenyListAction) RunGet(params struct {
	FirewallPolicyId int64
	ServerId         int64
}) {
	this.Data["featureIsOn"] = true
	this.Data["firewallPolicyId"] = params.FirewallPolicyId

	listId, err := dao.SharedIPListDAO.FindDenyIPListIdWithServerId(this.AdminContext(), params.ServerId)
	if err != nil {
		this.ErrorPage(err)
		return
	}

	// 创建
	if listId == 0 {
		listId, err = dao.SharedIPListDAO.CreateIPListForServerId(this.AdminContext(), params.ServerId, "black")
		if err != nil {
			this.ErrorPage(err)
			return
		}
	}

	this.Data["listId"] = listId

	// 数量
	countResp, err := this.RPC().IPItemRPC().CountIPItemsWithListId(this.AdminContext(), &pb.CountIPItemsWithListIdRequest{IpListId: listId})
	if err != nil {
		this.ErrorPage(err)
		return
	}
	count := countResp.Count
	page := this.NewPage(count)
	this.Data["page"] = page.AsHTML()

	// 列表
	itemsResp, err := this.RPC().IPItemRPC().ListIPItemsWithListId(this.AdminContext(), &pb.ListIPItemsWithListIdRequest{
		IpListId: listId,
		Offset:   page.Offset,
		Size:     page.Size,
	})
	if err != nil {
		this.ErrorPage(err)
		return
	}
	itemMaps := []maps.Map{}
	for _, item := range itemsResp.IpItems {
		expiredTime := ""
		if item.ExpiredAt > 0 {
			expiredTime = timeutil.FormatTime("Y-m-d H:i:s", item.ExpiredAt)
		}

		itemMaps = append(itemMaps, maps.Map{
			"id":          item.Id,
			"ipFrom":      item.IpFrom,
			"ipTo":        item.IpTo,
			"expiredTime": expiredTime,
			"reason":      item.Reason,
			"type":        item.Type,
			"isExpired":   item.ExpiredAt > 0 && item.ExpiredAt < time.Now().Unix(),
		})
	}
	this.Data["items"] = itemMaps

	// WAF是否启用
	webConfig, err := dao.SharedHTTPWebDAO.FindWebConfigWithServerId(this.AdminContext(), params.ServerId)
	if err != nil {
		this.ErrorPage(err)
		return
	}
	this.Data["wafIsOn"] = webConfig.FirewallRef != nil && webConfig.FirewallRef.IsOn

	this.Show()
}
