package gtype

type Opt interface {
	Tdb() TokenDatabase
	Wsc() SocketChannelCollection
	Appendix() (input, output Appendix)
}
