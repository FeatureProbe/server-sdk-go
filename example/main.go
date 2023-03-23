package main

import (
	"fmt"
	featureprobe "github.com/featureprobe/server-sdk-go/v2"
	"math/rand"
	"time"
)

func main() {

	// FeatureProbe server URL for local docker
	FeatureProbeServerUrl := "https://featureprobe.io/server" // "https://featureprobe.io/server";

	// Server Side SDK Key for your project and environment
	FeatureProbeServerSdkKey := "server-9e53c5db4fd75049a69df8881f3bc90edd58fb06"

	config := featureprobe.FPConfig{
		RemoteUrl:       FeatureProbeServerUrl,
		ServerSdkKey:    FeatureProbeServerSdkKey,
		RefreshInterval: 2 * time.Second,
		StartWait:       5 * time.Second,
	}

	// Init FeatureProbe, share this FeatureProbe instance in your project.
	fp := featureprobe.NewFeatureProbe(config)
	if !fp.Initialized() {
		fmt.Println("SDK failed to initialize!")
	}

	// Create one user.
	user := featureprobe.NewUser().With("userId", "00001") // "userId" is used in rules, should be filled in.

	// Get toggle result for this user.
	YourToggleKey := "campaign_allow_list"

	// Demo of Bool function.
	isOpen := fp.BoolValue(YourToggleKey, user, false)
	fmt.Println("feature for this user is :", isOpen)

	// Simulate conversion rate of 1000 users for a new feature
	YourCustomEventName := "new_feature_conversion"
	for i := 1; i <= 1000; i++ {
		eventUser := featureprobe.NewUser().StableRollout(fmt.Sprintf("%d", time.Now().UnixNano()/1000000))
		newFeature := fp.BoolValue(YourToggleKey, eventUser, false)
		rand.Seed(time.Now().UnixNano())
		randomNum := rand.Intn(101)
		if newFeature {
			if randomNum <= 55 {
				fp.Track(YourCustomEventName, eventUser, nil)
			}
		} else {
			if randomNum > 55 {
				fp.Track(YourCustomEventName, eventUser, nil)
			}
		}
	}

	fp.Close()
}
