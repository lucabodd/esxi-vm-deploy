package main

import (
	"flag"
	"os"
	"fmt"
	"log"
	ansibler "github.com/apenella/go-ansible"
	"bytes"
	"strings"
	"github.com/tidwall/gjson"
)

func main() {
	//set ansible env vars
	os.Setenv("ANSIBLE_STDOUT_CALLBACK", "json")
	os.Setenv("ANSIBLE_HOST_KEY_CHECKING", "False")

	//vars
	var conf string
	var esxi_host string

	/*vm_name := flag.String("vm-name", "", "Specify virtual machine name")
	vm_os := flag.String("vm-os", "debian10-64", "Specify virtual machine OS")
	vm_cpu := flag.Int("cpu", 1, "Specify virtual machine CPU")
	vm_disk_size := flag.Int("disk-size", 50, "Specify .vmdk size")
	vm_ram := flag.Int("ram", 16, "Specify RAM size")
	vm_net := flag.String("vm-net", "", "Specify virtual network")
	vm_ipv4 := flag.String("ip", "", "Virtual machine IP address")
	*/
	flag.StringVar(&conf, "c", "", "Specify the configuration file.")
	flag.StringVar(&esxi_host, "esxi", "", "Specify ESXi Host")
	/*
	helper_host := flag.String("helper", "", "Specify helper host")
	ip_automatic_lookup := flag.Bool("ip-automatic-lookup", false, "Specify if script should seek for an unallocated IP address")
	*/
    flag.Parse()

	if conf != "" {
		file, err := os.Open(conf)
		check(err)
		//parse configuration files
		fmt.Println(file)
	} else {
		fmt.Println("[*] Configuration file not provided parsing flags")
		ansiblePlaybookConnectionOptions := &ansibler.AnsiblePlaybookConnectionOptions{}
		ansiblePlaybookOptions := &ansibler.AnsiblePlaybookOptions{
			Inventory: esxi_host+",",
		}
		stdout := new(bytes.Buffer)
		playbook := &ansibler.AnsiblePlaybookCmd{
			Playbook:          "playbooks/gather-esxi-info.yml",
			ConnectionOptions: ansiblePlaybookConnectionOptions,
			Options:           ansiblePlaybookOptions,
			ExecPrefix:        "",
			Writer:				stdout,
		}
		err := playbook.Run()
		check(err)
		json_stdout := strings.Replace(stdout.String(), "=>", "", -1)
		json_stdout = strings.Replace(json_stdout, json_stdout[len(json_stdout)-23:], "", -1)
		esxi_vmnet := gjson.Get(json_stdout, "plays.0.tasks.0.hosts.*.stdout_lines")
		vmnets := esxi_vmnet.Array()
		esxi_datastores := gjson.Get(json_stdout, "plays.0.tasks.1.hosts.*.stdout_lines")
		datastores := esxi_datastores.Array()

		fmt.Println(vmnets)
		fmt.Println(datastores)

	}
}

func check(e error) {
	if e != nil {
		fmt.Println(e)
		panic(e)
	}
}
