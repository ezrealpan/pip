package output

import "ezreal.com.cn/pip/pip"

type Creator func() pip.Output

var Outputs = map[string]Creator{}

func Add(name string, creator Creator) {
	Outputs[name] = creator
}
