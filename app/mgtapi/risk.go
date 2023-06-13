package mgtapi

import (
	"net/http"
	"strconv"

	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/vela-ssoc/vela-manager/errcode"
	"github.com/xgfone/ship/v5"
)

func Risk(svc service.RiskService) route.Router {
	riskTypeCol := dynsql.StringColumn("risk_type", "风险类型").Build()
	subjectCol := dynsql.StringColumn("subject", "主题").Build()
	inetCol := dynsql.StringColumn("inet", "终端 IP").Build()
	levelCol := dynsql.StringColumn("level", "级别").Build()
	statusCol := dynsql.StringColumn("status", "状态").Build()
	fromCodeCol := dynsql.StringColumn("from_code", "来源模块").Build()

	table := dynsql.Builder().
		Filters(subjectCol, riskTypeCol, inetCol, levelCol, statusCol, fromCodeCol).
		Groups(subjectCol, riskTypeCol, inetCol, levelCol, statusCol, fromCodeCol).
		Build()
	return &riskREST{
		svc:   svc,
		table: table,
	}
}

type riskREST struct {
	svc   service.RiskService
	table dynsql.Table
}

func (rsk *riskREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/risk/cond").Data(route.Ignore()).GET(rsk.Cond)
	bearer.Route("/risk/attack").Data(route.Ignore()).GET(rsk.Attack)
	bearer.Route("/risk/group").Data(route.Ignore()).GET(rsk.Group)
	bearer.Route("/risk/recent").Data(route.Ignore()).GET(rsk.Recent)
	bearer.Route("/risks").Data(route.Ignore()).GET(rsk.Page)
	bearer.Route("/risk/csv").Data(route.Ignore()).GET(rsk.CSV)
	bearer.Route("/risk/pie").Data(route.Ignore()).GET(rsk.Pie)
}

func (rsk *riskREST) Cond(c *ship.Context) error {
	res := rsk.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rsk *riskREST) Page(c *ship.Context) error {
	var req param.PageSQL
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	scope, err := rsk.table.Inter(req.Input)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}
	page := req.Pager()
	ctx := c.Request().Context()

	count, dats := rsk.svc.Page(ctx, page, scope)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (rsk *riskREST) Attack(c *ship.Context) error {
	var req param.PageSQL
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	scope, err := rsk.table.Inter(req.Input)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}
	page := req.Pager()
	ctx := c.Request().Context()

	count, dats := rsk.svc.Attack(ctx, page, scope)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (rsk *riskREST) Group(c *ship.Context) error {
	var req param.PageSQL
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	if req.Group == "" {
		return errcode.ErrRequiredGroup
	}
	scope, err := rsk.table.Inter(req.Input)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}
	page := req.Pager()
	ctx := c.Request().Context()

	count, dats := rsk.svc.Group(ctx, page, scope)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (rsk *riskREST) Recent(c *ship.Context) error {
	day, _ := strconv.Atoi(c.Query("day"))
	if day > 30 || day < 1 { // 最多支持30天内查询，参数错误或超过有效范围默认为7天
		day = 7
	}

	ctx := c.Request().Context()
	res := rsk.svc.Recent(ctx, day)

	return c.JSON(http.StatusOK, res)
}

func (rsk *riskREST) CSV(c *ship.Context) error {
	return c.JSON(http.StatusOK, nil)
}

func (rsk *riskREST) Pie(c *ship.Context) error {
	group := c.Query("group")
	rtype := c.Query("risk_type")
	topN, _ := strconv.Atoi(c.Query("topn"))
	if topN <= 0 || topN >= 100 {
		topN = 10
	}

	ctx := c.Request().Context()
	tx := query.Risk.WithContext(ctx).UnderlyingDB()
	if rtype != "" {
		tx.Where("risk_type = ?", rtype)
	}

	res := &param.PieTopN{TopN: make([]*param.NameCount, 0, topN)}
	var count int64
	if tx.Count(&count); count == 0 {
		return c.JSON(http.StatusOK, res)
	}
	tx.Select("COUNT(*) count", group+" name").
		Group(group).
		Order("count DESC").
		Limit(topN).
		Scan(&res.TopN)

	var num int
	for _, tn := range res.TopN {
		num += tn.Count
	}
	res.Other = int(count) - num

	return c.JSON(http.StatusOK, res)
}
