package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func MinionTask(svc service.MinionTaskService) route.Router {
	statusEnums := dynsql.StringEnum().
		Set("running", "正常运行").
		Set("doing", "正在启动").
		Set("fail", "运行失败").
		Set("panic", "运行崩溃").
		Set("reg", "注册状态").
		Set("update", "正在更新")
	dialectEnums := dynsql.BoolEnum().True("私有").False("公有")
	filters := []dynsql.Column{
		dynsql.StringColumn("inet", "终端IP").Build(),
		dynsql.StringColumn("name", "名称").Build(),
		dynsql.BoolColumn("dialect", "属性").Enums(dialectEnums).Build(),
		dynsql.StringColumn("status", "状态").Enums(statusEnums).Build(),
		dynsql.IntColumn("minion_id", "节点ID").Build(),
		dynsql.IntColumn("substance_id", "配置ID").Build(),
	}

	table := dynsql.Builder().Filters(filters...).Build()

	return &minionTaskREST{
		svc:   svc,
		table: table,
	}
}

type minionTaskREST struct {
	svc   service.MinionTaskService
	table dynsql.Table
}

func (rest *minionTaskREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/tasks").GET(rest.Page)
	bearer.Route("/task").GET(rest.Detail)
	bearer.Route("/task/cond").GET(rest.Cond)
	bearer.Route("/task/minion").GET(rest.Minion)
	bearer.Route("/task/gathers").GET(rest.Gathers)
	bearer.Route("/task/count").GET(rest.Count)
}

func (rest *minionTaskREST) Cond(c *ship.Context) error {
	res := rest.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *minionTaskREST) Page(c *ship.Context) error {
	var req param.PageSQL
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	scope, err := rest.table.Inter(req.Input)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}

	page := req.Pager()
	ctx := c.Request().Context()
	count, dats := rest.svc.Page(ctx, page, scope)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (rest *minionTaskREST) Detail(c *ship.Context) error {
	var req param.MinionTaskDetailRequest
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	res, err := rest.svc.Detail(ctx, req.ID, req.SubstanceID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (rest *minionTaskREST) Minion(c *ship.Context) error {
	var req param.IntID
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	dats, err := rest.svc.Minion(ctx, req.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dats)
}

func (rest *minionTaskREST) Gathers(c *ship.Context) error {
	var req param.Page
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	page := req.Pager()
	count, dats := rest.svc.Gather(ctx, page)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (rest *minionTaskREST) Count(c *ship.Context) error {
	ctx := c.Request().Context()
	res := rest.svc.Count(ctx)
	return c.JSON(http.StatusOK, res)
}
