package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func SBOMComponent(svc service.SBOMComponentService) route.Router {
	inetCol := dynsql.StringColumn("inet", "终端IP").Build()
	nameCol := dynsql.StringColumn("name", "组件名").Build()
	versionCol := dynsql.StringColumn("version", "版本").Build()
	purlCol := dynsql.StringColumn("purl", "PURL").Build()
	totalNumCol := dynsql.IntColumn("total_num", "漏洞总数").Build()
	pidCol := dynsql.IntColumn("project_id", "文件ID").Build()
	midCol := dynsql.IntColumn("minion_id", "终端ID").Build()
	idCol := dynsql.IntColumn("id", "组件ID").Build()
	table := dynsql.Builder().
		Filters(inetCol, nameCol, versionCol, purlCol, totalNumCol, pidCol, midCol, idCol).
		Build()

	return &sbomComponentREST{
		svc:   svc,
		table: table,
	}
}

type sbomComponentREST struct {
	svc   service.SBOMComponentService
	table dynsql.Table
}

func (rest *sbomComponentREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/sbom/components").GET(rest.Page)
	bearer.Route("/sbom/component/cond").GET(rest.Cond)
}

func (rest *sbomComponentREST) Cond(c *ship.Context) error {
	res := rest.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *sbomComponentREST) Page(c *ship.Context) error {
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
