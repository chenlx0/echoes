package balance

import (
	"testing"

	"github.com/chenlx0/echoes/internal/config"
)

func getTestData() map[string][]*config.Upstream {
	confs := make(map[string][]*config.Upstream, 3)
	confs["a"] = []*config.Upstream{
		&config.Upstream{
			Host:   "test1",
			Weight: 4,
		},
		&config.Upstream{
			Host:   "test2",
			Weight: 3,
		},
		&config.Upstream{
			Host:   "test3",
			Weight: 1,
		},
	}
	return confs
}

func Test_RR_1(t *testing.T) {
	confs := getTestData()
	rr := GetRoundRobin(confs)
	res := map[string]int{
		"test1": 0,
		"test2": 0,
		"test3": 0,
	}

	for i := 0; i < 800; i++ {
		res[rr("a").Host]++
	}

	if res["test1"] != 400 || res["test2"] != 300 || res["test3"] != 100 {
		t.Errorf("Bad rr result: %v", res)
	}

}
