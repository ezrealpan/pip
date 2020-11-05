package input

import "ezreal.com.cn/pip/pip"

type Creator func() pip.Input

var Inputs = map[string]Creator{}

func Add(name string, creator Creator) {
	Inputs[name] = creator
}
