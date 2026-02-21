package sg

import "github.com/devthicket/willow"

type TestRunner = willow.TestRunner

func LoadTestScript(data []byte) (*TestRunner, error) {
	return willow.LoadTestScript(data)
}
