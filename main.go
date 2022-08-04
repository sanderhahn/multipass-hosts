package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func execMultipassList() *multipassList {
	out, err := exec.Command("multipass", "list", "--format", "json").Output()
	if err != nil {
		log.Fatal(err)
	}
	list := &multipassList{}
	json.Unmarshal(out, &list)
	return list
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

func readHostsFile() string {
	hostsFile := getHostsFile()
	hostsBytes, err := ioutil.ReadFile(hostsFile)
	if err != nil {
		log.Fatal(err)
	}
	return string(hostsBytes)
}

func writeHostsFile(hosts string) {
	hostsFile := getHostsFile()
	err := ioutil.WriteFile(hostsFile, []byte(hosts), 0o644)
	if err != nil {
		log.Printf("Failed to write %s content:\n%s", hostsFile, hosts)
		log.Fatal(err)
	}
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

func readConfig() *config {
	var config config
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	configBytes, err := ioutil.ReadFile(path.Join(home, configFile))
	if os.IsNotExist(err) {
		return &config
	}
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(configBytes, &config)
	return &config
}

func expandAliasses(list *multipassList, config *config) *multipassList {
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

func main() {
	list := execMultipassList()
	config := readConfig()
	newList := expandAliasses(list, config)
	block := generateBlock(newList)
	hosts := readHostsFile()
	newHosts := replaceOrAppendBlock(hosts, block)
	writeHostsFile(newHosts)
}
