package main

import (
	"flag"
	"fmt"
	ansibler "github.com/lucabodd/go-ansible"
	"github.com/schollz/progressbar"
	"github.com/tidwall/gjson"
	"go/build"
	"log"
	"os"
	"strings"
	"strconv"
	"time"
)

func main() {
	//set ansible env vars
	os.Setenv("ANSIBLE_STDOUT_CALLBACK", "json")
	os.Setenv("ANSIBLE_HOST_KEY_CHECKING", "False")
	current_version := "1.2.10"

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
	var verbose bool
	var version bool

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
	flag.BoolVar(&verbose, "verbose", false, "enable verbose mode")
	flag.Parse()

	// retrive bin directory
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	datadir := gopath+"/src/github.com/lucabodd/esxi-vm-deploy"


	if version {
		fmt.Println("esxi-vm-deploy version: ", current_version)
		fmt.Println("see CHANGELOG.md for latest version changes\ncopy available under "+datadir+"/CHANGELOG.md\nor at https://github.com/lucabodd/esxi-vm-deploy/blob/master/CHANGELOG.md")
		kill("")
	}
	if esxi_host == "" || vm_name == "" || vm_ipv4 == "" || helper_host == "" || help {
		fmt.Println("Usage: esxi-vm-deploy [OPTIONS]")
		flag.PrintDefaults()
		fmt.Println("One ore more required flag has not been prodided.")
		fmt.Println("Note that using less flag than required could lead program into errors \nOmit flags only if you are aware of what are you doin'")
		fmt.Println("[EXAMPLES]")
		fmt.Println("- Creation of machine with custom hardware")
		fmt.Println("esxi-vm-deploy --esxi [ESXi host defined in ssh config] --helper [unix host defined in ssh config]  --vm-ip [ip of new machine] --vm-name [name of new machine] --vm-ram 8  --vm-disk-size 16 --vm-cpu 4")
		fmt.Println("- Creation of machine with default values 3 CPU 50GB Disk 16GB RAM")
		fmt.Println("esxi-vm-deploy --esxi [ESXi host defined in ssh config] --helper [unix host defined in ssh config]  --vm-ip [ip of new machine] --vm-name [name of new machine]")
		kill("")
	}
	//end of flag parsing

	/*
		############################################################################
		#							VARS COLLECTION 							   #
		############################################################################
	*/
	log.Println("[*] Checking System...")

	playbook := &ansibler.PlaybookCmd{
		Playbook:          datadir+"/playbooks/esxi-check-duplicate.yml",
		ConnectionOptions: &ansibler.PlaybookConnectionOptions{},
		Options:           &ansibler.PlaybookOptions{
			Inventory: esxi_host + ",",
			ExtraVars: map[string]interface{}{
				"vm_name": vm_name,
			},
		},
	}
	res, err := playbook.Run()
	check(err)
	err = res.PlaybookResultsChecks()
	check(err)
	verboseOut(res.RawStdout, verbose)
	duplicate_stdout := gjson.Get(res.RawStdout, "plays.0.tasks.1.hosts.*.stdout")
	duplicate := duplicate_stdout.Int()
	if duplicate > 0 {
		log.Println("[-] A VM with name "+vm_name+" is already registered")
		kill("ERR: VMNAME ALREADY REGISTERED")
	}

	fmt.Println("[+] System checks passed ... starting")

	fmt.Println("[*] Gathering ESXi host info")
	playbook = &ansibler.PlaybookCmd{
		Playbook:          datadir+"/playbooks/esxi-gather-info.yml",
		ConnectionOptions: &ansibler.PlaybookConnectionOptions{},
		Options:           &ansibler.PlaybookOptions{
			Inventory: esxi_host + ",",
		},
	}
	res = &ansibler.PlaybookResults{}
	res,err = playbook.Run()
	check(err)
	err = res.PlaybookResultsChecks()
	check(err)
	verboseOut(res.RawStdout, verbose)
	esxi_vmnet := gjson.Get(res.RawStdout, "plays.0.tasks.0.hosts.*.stdout_lines")
	vmnets := esxi_vmnet.Array()
	esxi_datastores := gjson.Get(res.RawStdout, "plays.0.tasks.1.hosts.*.stdout_lines")
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

	//check datastore available space
	playbook = &ansibler.PlaybookCmd{
		Playbook:          datadir+"/playbooks/esxi-check-datastore.yml",
		ConnectionOptions: &ansibler.PlaybookConnectionOptions{},
		Options:           &ansibler.PlaybookOptions{
			Inventory: esxi_host + ",",
			ExtraVars: map[string]interface{}{
				"datastore": vm_datastore,
			},
		},
	}
	res = &ansibler.PlaybookResults{}
	res,err = playbook.Run()
	check(err)
	err = res.PlaybookResultsChecks()
	check(err)
	verboseOut(res.RawStdout, verbose)
	esxi_available_space := gjson.Get(res.RawStdout, "plays.0.tasks.0.hosts.*.stdout").Str

	//parse result and convert TB to GB
	var esxi_available_space_qty float64
	if strings.Contains(esxi_available_space, "T"){
		esxi_available_space = strings.Trim(esxi_available_space, "T")
		esxi_available_space_qty , err = strconv.ParseFloat(esxi_available_space, 32)
		check(err)
    	esxi_available_space_qty = esxi_available_space_qty * 1024
	} else if strings.Contains(esxi_available_space, "G"){
		esxi_available_space = strings.Trim(esxi_available_space, "G")
		esxi_available_space_qty , err = strconv.ParseFloat(esxi_available_space, 32)
		check(err)
	}

	vm_disk_size_qty , err := strconv.ParseFloat(vm_disk_size, 32)
	check(err)
	if vm_disk_size_qty > esxi_available_space_qty {
		log.Printf("[-] ESXi host has only %.2fG of free memory and cannot suite vm size", esxi_available_space_qty)
		kill("ERR: ESXi NOT ENAUGH SPACE")
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
	playbook = &ansibler.PlaybookCmd{
		Playbook:          datadir+"/playbooks/esxi-deploy-vmx.yml",
		ConnectionOptions: &ansibler.PlaybookConnectionOptions{},
		Options:           &ansibler.PlaybookOptions{
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
		},
	}
	res = &ansibler.PlaybookResults{}
	res, err = playbook.Run()
	check(err)
	err = res.PlaybookResultsChecks()
	check(err)
	verboseOut(res.RawStdout, verbose)
	log.Println("[+] Virtual machine created. disk initialization completed")
	log.Println("[*] Retrieveing VM mac address")
	log.Println("[*] Retrieveing Assigned VM Hypervisor ID")
	mac_stdout := gjson.Get(res.RawStdout, "plays.0.tasks.4.hosts.*.stdout")
	vm_mac_addr := mac_stdout.String()
	id_stdout := gjson.Get(res.RawStdout, "plays.0.tasks.6.hosts.*.stdout")
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

	playbook = &ansibler.PlaybookCmd{
		Playbook:          datadir+"/playbooks/bootp-server-deploy.yml",
		ConnectionOptions: &ansibler.PlaybookConnectionOptions{},
		Options:           &ansibler.PlaybookOptions{
			Inventory: helper_host + ",",
			ExtraVars: map[string]interface{}{
				"vm_ipv4":     vm_ipv4,
				"vm_mac_addr": vm_mac_addr,
				"vm_name":     vm_name,
				"vm_os":        vm_os,
			},
		},
	}
	res = &ansibler.PlaybookResults{}
	res, err = playbook.Run()
	check(err)
	err = res.PlaybookResultsChecks()
	check(err)
	verboseOut(res.RawStdout, verbose)
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
	playbook = &ansibler.PlaybookCmd{
		Playbook:          datadir+"/playbooks/esxi-vm-poweron.yml",
		ConnectionOptions: &ansibler.PlaybookConnectionOptions{},
		Options:           &ansibler.PlaybookOptions{
			Inventory: esxi_host + ",",
			ExtraVars: map[string]interface{}{
				"vm_id": vm_id,
			},
		},
	}
	res = &ansibler.PlaybookResults{}
	res, err = playbook.Run()
	check(err)
	err = res.PlaybookResultsChecks()
	check(err)
	verboseOut(res.RawStdout, verbose)
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
		time.Sleep(900 * time.Millisecond)
	}
	fmt.Println("\n")
	/*
		############################################################################
		#							FILE CLEANUP 		 						   #
		############################################################################
	*/
	log.Println("[*] Running Helper deplfiles cleanup")

	playbook = &ansibler.PlaybookCmd{
		Playbook:          datadir+"/playbooks/helper-depfiles-cleanup.yml",
		ConnectionOptions: &ansibler.PlaybookConnectionOptions{},
		Options:           &ansibler.PlaybookOptions{
			Inventory: helper_host + ",",
		},
	}
	res, err = playbook.Run()
    check(err)
	err = res.PlaybookResultsChecks()
    check(err)
	verboseOut(res.RawStdout, verbose)
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

func verboseOut(message string, verbose bool){
	if verbose {
		fmt.Println(message)
	}
}
