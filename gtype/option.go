package gtype

type Option interface {
	SetTokenChecker(v HttpHandle)
	SetCloud(v bool)
	SetNode(v bool)
}
