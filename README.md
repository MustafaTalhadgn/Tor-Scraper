# 🧅 Go Tor Scraper & CTI Collector

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![Network](https://img.shields.io/badge/Network-Tor%20Network-purple)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

A specialized automation tool developed in **Go (Golang)** to perform bulk scanning and data collection from **.onion** (Dark Web) addresses via the Tor network.

This project was built to demonstrate "Collection" and "Automation" competencies in Cyber Threat Intelligence (CTI) processes, solving the problem of manually analyzing hundreds of potential leak sites.

## 🎯 Project Architecture

The application consists of 4 main modules designed for anonymity and resilience:

1.  **Input Handler:** Parses target URLs from a `targets.yaml` file, cleaning whitespace and validating inputs.
2.  **Tor Proxy Client:** Routes all HTTP traffic through a local Tor SOCKS5 proxy (`127.0.0.1:9050/9150`) using a custom `http.Transport` to prevent IP leaks.
3.  **Fault Tolerance:** Implements robust error handling; dead or timed-out onion sites are logged and skipped without crashing the scraper.
4.  **Data Collector:** Saves the HTML content of accessible sites and generates a status report log.

## ✨ Key Features

* **Anonymous Routing:** Uses `golang.org/x/net/proxy` to tunnel traffic strictly through Tor.
* **Bulk Scanning:** Processes multiple targets sequentially from a configuration file.
* **Resilient Execution:** Continues operation even if individual targets are down (common in Tor network).
* **Logging:** Generates detailed success/error logs (`[INFO]` vs `[ERR]`) for post-scan analysis.

## 🛠️ Installation & Usage

### Prerequisites
* **Go** installed on your machine.
* **Tor Service** running in the background (listening on port `9050` or `9150`).

### 1. Configuration
Create a `targets.yaml` file in the root directory:

```yaml
sites:
  - [http://exampleonionaddress.onion](http://exampleonionaddress.onion)
  - [http://anotherhiddenwiki.onion](http://anotherhiddenwiki.onion)
  - [https://check.torproject.org](https://check.torproject.org)
