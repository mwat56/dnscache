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
		- [Example Use Cases](#example-use-cases)
			- [1. HTTP Client with DNS Caching](#1-http-client-with-dns-caching)
			- [2. Load Balancing with Random IP Selection](#2-load-balancing-with-random-ip-selection)
			- [3. Microservice Communication with DNS Caching](#3-microservice-communication-with-dns-caching)
			- [4. Graceful Shutdown](#4-graceful-shutdown)
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

The cache is thread safe. Create a new instance by specifying how long each entry should be cached (in minutes). Items will be refreshed in the background.

```go
// refresh items every 5 minutes
resolver := dnscache.New(5)

// get an array of net.IP
ips, _ := resolver.Fetch("api.google.de")

// get the first net.IP
ip, _ := resolver.FetchOne("api.google.de")

// get the first net.IP as string
ip, _ := resolver.FetchOneString("api.google.de")
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

#### 2. Load Balancing with Random IP Selection

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
// Create a DNS resolver with 5-minute refresh
resolver := dnscache.New(5)

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
	resolver.Close()

	// Perform other cleanup...
} // shutdown()
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
