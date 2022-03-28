package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type nslookup struct {
	nsType string
	value  string
}

var LookupAddress string

var outputNSlookup []nslookup

func init() {
	fmt.Println("Creating your DNSee File.")
	fmt.Print("Loading")
	time.Sleep(1 * time.Second)
	fmt.Print(".")
	time.Sleep(1 * time.Second)
	fmt.Print(".")
	time.Sleep(1 * time.Second)
	fmt.Print(".")
	fmt.Print(".")
	time.Sleep(1 * time.Second)
	fmt.Print(".")

	flag.StringVar(&LookupAddress, "d", "google.com", "Type here what domain you want to check.")
	flag.Parse()
}

func main() {

	ns, e := NameLookup(LookupAddress)

	if e != nil {
		log.Fatal(e)
	}

	AnyLookup(LookupAddress, ns)

	WriteToFile(LookupAddress)
}

func ImportData(F_nsType string, data []string) {

	for _, c := range data {
		if strings.Contains(c, F_nsType) {

			switch F_nsType {
			case "nameserver":
				NSvalue := strings.Split(c, "= ")
				outputNSlookup = append(outputNSlookup, nslookup{
					nsType: "NS",
					value:  NSvalue[1],
				})
			case "Address":
				NSvalue := strings.Split(c, ": ")
				outputNSlookup = append(outputNSlookup, nslookup{
					nsType: "A",
					value:  NSvalue[1],
				})
			case "internet address":
				NSvalue := strings.Split(c, "= ")
				outputNSlookup = append(outputNSlookup, nslookup{
					nsType: "A",
					value:  NSvalue[1],
				})
			case "AAAA":
				NSvalue := strings.Split(c, "address ")
				outputNSlookup = append(outputNSlookup, nslookup{
					nsType: "AAAA",
					value:  NSvalue[1],
				})
			case "AAAA IPv6":
				NSvalue := strings.Split(c, "= ")
				outputNSlookup = append(outputNSlookup, nslookup{
					nsType: "AAAA",
					value:  NSvalue[1],
				})
			case "canonical name":
				NSvalue := strings.Split(c, "= ")
				outputNSlookup = append(outputNSlookup, nslookup{
					nsType: "CNAME",
					value:  NSvalue[1],
				})
			case "text":
				NSvalue := strings.Split(c, "= ")
				outputNSlookup = append(outputNSlookup, nslookup{
					nsType: "TXT",
					value:  NSvalue[1],
				})
			case "\"":
				NSvalue := strings.Split(c, "\t")
				outputNSlookup = append(outputNSlookup, nslookup{
					nsType: "TXT",
					value:  NSvalue[1],
				})
			case "mail exchanger":
				NSvalue := strings.Split(c, "= ")
				outputNSlookup = append(outputNSlookup, nslookup{
					nsType: "MX",
					value:  NSvalue[1],
				})
			case "MX preference":
				split := strings.Split(c, ", ")

				preference := strings.Split(split[0], "= ")
				mx := strings.Split(split[1], "= ")

				NSvalue := preference[1] + " " + mx[1]
				outputNSlookup = append(outputNSlookup, nslookup{
					nsType: "MX",
					value:  NSvalue,
				})

			default:
				NSvalue := strings.Split(c, "= ")
				outputNSlookup = append(outputNSlookup, nslookup{
					nsType: F_nsType,
					value:  NSvalue[1],
				})
			}
		}
	}
}

func AnyLookup(address string, nameserver string) {

	cmd, _ := exec.Command("nslookup", "-type=any", address, nameserver).Output()

	stdout := string(cmd)

	var outputArray []string

	if runtime.GOOS == "darwin" {
		OutputAll := strings.Split(stdout, "#53\n")

		output := OutputAll[1]

		outputArray = strings.Split(output, "\n")
	} else {
		outputArray = strings.Split(stdout, "\n")
	}

	ImportData("nameserver", outputArray)
	if runtime.GOOS == "darwin" {
		ImportData("Address", outputArray)
		ImportData("AAAA", outputArray)
		ImportData("text", outputArray)
	} else {
		ImportData("internet address", outputArray)
		ImportData("AAAA IPv6", outputArray)
		ImportData("\"", outputArray)
	}
	ImportData("canonical name", outputArray)
	ImportData("mail exchanger", outputArray)
	ImportData("MX preference", outputArray)
}

func NameLookup(address string) (string, error) {

	cmd, e := exec.Command("nslookup", "-type=ns", address, "1.1.1.1").Output()

	stdout := string(cmd)

	// Check for any availble network connection
	_, netErrors := http.Get("https://www.google.com")

	if netErrors != nil {
		return "", fmt.Errorf("Connection time out: Can't connect to server or domain. Please check your internet connection")
	}

	if runtime.GOOS != "windows" {
		if e != nil {
			return "", fmt.Errorf("There has been an error within your command. Please check your input")
		}
	}

	output := strings.Split(stdout, "\n")
	var newOutput []string

	// Windows doesnt return any errors so we need to check if the OS used is currently Windows for the correct error response
	if runtime.GOOS == "windows" {

		arrayLenght := len(output) - 3

		if strings.HasPrefix(output[arrayLenght], "Address:") {
			return "", fmt.Errorf("%s is not a valid domain. Please check your domainname for any mistakes", address)
		}
	}

	for _, c := range output {
		if strings.HasPrefix(c, address) {
			newOutput = append(newOutput, c)
		}
	}

	output = newOutput

	output = strings.Split(output[0], "= ")
	// Removing any enters that may have entered the nameserver.
	output = strings.Split(output[1], "\r")

	return output[0], nil
}

func WriteToFile(address string) error {
	timeNow := strconv.Itoa(time.Now().Day()) + "-" + strconv.Itoa(int(time.Now().Month())) + "-" + strconv.Itoa(time.Now().Year())
	f, e := os.OpenFile(address+" - "+timeNow+".txt", os.O_CREATE|os.O_WRONLY, 0644)

	if e != nil {
		return fmt.Errorf(e.Error())
	}

	var err error
	defer f.Close()

	_, err = f.WriteString("NameLookup - " + address + "\n")
	_, err = f.WriteString("\n")
	_, err = f.WriteString("> -- Nameserver Records -- <\n")
	for _, c := range outputNSlookup {
		if c.nsType == "NS" {
			_, err = f.WriteString(c.value + "\n")
		}
	}
	_, err = f.WriteString("\n")
	_, err = f.WriteString("> -- IP Address -- <\n")
	for _, c := range outputNSlookup {
		if c.nsType == "A" {
			_, err = f.WriteString(c.value)
		}
	}

	_, err = f.WriteString("\n\n")
	_, err = f.WriteString("> -- AAAA Records -- <\n")
	for _, c := range outputNSlookup {
		if c.nsType == "AAAA" {
			_, err = f.WriteString(c.value)
		}
	}

	if err != nil {
		return fmt.Errorf(e.Error())
	}

	return nil
}
