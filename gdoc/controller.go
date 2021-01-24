package gdoc

import (
	"github.com/csby/gwsf/gtype"
)

type controller struct {
	doc  gtype.Doc
	info gtype.ServerInfo
}

func (s *controller) GetInformation(ctx gtype.Context, ps gtype.Params) {
	ctx.Success(&s.info)
}

func (s *controller) GetCatalogTree(ctx gtype.Context, ps gtype.Params) {
	if s.doc == nil {
		ctx.Error(gtype.ErrInternal, "doc is nil")
		return
	}

	ctx.Success(s.doc.Catalogs())
}

func (s *controller) GetFunctionDetail(ctx gtype.Context, ps gtype.Params) {
	if s.doc == nil {
		ctx.Error(gtype.ErrInternal, "doc is nil")
		return
	}

	id := ps.ByName("id")
	fun, err := s.doc.Function(id, ctx.Schema(), ctx.Host())
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}

	ctx.Success(fun)
}

func (s *controller) GetTokenUI(ctx gtype.Context, ps gtype.Params) {
	if s.doc == nil {
		ctx.Error(gtype.ErrInternal, "doc is nil")
		return
	}

	id := ps.ByName("id")
	ui, err := s.doc.TokenUI(id)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}

	ctx.Success(ui)
}

func (s *controller) CreateToken(ctx gtype.Context, ps gtype.Params) {
	if s.doc == nil {
		ctx.Error(gtype.ErrInternal, "doc is nil")
		return
	}

	items := make([]gtype.TokenAuth, 0)
	err := ctx.GetJson(&items)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}

	id := ps.ByName("id")
	token, code := s.doc.TokenCreate(id, items, ctx)
	if code != nil {
		ctx.Error(code)
		return
	}

	ctx.Success(token)
}
