# FeatureProbe Server Side SDK for Golang (Alpha Version)

[![Top Language](https://img.shields.io/github/languages/top/FeatureProbe/server-sdk-go)](https://github.com/FeatureProbe/server-sdk-go/search?l=go)
[![codecov](https://codecov.io/gh/featureprobe/server-sdk-go/branch/main/graph/badge.svg?token=TAN3AU4CK2)](https://codecov.io/gh/featureprobe/server-sdk-go)
[![Github Star](https://img.shields.io/github/stars/FeatureProbe/server-sdk-go)](https://github.com/FeatureProbe/server-sdk-go/stargazers)
[![Apache-2.0 license](https://img.shields.io/github/license/FeatureProbe/FeatureProbe)](https://github.com/FeatureProbe/FeatureProbe/blob/main/LICENSE)

This is alpha version and should not be considered ready for production use while this message is visible.

Feature Probe is an open source feature management service. This SDK is used to control features in Golang programs. This
SDK is designed primarily for use in multi-user systems such as web servers and applications.

## Fundamental Concept

Reading the short [fundamental concept](https://github.com/FeatureProbe/FeatureProbe/blob/main/BASIC_CONCEPTS.md) will help to understand the code blow more easily.  [中文](https://github.com/FeatureProbe/FeatureProbe/blob/main/BASIC_CONCEPTS_CN.md)

## Try Out Demo Code

We provide a runnable demo code for you to understand how FeatureProbe SDK is used.

1. Start FeatureProbe Service with docker composer. [How to](https://github.com/FeatureProbe/FeatureProbe#1-starting-featureprobe-service-with-docker-compose)
2. Download this repo and run the demo program:
```bash
git clone https://github.com/FeatureProbe/server-sdk-go.git
cd server-sdk-go
go run example/main.go
```
3. Find the Demo code in [example](https://github.com/FeatureProbe/server-sdk-go/tree/main/example), 
do some change and run the program again.
```bash
go run main.go
```

## Step-by-Step Guide

In this guide we explain how to use feature toggles in a Golang application using FeatureProbe.

### Step 1. Install the Golang SDK

Fisrt import the FeatureProbe SDK in your application code:

```go
import "github.com/featureprobe/server-sdk-go"
```

Fetch the FeatureProbe SDK as a dependency in `go.mod`.

```shell
go get github.com/featureprobe/server-sdk-go
```

### Step 2. Create a FeatureProbe instance

After you install and import the SDK, create a single, shared instance of the FeatureProbe sdk.

```go

config := featureprobe.FPConfig{
    RemoteUrl:       "https://127.0.0.1:4007",
    ServerSdkKey:    "serverSdkKey",
    RefreshInterval: 1000,
}

fp, err := featureprobe.NewFeatureProbe(config)
```

### Step 3. Use the feature toggle

You can use sdk to check which variation a particular user will receive for a given feature flag.

```go
user := featureprobe.NewUser("user")
val := fp.BoolValue("bool_toggle", user, true)
```

### Step 4. Unit Testing (Optional)

```go
toggles := map[string]interface{}{}
toggles["bool_toggle"] = true

fp := featureprobe.NewFeatureProbeForTest(toggles)
user := featureprobe.NewUser("user")

assert.Equal(t, fp.BoolValue("bool_toggle", user, false), true)
```

## Testing SDK

We have unified integration tests for all our SDKs. Integration test cases are added as submodules for each SDK repo. So
be sure to pull submodules first to get the latest integration tests before running tests.

```shell
git pull --recurse-submodules
go test
```

## Contributing

We are working on continue evolving FeatureProbe core, making it flexible and easier to use.
Development of FeatureProbe happens in the open on GitHub, and we are grateful to the
community for contributing bugfixes and improvements.

Please read [CONTRIBUTING](https://github.com/FeatureProbe/featureprobe/blob/master/CONTRIBUTING.md)
for details on our code of conduct, and the process for taking part in improving FeatureProbe.
