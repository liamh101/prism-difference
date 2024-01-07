package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
)

type PrismBaseFile struct {
	Version int         `json:"version"`
	Issues  []PrismItem `json:"issues"`
}

type PrismItem struct {
	Name               string      `json:"name"`
	OriginalRiskRating string      `json:"original_risk_rating"`
	AffectedHosts      []PrismHost `json:"affected_hosts"`
}

type PrismHost struct {
	Ip       string `json:"ip"`
	Hostname string `json:"hostname"`
	Name     string `json:"name"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

type ShorthandIssue struct {
	Name  string
	Hosts []ShorthandHost
}

type ShorthandHost struct {
	Id       string
	Ip       string
	Hostname string
	Name     string
	Port     string
	Protocol string
}

func main() {
	var filenameOne = os.Args[1]
	var filenameTwo = os.Args[2]
	fmt.Println("Looking for Prism File: " + filenameOne)

	prismOneResult := parsePrismFile(filenameOne)

	fmt.Println("Looking for Second Prism File")

	prismTwoResult := parsePrismFile(filenameTwo)

	fmt.Println("Comparing File")

	compareFiles(prismOneResult, prismTwoResult)
}

func parsePrismFile(filename string) PrismBaseFile {
	jsonFile, err := os.Open(filename)

	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("File found")
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result PrismBaseFile
	json.Unmarshal([]byte(byteValue), &result)

	return result
}

func compareFiles(prismOne PrismBaseFile, prismTwo PrismBaseFile) {
	var differences []ShorthandIssue

	shortHandOne := buildShortHandIssues(prismOne)
	shortHandTwo := buildShortHandIssues(prismTwo)

	for _, originalIssue := range shortHandOne {
		found, comparitorIssue, multipleInstances := findIssue(originalIssue, shortHandTwo)

		if !found {
			fmt.Println("Issue not present in latest scan: " + originalIssue.Name)
			continue
		}

		if multipleInstances {
			fmt.Println("Found multiple instances of " + originalIssue.Name)
			continue
		}

		var difference ShorthandIssue
		difference.Name = originalIssue.Name

		for _, originalHost := range originalIssue.Hosts {
			if hasHost(comparitorIssue.Hosts, originalHost) {
				continue
			}

			difference.Hosts = append(difference.Hosts, originalHost)
		}

		if difference.Hosts != nil {
			differences = append(differences, difference)
		}
	}

	displayDifferences(differences)
}

func buildShortHandIssues(baseFile PrismBaseFile) []ShorthandIssue {
	var finalIssues []ShorthandIssue

	for _, issue := range baseFile.Issues {
		var newIssue ShorthandIssue

		newIssue.Name = issue.Name

		for _, host := range issue.AffectedHosts {
			newIssue.Hosts = append(newIssue.Hosts, buildShorthandHost(host))
		}

		finalIssues = append(finalIssues, newIssue)
	}

	return finalIssues
}

func buildShorthandHost(host PrismHost) ShorthandHost {
	var sh ShorthandHost

	sh.Hostname = host.Hostname
	sh.Ip = host.Ip
	sh.Name = host.Name
	sh.Port = strconv.Itoa(host.Port)
	sh.Protocol = host.Protocol
	sh.Id = sh.Ip + sh.Hostname + sh.Name + sh.Port + sh.Protocol

	return sh
}

func findIssue(originalIssue ShorthandIssue, shortHandComparator []ShorthandIssue) (bool, ShorthandIssue, bool) {
	var foundInstance ShorthandIssue
	found := false
	multipleInstances := false

	for _, issue := range shortHandComparator {
		if issue.Name == originalIssue.Name {
			if found {
				multipleInstances = true
			}

			found = true
			foundInstance = issue
		}
	}

	if !found {
		return false, foundInstance, false
	}

	return true, foundInstance, multipleInstances
}

func hasHost(collection []ShorthandHost, host ShorthandHost) bool {
	for _, potentialHost := range collection {
		if potentialHost.Id == host.Id {
			return true
		}
	}

	return false
}

func displayDifferences(differences []ShorthandIssue) {
	if len(differences) == 0 {
		fmt.Println("No differences found!")
		return
	}

	for _, issue := range differences {
		fmt.Println(issue.Name)

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Hostname", "Ip", "Port", "Protocol"})

		for _, host := range issue.Hosts {
			table.Append([]string{host.Name, host.Hostname, host.Ip, host.Port, host.Protocol})
		}

		table.Render()
		fmt.Println()

	}
}
