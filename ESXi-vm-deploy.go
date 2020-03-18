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
	"bufio"
	"strconv"
)

func main() {
	//set ansible env vars
	os.Setenv("ANSIBLE_STDOUT_CALLBACK", "json")
	os.Setenv("ANSIBLE_HOST_KEY_CHECKING", "False")

	//vars
	var conf string
	var esxi_host string
	var vm_name string
	var vm_os string
	var vm_net string
	var vm_datastore string
	var vm_cpu string
	var vm_ram string
	var vm_ipv4 string
	var vm_disk_size string
	var helper_host string
	reader := bufio.NewReader(os.Stdin)

	//Flag parsing
	flag.StringVar(&conf, "c", "", "Specify the configuration file.")
	flag.StringVar(&esxi_host, "esxi", "", "Specify ESXi Host")
	flag.StringVar(&vm_name, "vm-name", "", "Specify virtual machine name")
	flag.StringVar(&vm_os, "vm-os", "debian10-64", "Specify virtual machine OS")
	flag.StringVar(&vm_ram, "vm-ram", "16", "Specify RAM size")
	flag.StringVar(&vm_cpu, "vm-cpu", "2", "Specify RAM size")
	flag.StringVar(&vm_ipv4, "vm-ip", "", "Virtual machine IP address")
	flag.StringVar(&vm_disk_size, "vm-disk-size", "50", "Virtual machine IP address")
	flag.StringVar(&helper_host, "helper", "", "Virtual machine IP address")
	flag.BoolVar(&help, "help", false, "prints this help message")
    flag.Parse()
	if help {
		flag.PrintDefaults()
		kill("ERR: NOT ENAUGH ARGS")
	}
	//end of flag parsing

	//check if configuration file is provided
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

		//VMNET selection
		if len(vmnets) == 1 {
			vm_net = vmnets[0].Str
			log.Println("[+] Only 1 available networks found, selecting "+ vm_net)
		} else {
			log.Println("[*] Found more than one virtual network, choose where you want to deploy VM")
			fmt.Println("Available networks:")
			for index,element := range vmnets {
				fmt.Println(index,"-",element)
			}
			fmt.Print("Insert ID: ")
			in, _ := reader.ReadString('\n')
			id, _ := strconv.Atoi(in)
			vm_net = vmnets[id].Str
		}
		//DATASTORE selection
		if len(datastores) == 1 {
			vm_datastore = datastores[0].Str
			log.Println("[+] Only 1 available datastore found, selecting "+ vm_datastore)
		} else {
			log.Println("[*] Found more than one datastore, choose where you want to deploy VM")
			fmt.Println("Available datastores:")
			for index,element := range datastores {
				fmt.Println(index,"-",element)
			}
			fmt.Print("Insert ID: ")
			in, _ := reader.ReadString('\n')
			id, _ := strconv.Atoi(in)
			vm_datastore = datastores[id].Str
		}
	}

	//
}

func check(e error) {
	if e != nil {
		fmt.Println(e)
		panic(e)
	}
}
func kill(reason string) {
	fmt.Println(reason)
	os.Exit(1)
}
