# MySQL Brute Forcer

is a simple multithreaded MySQL brute-forcing utility written in Go. It supports custom wordlists for usernames and passwords, proxy support, and optional TLS connections. The tool is optimized for speed and flexibility, making it suitable for penetration testing and security research.

---

## Features

- Multithreaded brute-forcing with adjustable concurrency.
- Supports username and password wordlists.
- Proxy support for routing traffic.
- Rate limiting for evasion.
- Batch processing of large target lists.
- Optional TLS support for MySQL connections.
- Suppresses excessive logging and handles transient errors gracefully.

---

## Installation

`go install github.com/Vulnpire/msforce@latest`

## Command-Line Options

```
Usage of mysql-brute:
  -c int
        Number of concurrent threads (default 5)
  -proxy string
        HTTP/HTTPS proxy (e.g., http://127.0.0.1:8080)
  -tls
        Enable TLS for MySQL connections
  -t int
        Connection timeout in seconds (default 5)
  -u string
        Path to username wordlist
  -v    Enable verbose output
  -w string
        Path to password wordlist

```

### Input Format

Provide target IP addresses or hostnames via standard input.

    ```
    echo -e "192.168.1.10\n192.168.1.20" | ./main -w passwords.txt -u usernames.txt
    ```

### Example Usage

    echo "192.168.1.10" | ./main -w passwords.txt -u usernames.txt -c 10 -v

### TLS Connections

    echo "secure.mysqlserver.com" | ./main -w passwords.txt -u usernames.txt -tls

### Wordlist Format

Username Wordlist: A text file with one username per line. Example:

    root
    admin
    user

Password Wordlist: A text file with one password per line. Example:

    password
    123456
    admin

### Axiom Support

```
» cat ~/.axiom/modules/msforce.json
[{
        "command":"cat input | msforce -u ~/lists/mysql-user.txt -p ~/lists/mysql-pass.txt | anew output",
        "ext":"txt"
}]
```

### Known Issues

Let me know?

### Disclaimer

This tool is intended for educational purposes and authorized penetration testing only. Unauthorized use is strictly prohibited. Use responsibly.
