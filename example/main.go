package main

import (
	"fmt"

	featureprobe "github.com/featureprobe/server-sdk-go"
)

func main() {

	config := featureprobe.FPConfig{
		RemoteUrl: "https://featureprobe.io/server",
		// RemoteUrl:       "http://127.0.0.1.4007", // for local docker
		ServerSdkKey:    "server-bd2f4bf8ec431370d4f9c99b57d33d1f74375962",
		RefreshInterval: 2000, // ms
		StartWait:       5000,
	}
	fp, err := featureprobe.NewFeatureProbe(config)
	if fp.Initialized() {
		fmt.Println("SDK successfully initialized")
	} else {
		fmt.Println("SDK failed to initialize!", err)
		return
	}
	user := featureprobe.NewUser().With("userId", "00001")

	detail := fp.BoolDetail("campaign_allow_list", user, false)
	fmt.Println("Result =>", detail.Value)
	fmt.Println("       => reason:", detail.Reason)
	fmt.Println("       => rule index:", detail.RuleIndex)

	fp.Close()
}
