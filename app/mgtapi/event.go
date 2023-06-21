package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/vela-ssoc/vela-manager/errcode"
	"github.com/xgfone/ship/v5"
)

func Event(svc service.EventService) route.Router {
	levels := []string{
		model.ELvlCritical.String(),
		model.ELvlMajor.String(),
		model.ELvlMinor.String(),
		model.ELvlNote.String(),
	}
	levelEnums := dynsql.StringEnum().Sames(levels)
	filters := []dynsql.Column{
		dynsql.StringColumn("level", "级别").Enums(levelEnums).Build(),
		dynsql.StringColumn("inet", "终端 IP").Build(),
		dynsql.IntColumn("minion_id", "终端 ID").Build(),
		dynsql.StringColumn("from_code", "来源").Build(),
		dynsql.TimeColumn("occur_at", "时间").Build(),
		dynsql.StringColumn("subject", "主题").Build(),
		dynsql.StringColumn("typeof", "类型").Build(),
		dynsql.StringColumn("remote_addr", "远端地址").Build(),
		dynsql.IntColumn("remote_port", "远端端口").Build(),
		dynsql.StringColumn("user", "用户").Build(),
		dynsql.StringColumn("auth", "授权信息").Build(),
		dynsql.StringColumn("msg", "信息").Build(),
		dynsql.StringColumn("error", "错误信息").Build(),
		dynsql.StringColumn("region", "IP归属地").Build(),
		dynsql.BoolColumn("send_alert", "是否发送告警").
			Enums(dynsql.BoolEnum().True("是").False("否")).Build(),
		dynsql.TimeColumn("created_at", "创建时间").Build(),
		dynsql.IntColumn("minion_id", "终端ID").Build(),
	}

	table := dynsql.Builder().
		Filters(filters...).
		Build()
	return &eventREST{
		svc:   svc,
		table: table,
	}
}

type eventREST struct {
	svc   service.EventService
	table dynsql.Table
}

func (evt *eventREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/event/cond").Data(route.Ignore()).GET(evt.Cond)
	bearer.Route("/events").Data(route.Ignore()).GET(evt.Page)
	bearer.Route("/event/confirm").
		Data(route.Named("忽略事件")).DELETE(evt.Confirm)
	bearer.Route("/event").
		Data(route.Named("批量删除事件")).DELETE(evt.Delete)
}

func (evt *eventREST) Cond(c *ship.Context) error {
	res := evt.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (evt *eventREST) Page(c *ship.Context) error {
	var req param.PageSQL
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	scope, err := evt.table.Inter(req.Input)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}
	page := req.Pager()
	ctx := c.Request().Context()

	count, dats := evt.svc.Page(ctx, page, scope)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (evt *eventREST) Confirm(c *ship.Context) error {
	var req param.OptionalIDs
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return evt.svc.Confirm(ctx, req.ID)
}

func (evt *eventREST) Delete(c *ship.Context) error {
	var req dynsql.Input
	if err := c.Bind(&req); err != nil {
		return err
	}
	if len(req.Filters) == 0 {
		return errcode.ErrRequiredFilter
	}
	scope, err := evt.table.Inter(req)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()

	return evt.svc.Delete(ctx, scope)
}
