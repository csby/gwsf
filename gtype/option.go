package gtype

type Option interface {
	SetTokenChecker(v HttpHandle)
	SetCluster(v bool)
	SetCloud(v bool)
	SetNode(v bool)
}
