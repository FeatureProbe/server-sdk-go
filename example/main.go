package main

import (
	"fmt"

	featureprobe "github.com/featureprobe/server-sdk-go"
)

func main() {
	fp, err := featureprobe.NewFeatureProbe("https://featureprobe.io/server", "bd2f4bf8ec431370d4f9c99b57d33d1f74375962", featureprobe.WithRefreshInterval(5000), featureprobe.WithWaitFirstResp(true))
	if err != nil {
		fmt.Println(err)
		return
	}
	user := featureprobe.NewUser("testUSerKey").With("userId", "00001")

	detail := fp.BoolDetail("campaign_allow_list", user, false)
	fmt.Println("Result =>", detail.Value)
	fmt.Println("       => reason:", detail.Reason)
	fmt.Println("       => rule index:", detail.RuleIndex)

	fp.Close()
}
