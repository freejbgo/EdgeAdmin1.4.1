package waf

import (
	"github.com/TeaOSLab/EdgeAdmin/internal/oplogs"
	"github.com/TeaOSLab/EdgeAdmin/internal/web/actions/actionutils"
	"github.com/TeaOSLab/EdgeAdmin/internal/web/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/firewallconfigs"
)

type SortGroupsAction struct {
	actionutils.ParentAction
}

func (this *SortGroupsAction) RunPost(params struct {
	FirewallPolicyId int64
	Type             string
	GroupIds         []int64
}) {
	// 日志
	defer this.CreateLog(oplogs.LevelInfo, "修改WAF策略 %d 中的规则分组中的排序", params.FirewallPolicyId)

	firewallPolicy, err := models.SharedHTTPFirewallPolicyDAO.FindEnabledPolicyConfig(this.AdminContext(), params.FirewallPolicyId)
	if err != nil {
		this.ErrorPage(err)
		return
	}

	if firewallPolicy == nil {
		this.NotFound("firewallPolicy", params.FirewallPolicyId)
		return
	}

	switch params.Type {
	case "inbound":
		refMapping := map[int64]*firewallconfigs.HTTPFirewallRuleGroupRef{}
		for _, ref := range firewallPolicy.Inbound.GroupRefs {
			refMapping[ref.GroupId] = ref
		}
		newRefs := []*firewallconfigs.HTTPFirewallRuleGroupRef{}
		for _, groupId := range params.GroupIds {
			ref, ok := refMapping[groupId]
			if ok {
				newRefs = append(newRefs, ref)
			}
		}
		firewallPolicy.Inbound.GroupRefs = newRefs
	case "outbound":
		refMapping := map[int64]*firewallconfigs.HTTPFirewallRuleGroupRef{}
		for _, ref := range firewallPolicy.Outbound.GroupRefs {
			refMapping[ref.GroupId] = ref
		}
		newRefs := []*firewallconfigs.HTTPFirewallRuleGroupRef{}
		for _, groupId := range params.GroupIds {
			ref, ok := refMapping[groupId]
			if ok {
				newRefs = append(newRefs, ref)
			}
		}
		firewallPolicy.Outbound.GroupRefs = newRefs
	}

	inboundJSON, err := firewallPolicy.InboundJSON()
	if err != nil {
		this.ErrorPage(err)
		return
	}

	outboundJSON, err := firewallPolicy.OutboundJSON()
	if err != nil {
		this.ErrorPage(err)
		return
	}

	_, err = this.RPC().HTTPFirewallPolicyRPC().UpdateHTTPFirewallPolicyGroups(this.AdminContext(), &pb.UpdateHTTPFirewallPolicyGroupsRequest{
		HttpFirewallPolicyId: params.FirewallPolicyId,
		InboundJSON:          inboundJSON,
		OutboundJSON:         outboundJSON,
	})
	if err != nil {
		this.ErrorPage(err)
		return
	}

	this.Success()
}
