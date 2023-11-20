package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

const (
	startMarker = "#multipass-hosts" + lineBreak
	endMarker   = "#/multipass-hosts" + lineBreak
	configFile  = ".multipass-hosts.json"
)

type multipassEntry struct {
	Name string   `json:"name"`
	IPv4 []string `json:"ipv4"`
}

type multipassList struct {
	List []multipassEntry `json:"list"`
}

func (l *multipassList) findIPv4(name string) ([]string, bool) {
	for _, entry := range l.List {
		if entry.Name == name {
			return entry.IPv4, true
		}
	}
	return nil, false
}

func execMultipassList() (list *multipassList, err error) {
	out, err := exec.Command("multipass", "list", "--format", "json").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run multipass: %w", err)
	}
	list = &multipassList{}
	err = json.Unmarshal(out, &list)
	return list, err
}

func generateBlock(list *multipassList) string {
	buf := &bytes.Buffer{}
	buf.WriteString(startMarker)
	for _, entry := range list.List {
		for i, ip := range entry.IPv4 {
			if i == 0 {
				// only output first ip address
				fmt.Fprintf(buf, "%s %s%s", ip, entry.Name, lineBreak)
			}
		}
	}
	buf.WriteString(endMarker)
	return buf.String()
}

func readHostsFile() (string, error) {
	hostsBytes, err := os.ReadFile(hostsFile)
	if err != nil {
		return "", fmt.Errorf("failed to read hosts file %q: %w", hostsFile, err)
	}
	return string(hostsBytes), nil
}

func writeHostsFile(hosts string) error {
	err := os.WriteFile(hostsFile, []byte(hosts), 0o644)
	if err != nil {
		return fmt.Errorf("failed to write hosts to %s: %w", hostsFile, err)
	}
	return nil
}

func replaceOrAppendBlock(hosts string, block string) string {
	start := strings.Index(hosts, startMarker)
	end := strings.Index(hosts, endMarker)
	if start != -1 && end != -1 {
		return hosts[:start] + block + hosts[end+len(endMarker):]
	}
	return hosts + block
}

type config struct {
	Aliasses map[string][]string `json:"aliasses"`
}

func readConfig() (config config, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return config, fmt.Errorf("read config: failed to get user home dir: %w", err)
	}
	configBytes, err := os.ReadFile(path.Join(home, configFile))
	if os.IsNotExist(err) {
		return config, nil
	}
	if err != nil {
		return config, fmt.Errorf("read config: failed to read: %w", err)
	}
	err = json.Unmarshal(configBytes, &config)
	return config, err
}

func expandAliasses(list *multipassList, config config) *multipassList {
	for name, aliasses := range config.Aliasses {
		ipv4, ok := list.findIPv4(name)
		if !ok {
			continue
		}
		for _, alias := range aliasses {
			list.List = append(list.List, multipassEntry{
				Name: alias,
				IPv4: ipv4,
			})
		}
	}
	return list
}

var flagPrint = flag.Bool("print", false, "Set to true to print the output to stdout")
var flagUpdate = flag.Bool("update", true, "Set to false to skip updating the hosts file")

func main() {
	flag.Parse()

	list, err := execMultipassList()
	if err != nil {
		log.Fatal(err)
	}
	config, err := readConfig()
	if err != nil {
		log.Fatal(err)
	}
	newList := expandAliasses(list, config)
	block := generateBlock(newList)
	hosts, err := readHostsFile()
	if err != nil {
		log.Fatal(err)
	}
	newHosts := replaceOrAppendBlock(hosts, block)

	if *flagPrint {
		fmt.Println(newHosts)
	}

	if !*flagUpdate {
		return
	}
	if err = writeHostsFile(newHosts); err != nil {
		log.Fatal(err)
	}
}
