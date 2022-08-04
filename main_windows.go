package main

import "os"

const lineBreak = "\r\n"

func getHostsFile() string {
	return os.Getenv("SystemRoot") + `\System32\drivers\etc\hosts`
}
