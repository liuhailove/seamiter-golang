package gray

import (
	"fmt"
	"github.com/go-basic/uuid"
	"math/rand"
	"testing"
	"time"
)

func TestNewWeightTrafficSelector(t *testing.T) {
	var gWeight1 = GWeight{
		EffectiveAddresses: "",
		TargetResource:     "accountService.AccountService.QueryV1",
		TargetVersion:      "",
		Weight:             1,
	}

	var gWeight2 = GWeight{
		EffectiveAddresses: "",
		TargetResource:     "accountService.AccountService.QueryV2",
		TargetVersion:      "",
		Weight:             1,
	}

	var gWeight3 = GWeight{
		EffectiveAddresses: "",
		TargetResource:     "accountService.AccountService.QueryV3",
		TargetVersion:      "",
		Weight:             0,
	}

	var gWeight4 = GWeight{
		EffectiveAddresses: "",
		TargetResource:     "accountService.AccountService.QueryV4",
		TargetVersion:      "",
		Weight:             1,
	}

	var weights = []GWeight{gWeight1, gWeight2, gWeight3, gWeight4}

	var rule = new(Rule)
	rule.ID = uuid.New()
	rule.Resource = "accountService.AccountService.Query"
	rule.GrayTag = "grayTag"
	rule.LinkPass = true
	rule.RouterStrategy = WeightRouter
	rule.Force = true
	rule.GrayWeightList = weights

	var countMap = make(map[string]int)
	var tsc = &TrafficSelectorController{rule: rule}
	var ts = NewWeightTrafficSelector(tsc, rule)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 100; i++ {
		//time.Sleep(time.Duration(rand.Intn(10000)))
		var res, _ = ts.CalculateAllowedResource(nil)
		var cnt = countMap[res]
		cnt += 1
		countMap[res] = cnt
		fmt.Println(res)
	}

	for k, v := range countMap {
		fmt.Println(fmt.Sprintf("res=%s,cnt=%d", k, v))
	}
}

func TestIsValidRule(t *testing.T) {
	var dmap map[string]string
	fmt.Println(dmap["abc"])
	fmt.Println([]byte(dmap["abc"]))
}
