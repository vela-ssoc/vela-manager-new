package mgtapi

import (
	"context"
	"net/http"

	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/vela-ssoc/vela-manager/app/session"
	"github.com/xgfone/ship/v5"
)

func Substance(svc service.SubstanceService) route.Router {
	return &substanceREST{
		svc: svc,
	}
}

type substanceREST struct {
	svc service.SubstanceService
}

func (rest *substanceREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/substances").GET(rest.Page)
	bearer.Route("/substance/indices").GET(rest.Indices)
	bearer.Route("/substance").
		GET(rest.Detail).
		POST(rest.Create).
		PATCH(rest.Update)
}

func (rest *substanceREST) Indices(c *ship.Context) error {
	var req param.Index
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	idx := req.Indexer()
	ctx := c.Request().Context()
	dats := rest.svc.Indices(ctx, idx)

	return c.JSON(http.StatusOK, dats)
}

func (rest *substanceREST) Page(c *ship.Context) error {
	var req param.Page
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	page := req.Pager()
	ctx := c.Request().Context()
	count, dats := rest.svc.Page(ctx, page)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (rest *substanceREST) Detail(c *ship.Context) error {
	var req param.IntID
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	id := req.ID
	res, err := rest.svc.Detail(ctx, id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (rest *substanceREST) Create(c *ship.Context) error {
	var req param.SubstanceCreate
	if err := c.Bind(&req); err != nil {
		return err
	}

	cu := session.Cast(c.Any)
	ctx := c.Request().Context()

	return rest.svc.Create(ctx, &req, cu.ID)
}

func (rest *substanceREST) Update(c *ship.Context) error {
	var req param.Page
	if err := c.Bind(&req); err != nil {
		return err
	}

	// 62594160710045696
	rest.svc.Update(context.Background(), &param.SubstanceUpdate{ID: 62594160710045696}, 62594160710045696)

	return nil
}