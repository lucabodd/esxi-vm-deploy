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
	var vm_net string
	var vm_datastore string
	reader := bufio.NewReader(os.Stdin)

	/*vm_name := flag.String("vm-name", "", "Specify virtual machine name")
	vm_os := flag.String("vm-os", "debian10-64", "Specify virtual machine OS")
	vm_cpu := flag.Int("cpu", 1, "Specify virtual machine CPU")
	vm_disk_size := flag.Int("disk-size", 50, "Specify .vmdk size")
	vm_ram := flag.Int("ram", 16, "Specify RAM size")
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


		fmt.Println(vm_net)
		fmt.Println(vm_datastore)

	}
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
