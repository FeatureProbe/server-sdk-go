package main

import (
	"fmt"
	"time"

	featureprobe "github.com/featureprobe/server-sdk-go"
)

func main() {
	config := featureprobe.FPConfig{
		RemoteUrl: "https://featureprobe.io/server",
		// RemoteUrl:       "http://127.0.0.1.4007", // for local docker
		ServerSdkKey:    "server-8ed48815ef044428826787e9a238b9c6a479f98c",
		RefreshInterval: 1000, // ms
		WaitFirstResp:   true,
	}
	fp, err := featureprobe.NewFeatureProbe(config)
	if err != nil {
		fmt.Println(err)
		return
	}
	userId := "uniq_user_id" // unique user id in your business logic
	user := featureprobe.NewUser(userId).With("city", "Paris")

	for {
		detail := fp.NumberDetail("promotion_activity", user, 3.0)
		fmt.Println(detail)
		time.Sleep(time.Duration(5) * time.Second)
	}
}
