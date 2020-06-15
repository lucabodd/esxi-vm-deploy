package main

import (
	"bytes"
	"flag"
	"fmt"
	ansibler "github.com/apenella/go-ansible"
	"github.com/schollz/progressbar"
	"github.com/tidwall/gjson"
	"go/build"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	//set ansible env vars
	os.Setenv("ANSIBLE_STDOUT_CALLBACK", "json")
	os.Setenv("ANSIBLE_HOST_KEY_CHECKING", "False")
	version := "0.0"

	//vars
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
	var help bool

	//Flag parsing
	flag.StringVar(&esxi_host, "esxi", "", "ESXi Hypervisor")
	flag.StringVar(&vm_name, "vm-name", "", "Specify virtual machine name")
	flag.StringVar(&vm_os, "vm-os", "debian10-64", "Specify virtual machine OS available: debian9-64, debian10-64")
	flag.StringVar(&vm_ram, "vm-ram", "16", "Specify RAM size")
	flag.StringVar(&vm_cpu, "vm-cpu", "2", "Specify RAM size")
	flag.StringVar(&vm_ipv4, "vm-ip", "", "Virtual machine IP address")
	flag.StringVar(&vm_disk_size, "vm-disk-size", "50", "Virtual machine Disk size")
	flag.StringVar(&helper_host, "helper", "", "BOOTP server address, specified host will provide configurations to booting (PXE) virtual machine")
	flag.BoolVar(&version, "version", false, "Display current version of script")
	flag.BoolVar(&help, "help", false, "prints this help message")
	flag.Parse()
	if esxi_host == "" || vm_name == "" || vm_ipv4 == "" || helper_host == "" || help {
		fmt.Println("Usage: ESXi-vm-deploy [OPTIONS]")
		flag.PrintDefaults()
		fmt.Println("One ore more required flag has not been prodided.")
		fmt.Println("Note that using less flag than required could lead program into errors \nOmit flags only if you are aware of what are you doin'")
		fmt.Println("[EXAMPLES]")
		fmt.Println("- Creation of machine with custom hardware")
		fmt.Println("ESXi-vm-deploy --esxi [ESXi host defined in ssh config] --helper [unix host defined in ssh config]  --vm-ip [ip of new machine] --vm-name [name of new machine] --vm-ram 8  --vm-disk-size 16 --vm-cpu 4")
		fmt.Println("- Creation of machine with default values 3 CPU 50GB Disk 16GB RAM")
		fmt.Println("ESXi-vm-deploy --esxi [ESXi host defined in ssh config] --helper [unix host defined in ssh config]  --vm-ip [ip of new machine] --vm-name [name of new machine]")
		kill("")
	}
	if version {
		fmt.Println("ESXi-vm-deploy version: ", version)
	}
	//end of flag parsing
	// retrive bin directory
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	datadir := gopath+"/src/github.com/lucabodd/ESXi-vm-deploy"

	/*
		############################################################################
		#							VARS COLLECTION 							   #
		############################################################################
	*/
	fmt.Println("[*] Configuration file not provided parsing flags")
	fmt.Println("[*] Checking System...")
	ansiblePlaybookConnectionOptions := &ansibler.AnsiblePlaybookConnectionOptions{}
	ansiblePlaybookOptions := &ansibler.AnsiblePlaybookOptions{
		Inventory: esxi_host + ",",
		ExtraVars: map[string]interface{}{
			"vm_name": vm_name,
		},
	}
	stdout := new(bytes.Buffer)
	playbook := &ansibler.AnsiblePlaybookCmd{
		Playbook:          datadir+"/playbooks/esxi-check-duplicate.yml",
		ConnectionOptions: ansiblePlaybookConnectionOptions,
		Options:           ansiblePlaybookOptions,
		ExecPrefix:        "",
		Writer:            stdout,
	}
	_ = playbook.Run()
	json_stdout := strings.Replace(stdout.String(), "=>", "", -1)
	//fmt.Println(json_stdout)
	//kill("Breakpoint")
	json_stdout = strings.Replace(json_stdout, json_stdout[len(json_stdout)-23:], "", -1)
	duplicate_stdout := gjson.Get(json_stdout, "plays.0.tasks.1.hosts.*.stdout")
	duplicate := duplicate_stdout.String()
	if duplicate != "" {
		kill("ERR: VMNAME ALREADY REGISTERED")
	}
	fmt.Println("[+] System checks passed ... starting")

	fmt.Println("[*] Gathering ESXi host info")
	ansiblePlaybookConnectionOptions = &ansibler.AnsiblePlaybookConnectionOptions{}
	ansiblePlaybookOptions = &ansibler.AnsiblePlaybookOptions{
		Inventory: esxi_host + ",",
	}
	stdout = new(bytes.Buffer)
	playbook = &ansibler.AnsiblePlaybookCmd{
		Playbook:          datadir+"/playbooks/esxi-gather-info.yml",
		ConnectionOptions: ansiblePlaybookConnectionOptions,
		Options:           ansiblePlaybookOptions,
		ExecPrefix:        "",
		Writer:            stdout,
	}
	err := playbook.Run()
	check(err)
	json_stdout = strings.Replace(stdout.String(), "=>", "", -1)
	json_stdout = strings.Replace(json_stdout, json_stdout[len(json_stdout)-23:], "", -1)
	esxi_vmnet := gjson.Get(json_stdout, "plays.0.tasks.0.hosts.*.stdout_lines")
	vmnets := esxi_vmnet.Array()
	esxi_datastores := gjson.Get(json_stdout, "plays.0.tasks.1.hosts.*.stdout_lines")
	datastores := esxi_datastores.Array()

	//VMNET selection
	if len(vmnets) == 1 {
		vm_net = vmnets[0].Str
		log.Println("[+] Only 1 available networks found, selecting " + vm_net)
	} else {
		log.Println("[*] Found more than one virtual network, choose where you want to deploy VM")
		fmt.Println("Available networks:")
		for index, element := range vmnets {
			fmt.Println(index, "-", element)
		}
		fmt.Print("Insert ID: ")
		var id int
		_, err := fmt.Scanf("%d", &id)
		check(err)
		vm_net = vmnets[id].Str

	}
	//DATASTORE selection
	if len(datastores) == 1 {
		vm_datastore = datastores[0].Str
		log.Println("[+] Only 1 available datastore found, selecting " + vm_datastore)
	} else {
		log.Println("[*] Found more than one datastore, choose where you want to deploy VM")
		fmt.Println("Available datastores:")
		for index, element := range datastores {
			fmt.Println(index, "-", element)
		}
		fmt.Print("Insert ID: ")
		var id int
		_, err := fmt.Scanf("%d", &id)
		check(err)
		vm_datastore = datastores[id].Str
	}
	/*
		############################################################################
		#							VARS COLLECTION END 						   #
		############################################################################
	*/
	/*
		############################################################################
		#							VMX DEPLOYMENT  							   #
		############################################################################
	*/
	log.Println("[+] RA passed, deploying .vmx and allocating disk space (thick)")
	log.Println("[*] Deploying .vmx file")
	ansiblePlaybookConnectionOptions = &ansibler.AnsiblePlaybookConnectionOptions{}
	ansiblePlaybookOptions = &ansibler.AnsiblePlaybookOptions{
		Inventory: esxi_host + ",",
		ExtraVars: map[string]interface{}{
			"vm_name":      vm_name,
			"vm_os":        vm_os,
			"vm_cpu":       vm_cpu,
			"vm_disk_size": vm_disk_size,
			"vm_ram":       vm_ram,
			"vm_net":       vm_net,
			"vm_datastore": vm_datastore,
		},
	}
	stdout = new(bytes.Buffer)
	playbook = &ansibler.AnsiblePlaybookCmd{
		Playbook:          datadir+"/playbooks/esxi-deploy-vmx.yml",
		ConnectionOptions: ansiblePlaybookConnectionOptions,
		Options:           ansiblePlaybookOptions,
		ExecPrefix:        "",
		Writer:            stdout,
	}
	err = playbook.Run()
	check(err)
	log.Println("[+] Virtual machine created. disk initialization completed")
	log.Println("[*] Retrieveing VM mac address")
	log.Println("[*] Retrieveing Assigned VM Hypervisor ID")
	json_stdout = strings.Replace(stdout.String(), "=>", "", -1)
	json_stdout = strings.Replace(json_stdout, json_stdout[len(json_stdout)-24:], "", -1)
	mac_stdout := gjson.Get(json_stdout, "plays.0.tasks.4.hosts.*.stdout")
	vm_mac_addr := mac_stdout.String()
	id_stdout := gjson.Get(json_stdout, "plays.0.tasks.6.hosts.*.stdout")
	vm_id := id_stdout.String()
	log.Println("[+] Got physical address: " + vm_mac_addr)
	log.Println("[+] Got VMID: " + vm_id)
	/*
		############################################################################
		#							VMX DEPLOYMENT - END						   #
		############################################################################
	*/
	/*
		############################################################################
		#							PYBOOTD  DEPLOYMENT 						   #
		############################################################################
	*/
	log.Println("[*] uploading BOOTP server")
	ansiblePlaybookConnectionOptions = &ansibler.AnsiblePlaybookConnectionOptions{}
	ansiblePlaybookOptions = &ansibler.AnsiblePlaybookOptions{
		Inventory: helper_host + ",",
		ExtraVars: map[string]interface{}{
			"vm_ipv4":     vm_ipv4,
			"vm_mac_addr": vm_mac_addr,
			"vm_name":     vm_name,
			"vm_os":        vm_os,
		},
	}
	stdout = new(bytes.Buffer)
	playbook = &ansibler.AnsiblePlaybookCmd{
		Playbook:          datadir+"/playbooks/bootp-server-deploy.yml",
		ConnectionOptions: ansiblePlaybookConnectionOptions,
		Options:           ansiblePlaybookOptions,
		ExecPrefix:        "",
		Writer:            stdout,
	}
	err = playbook.Run()
	json_stdout = strings.Replace(stdout.String(), "=>", "", -1)
	json_stdout = strings.Replace(json_stdout, json_stdout[len(json_stdout)-24:], "", -1)
	log.Println("[+] BOOTP server running")
	/*
		############################################################################
		#							PYBOOTD  DEPLOYMENT END						   #
		############################################################################
	*/
	/*
		############################################################################
		#							VM POWERUP     		 						   #
		############################################################################
	*/
	log.Println("[*] Powering up VM")
	ansiblePlaybookConnectionOptions = &ansibler.AnsiblePlaybookConnectionOptions{}
	ansiblePlaybookOptions = &ansibler.AnsiblePlaybookOptions{
		Inventory: esxi_host + ",",
		ExtraVars: map[string]interface{}{
			"vm_id": vm_id,
		},
	}
	stdout = new(bytes.Buffer)
	playbook = &ansibler.AnsiblePlaybookCmd{
		Playbook:          datadir+"/playbooks/esxi-vm-poweron.yml",
		ConnectionOptions: ansiblePlaybookConnectionOptions,
		Options:           ansiblePlaybookOptions,
		ExecPrefix:        "",
		Writer:            stdout,
	}
	err = playbook.Run()
	check(err)
	log.Println("[+] VM " + vm_id + " powered on as " + vm_name)
	/*
		############################################################################
		#							VM POWERUP END 		 						   #
		############################################################################
	*/
	fmt.Println("\n")
	bar := progressbar.Default(100)
	for i := 0; i < 100; i++ {
		bar.Add(1)
		time.Sleep(1200 * time.Millisecond)
	}
	fmt.Println("\n")
	/*
		############################################################################
		#							FILE CLEANUP 		 						   #
		############################################################################
	*/
	log.Println("[*] Running Helper deplfiles cleanup")
	ansiblePlaybookConnectionOptions = &ansibler.AnsiblePlaybookConnectionOptions{}
	ansiblePlaybookOptions = &ansibler.AnsiblePlaybookOptions{
		Inventory: helper_host + ",",
	}
	stdout = new(bytes.Buffer)
	playbook = &ansibler.AnsiblePlaybookCmd{
		Playbook:          datadir+"/playbooks/helper-depfiles-cleanup.yml",
		ConnectionOptions: ansiblePlaybookConnectionOptions,
		Options:           ansiblePlaybookOptions,
		ExecPrefix:        "",
		Writer:            stdout,
	}
	err = playbook.Run()
	check(err)
	log.Println("[+] Cleanup Completed")
	/*
		############################################################################
		#							VM CLEANUP END 		 						   #
		############################################################################
	*/
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
