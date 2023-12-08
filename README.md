<div align="center">
	<h1>WorkQueue</h1>
	<img src="assets/logo.png" alt="logo" width="300px">
</div>

# Introduction

WorkQueue is a simple, fast, reliable work queue written in Go. It is designed to be easy to use and to have minimal dependencies. It supports multiple queue types and is designed to be easily extensible which mean you can easily write a new queue type and use it with WorkQueue.

# Queue Types

-   [x] Queue
-   [x] Delaying Queue
-   [x] Priority Queue
-   [x] RateLimiting Queue

# Advantage

-   Simple and easy to use
-   No third-party dependencies
-   High performance
-   Low memory usage
-   Use quadruple heap to maintain the expiration time, effectively reduce the height of the tree, and improve the insertion performance

# Benchmark

# Installation

```bash
go get github.com/shengyanli1982/workqueue
```

# Quick Start

## Queue

```go

```

## Delaying Queue

```go

```

## Priority Queue

```go

```

## RateLimiting Queue

```go

```
