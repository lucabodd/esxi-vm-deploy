# ESXi-vm-deploy
Offering a golang program to automate vm installation on a ESXi host.
The program will be asking for some informations, templating a .vmk file, creating a virtual disk and installing operating system via embedded BOOTP python server.
Before starting the automatic deployment of virtual machines you will need to tune up your system in order to be able to run program.
To run this program you will need:
* go > 1.13
* ansible >= 2.7

## System Setup
In order to setup the system running golang program you will need to follow the follofing steps:
install ansible via package manager (Debian):
```
sudo apt-get install ansible
```
Install golang via package manager
```
sudo apt-get install golang
```
OR from sources
```
wget https://dl.google.com/go/go1.13.3.linux-amd64.tar.gz
tar -xvf go1.13.3.linux-amd64.tar.gz
mv go /usr/local
```
Install ESXi-vm-deploy
```
go get github.com/lucabodd/ESXi-vm-deploy
go install github.com/lucabodd/ESXi-vm-deploy
```
In some cases you might need to export this environment vars
```
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
```

## Usage
Usage is described running the program with the help flag.
Before running the program, please consider the following checklist:
* "helper" host and "vm-ip" must be in the same target subnet (unless you are able to broadcast UDP packets :67,:68 outside the helper's server)
* one helper can deploy one vm per time, if you want to deploy more than one VM at once, you'll need more helpers (or wait :'))

```
Usage: ESXi-vm-deploy [OPTIONS]
One oe more required flag has not been prodided.
Note that using less flag than defined could lead program into errors
Omit flags only if you are aware of what are you doin'
  -esxi string
        Specify ESXi Host
  -help
        prints this help message
  -helper string
        BOOTP server address, specified host will provide configurations to booting (PXE) virtual machine
  -vm-cpu string
        Specify RAM size (default "2")
  -vm-disk-size string
        Virtual machine Disk size (default "50")
  -vm-ip string
        Virtual machine IP address
  -vm-name string
        Specify virtual machine name
  -vm-os string
        Specify virtual machine OS (default "debian10-64")
  -vm-ram string
        Specify RAM size (default "16")
```
