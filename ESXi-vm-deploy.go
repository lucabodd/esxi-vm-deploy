package main

import (
	"flag"
	"os"
	"fmt"
)

func main() {
	fmt.Println("vim-go")
	conf := flag.String("c", "","Specify the configuration file.")
	vm_name := flag.String("vm-name", "", "Specify virtual machine name")
	vm_os := flag.String("vm-os", "debian10-64", "Specify virtual machine OS")
	vm_cpu := flag.Int("cpu", 1, "Specify virtual machine CPU")
	vm_disk_size := flag.Int("disk-size", 50, "Specify .vmdk size")
	vm_ram := flag.Int("ram", 16, "Specify RAM size")
	vm_net := flag.String("vm-net", "", "Specify virtual network")
	vm_ipv4 := flag.String("ip", "", "Virtual machine IP address")
	esxi_datastore := flag.String("datastore", "", "Specify Datastore")
	esxi_host := flag.String("esxi", "", "Specify ESXi Host")
	helper_host := flag.String("helper", "", "Specify helper host")
	ip_automatic_lookup := flag.Bool("ip-automatic-lookup", false, "Specify if script should seek for an unallocated IP address")

    flag.Parse()

	if *conf != "" {
		file, err := os.Open(*conf)
		check(err)
		fmt.Println(file)
	} else {
		fmt.Println("[*] Configuration file checking other flags")

	}
}

func check(e error) {
	if e != nil {
		fmt.Println(e)
		panic(e)
	}
}
