package gtype

type Uri interface {
	Path() string
	IsWebsocket() bool
	SetIsWebsocket(isWebsocket bool) Uri
	TokenPlace() int
	SetTokenPlace(place int) Uri
	TokenUI() func() []TokenUI
	SetTokenUI(ui func() []TokenUI) Uri
	TokenCreate() func(items []TokenAuth, ctx Context) (string, Error)
	SetTokenCreate(create func(items []TokenAuth, ctx Context) (string, Error)) Uri
}
