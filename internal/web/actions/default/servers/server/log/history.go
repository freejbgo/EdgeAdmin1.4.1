package log

import (
	"github.com/TeaOSLab/EdgeAdmin/internal/web/actions/actionutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/lists"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"regexp"
	"strings"
)

type HistoryAction struct {
	actionutils.ParentAction
}

func (this *HistoryAction) Init() {
	this.Nav("", "log", "")
	this.SecondMenu("history")
}

func (this *HistoryAction) RunGet(params struct {
	ServerId int64
	Day      string
	Hour     string
	Keyword  string
	Ip       string
	Domain   string
	HasWAF   int

	RequestId string
	HasError  int

	ClusterId int64
	NodeId    int64

	Partition int32 `default:"-1"`

	PageSize int
}) {
	if len(params.Day) == 0 {
		params.Day = timeutil.Format("Y-m-d")
	}

	this.Data["path"] = this.Request.URL.Path
	this.Data["day"] = params.Day
	this.Data["hour"] = params.Hour
	this.Data["keyword"] = params.Keyword
	this.Data["ip"] = params.Ip
	this.Data["domain"] = params.Domain
	this.Data["accessLogs"] = []interface{}{}
	this.Data["hasError"] = params.HasError
	this.Data["hasWAF"] = params.HasWAF
	this.Data["pageSize"] = params.PageSize
	this.Data["clusterId"] = params.ClusterId
	this.Data["nodeId"] = params.NodeId
	this.Data["partition"] = params.Partition

	day := params.Day
	ipList := []string{}

	if len(day) > 0 && regexp.MustCompile(`\d{4}-\d{2}-\d{2}`).MatchString(day) {
		day = strings.ReplaceAll(day, "-", "")
		size := int64(params.PageSize)
		if size < 1 {
			size = 20
		}

		this.Data["hasError"] = params.HasError

		resp, err := this.RPC().HTTPAccessLogRPC().ListHTTPAccessLogs(this.AdminContext(), &pb.ListHTTPAccessLogsRequest{
			Partition:         params.Partition,
			RequestId:         params.RequestId,
			ServerId:          params.ServerId,
			HasError:          params.HasError > 0,
			HasFirewallPolicy: params.HasWAF > 0,
			Day:               day,
			HourFrom:          params.Hour,
			HourTo:            params.Hour,
			Keyword:           params.Keyword,
			Ip:                params.Ip,
			Domain:            params.Domain,
			NodeId:            params.NodeId,
			NodeClusterId:     params.ClusterId,
			Size:              size,
		})
		if err != nil {
			this.ErrorPage(err)
			return
		}

		if len(resp.HttpAccessLogs) == 0 {
			this.Data["accessLogs"] = []interface{}{}
		} else {
			this.Data["accessLogs"] = resp.HttpAccessLogs
			for _, accessLog := range resp.HttpAccessLogs {
				if len(accessLog.RemoteAddr) > 0 {
					if !lists.ContainsString(ipList, accessLog.RemoteAddr) {
						ipList = append(ipList, accessLog.RemoteAddr)
					}
				}
			}
		}
		this.Data["hasMore"] = resp.HasMore
		this.Data["nextRequestId"] = resp.RequestId

		// 上一个requestId
		this.Data["hasPrev"] = false
		this.Data["lastRequestId"] = ""
		if len(params.RequestId) > 0 {
			this.Data["hasPrev"] = true
			prevResp, err := this.RPC().HTTPAccessLogRPC().ListHTTPAccessLogs(this.AdminContext(), &pb.ListHTTPAccessLogsRequest{
				Partition:         params.Partition,
				RequestId:         params.RequestId,
				ServerId:          params.ServerId,
				HasError:          params.HasError > 0,
				HasFirewallPolicy: params.HasWAF > 0,
				Day:               day,
				Keyword:           params.Keyword,
				Ip:                params.Ip,
				Domain:            params.Domain,
				NodeId:            params.NodeId,
				NodeClusterId:     params.ClusterId,
				Size:              size,
				Reverse:           true,
			})
			if err != nil {
				this.ErrorPage(err)
				return
			}
			if int64(len(prevResp.HttpAccessLogs)) == size {
				this.Data["lastRequestId"] = prevResp.RequestId
			}
		}
	}

	// 根据IP查询区域
	regionMap := map[string]string{} // ip => region
	if len(ipList) > 0 {
		resp, err := this.RPC().IPLibraryRPC().LookupIPRegions(this.AdminContext(), &pb.LookupIPRegionsRequest{IpList: ipList})
		if err != nil {
			this.ErrorPage(err)
			return
		}
		if resp.IpRegionMap != nil {
			for ip, region := range resp.IpRegionMap {
				regionMap[ip] = region.Summary
			}
		}
	}
	this.Data["regions"] = regionMap

	this.Show()
}
