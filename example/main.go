package main

import (
	"fmt"
	"time"

	featureprobe "github.com/featureprobe/server-sdk-go"
)

func main() {
	config := featureprobe.FPConfig{
		RemoteUrl:       "http://127.0.0.1:4007",
		ServerSdkKey:    "server-8ed48815ef044428826787e9a238b9c6a479f98c",
		RefreshInterval: 1000, // ms
		WaitFirstResp:   true,
	}
	fp, err := featureprobe.NewFeatureProbe(config)
	if err != nil {
		fmt.Println(err)
		return
	}
	user := featureprobe.NewUser("user")

	for {
		detail := fp.StrDetail("color_ab_test", user, "black")
		fmt.Println(detail)
		time.Sleep(time.Duration(5) * time.Second)
	}
}
