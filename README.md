# dnscache

[![golang](https://img.shields.io/badge/Language-Go-green.svg)](https://golang.org/)
[![GoDoc](https://godoc.org/github.com/mwat56/dnscache?status.svg)](https://godoc.org/github.com/mwat56/dnscache)
[![Go Report](https://goreportcard.com/badge/github.com/mwat56/dnscache)](https://goreportcard.com/report/github.com/mwat56/dnscache)
[![Issues](https://img.shields.io/github/issues/mwat56/dnscache.svg)](https://github.com/mwat56/dnscache/issues?q=is%3Aopen+is%3Aissue)
[![Size](https://img.shields.io/github/repo-size/mwat56/dnscache.svg)](https://github.com/mwat56/dnscache/)
[![Tag](https://img.shields.io/github/tag/mwat56/dnscache.svg)](https://github.com/mwat56/dnscache/tags)
[![View examples](https://img.shields.io/badge/learn%20by-examples-0077b3.svg)](https://github.com/mwat56/dnscache/blob/main/_demo/demo.go)
[![License](https://img.shields.io/github/mwat56/dnscache.svg)](https://github.com/mwat56/dnscache/blob/main/LICENSE)

- [dnscache](#dnscache)
	- [Purpose](#purpose)
	- [Installation](#installation)
	- [Usage](#usage)
		- [Configuration Options](#configuration-options)
			- [Available Options](#available-options)
				- [Basic usage with default options except refresh interval:](#basic-usage-with-default-options-except-refresh-interval)
				- [Advanced usage with custom options:](#advanced-usage-with-custom-options)
		- [Example Use Cases](#example-use-cases)
			- [1. HTTP Client with DNS Caching](#1-http-client-with-dns-caching)
			- [2. Load Balancing by Random IP Selection](#2-load-balancing-by-random-ip-selection)
			- [3. Microservice Communication with DNS Caching](#3-microservice-communication-with-dns-caching)
			- [4. Graceful Shutdown](#4-graceful-shutdown)
		- [Runtime Metrics](#runtime-metrics)
	- [Libraries](#libraries)
	- [Licence](#licence)

----

## Purpose

The `dnscache` package provides a simple, thread-safe DNS caching mechanism for Go applications. It helps reduce DNS lookup latency and network overhead by caching resolved IP addresses and optionally refreshing them in the background.

Key features:

- Thread-safe DNS resolution caching,
- Optional background refresh of cached entries,
- Simple API for fetching IPs (as arrays, single values, or strings),
- Random IP selection for load balancing.

## Installation

You can use `Go` to install this package for you:

```bash
go get github.com/mwat56/dnscache
```

## Usage

The cache is thread safe. Create a new instance by specifying how long each entry should be cached (in minutes). IPs can be refreshed in the background; the default refresh interval is `0` (disabled). The default number of lookup retries is `3`.

```go
// Refresh items every 5 minutes
resolver := dnscache.New(5)

// get an array of net.IP
ips, _ := resolver.Fetch("api.google.de")

// get the first net.IP
ip, _ := resolver.FetchOne("api.google.de")

// get the first net.IP as string
ip, _ := resolver.FetchOneString("api.google.de")
```

### Configuration Options

The `dnscache` package offers several configuration options through the `TResolverOptions` struct:

```go
type TResolverOptions struct {
	DNSservers      []string
	CacheSize       int
	Resolver        *net.Resolver
	ExpireInterval  uint8
	MaxRetries      uint8
	RefreshInterval uint8
	TTL             uint8
}
```

#### Available Options

- `DNSservers`: List of DNS servers to use, `nil` means use system default.
- `CacheSize`: Initial size of the DNS cache, `0` means use default ( `64`)
- `Resolver`: Custom DNS resolver to use, `nil` means use default (`net.DefaultResolver`)
- `ExpireInterval`: How often to remove expired entries in minutes (`0` disables background expiration).
- `MaxRetries`: Maximum number of retry attempts for DNS lookups, `0` means use default (`3`).
- `RefreshInterval`: How often to refresh cached entries in minutes, `0` disables background refresh.
- `TTL`: Time to live for cache entries in minutes, `0` means use default (`64`).

One may use any of the options or just a subset of them – every option has a default value to use if not explicitly specified.

##### Basic usage with default options except refresh interval:

```go
// Refresh cached hosts every 5 minutes
resolver := dnscache.New(5)
```

This way is probably the most common practice to create a resolver. It uses all default values except for the refresh interval.

##### Advanced usage with custom options:

```go
// Refresh items every 10 minutes, use a custom resolver, and retry each
// lookup up to 5 times, using 2 DNS servers while assigning each lookup
// result a TTL of 16 minutes:
resolver := dnscache.NewWithOptions(dnscache.TResolverOptions{
	DNSservers:      []string{"8.8.8.8", "8.8.4.4"},
	CacheSize:       128,
	Resolver:        myCustomResolver,
	MaxRetries:      5,
	RefreshInterval: 10,
	TTL:             16,
})
```

### Example Use Cases

#### 1. HTTP Client with DNS Caching

Improve HTTP client performance by caching DNS lookups:

```go
// Create a DNS resolver with 10-minute refresh
resolver := dnscache.New(10)

// Create an HTTP client with custom transport
client := &http.Client{
	Transport: &http.Transport{
		MaxIdleConnsPerHost: 64,
		Dial: func(aNetwork string, aAddress string) (net.Conn, error) {
			separator := strings.LastIndex(aAddress, ":")
			host := aAddress[:separator]
			port := aAddress[separator:]

			ip, err := resolver.FetchRandomString(host)
			if nil != err {
				return nil, err
			}

			// Connect using the resolved IP
			return net.Dial("tcp", ip + port)
		}, // Dial
		TLSHandshakeTimeout: 10 * time.Second,
	}, // Transport
	Timeout: 30 * time.Second,
} // client

// Use the client
response, err := client.Get("https://example.com/")
if nil != err {
	log.Fatalf("Error: %v", err)
}
defer response.Body.Close()
```

#### 2. Load Balancing by Random IP Selection

Distribute connections across multiple IPs for a single hostname:

```go
// Create a DNS resolver with 15-minute refresh
resolver := dnscache.New(15)

func connectToService(aService string) (net.Conn, error) {
	// Get a random IP for the service (using Yoda-style comparison)
	ip, err := resolver.FetchRandomString(aService)
	if nil != err {
		return nil, fmt.Errorf("DNS resolution failed: %w", err)
	}

	// Connect to the randomly selected IP
	conn, err := net.Dial("tcp", ip + ":443")
	if nil != err {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	return conn, nil
} // connectToService()
```

#### 3. Microservice Communication with DNS Caching

Improve inter-service communication in a microservice architecture:

```go
// Create a DNS resolver with 3-minute refresh
resolver := dnscache.New(3)

// Service discovery function
func getServiceEndpoint(aServiceName string) (string, error) {
	ip, err := resolver.FetchOneString(aServiceName + ".internal")
	if nil != err {
		return "", fmt.Errorf("service discovery failed: %w", err)
	}

	return fmt.Sprintf("http://%s:8080", ip), nil
} // getServiceEndpoint()

// Call another service
func callUserService(aUserID string) (*User, error) {
	endpoint, err := getServiceEndpoint("user-service")
	if nil != err {
		return nil, err
	}

	response, err := http.Get(endpoint + "/users/" + aUserID)
	if nil != err {
		return nil, fmt.Errorf("service call failed: %w", err)
	}
	defer response.Body.Close()

	// Process response...
} // callUserService()
```

#### 4. Graceful Shutdown

Properly close the resolver when your application shuts down:

```go
// Create a DNS resolver with background refresh
resolver := dnscache.New(10)

// Use resolver throughout application lifetime...

// When shutting down
func shutdown() {
	// Stop background refresh goroutine
	resolver.StopRefresh()
	// Stop background expiration goroutine
	resolver.StopExpire()

	// Perform other cleanup...
} // shutdown()
```

### Runtime Metrics

The `dnscache` package provides metrics for monitoring the performance and health of the DNS cache. The metrics can be accessed through the `Metrics()` method of the `TResolver` instance:

```go
// Create a DNS resolver with background refresh
resolver := dnscache.New(10)

// … do some work …

// Get the current metrics data
metrics := resolver.Metrics()
```

The metrics are returned as a `*TMetrics` struct, which provides the following fields:

- `Lookups`: Total number of DNS lookups,
- `Hits`: Number of cache hits,
- `Misses`: Number of cache misses,
- `Retries`: Number of lookup retries,
- `Errors`: Number of lookup errors,
- `Peak`: Peak number of cached entries.

The field values are a snapshot of the current state at the time of requesting the metrics and get updated atomically as the resolver does its work. In other words, the metrics may change while you are reading them. Hence, in case some sort of statistics are to be calculated, it is recommended to request the metrics data at regular intervals and then work with the respective snapshot.

The metrics can be printed to the console using the `String()` method of the `TMetrics` struct:

```go
metrics := resolver.Metrics()
// … do some analysing …
fmt.Println(metrics.String())
```

## Libraries

The following external libraries were used building `dnscache`:

* _No external libraries were used to build this library._

## Licence

        Copyright © 2025 M.Watermann, 10247 Berlin, Germany
                        All rights reserved
                    EMail : <support@mwat.de>

> This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.
>
> This software is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
>
> You should have received a copy of the GNU General Public License along with this program. If not, see the [GNU General Public License](http://www.gnu.org/licenses/gpl.html) for details.

----
[![GFDL](https://www.gnu.org/graphics/gfdl-logo-tiny.png)](http://www.gnu.org/copyleft/fdl.html)
