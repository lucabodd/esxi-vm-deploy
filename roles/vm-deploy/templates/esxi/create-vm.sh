#!/bin/bash
#
##### VM Creation Script #####################################
#Script Version 1.0
#Author L. Bodini
#
#----------------------------------------+
#Custom Variable Section for Modification|
#----------------------------------------+---------------------
#NVM is name of virtual machine(NVM). No Spaces allowed in name
#NVMDIR is the directory which holds all the VM files
#NVMOS specifies VM Operating System, below a collection for some possible values
#NVMSIZE is the size of the virtual disk to be created
#VMMEMSIZE defines the size of the physical memory
#--------------------------------------------------------------
###############################################################
### Default Variable settings - change this to your preferences
NVM="{{ vm_name  }}"
NVMDIR=$NVM
#NVM OS example
#debain10-64
NVMOS="{{ vm_os }}"
VMVCPU="{{ vm_cpu  }}"
NVMSIZE="{{ vm_disk_size  }}g" # Size of Virtual Machine Disk
VMMEMSIZE={{ vm_ram  }} # Default Memory Size In GB
VMMEMSIZE=$(( $VMMEMSIZE*1024  ))
VMNET="{{ vm_net }}"
STORAGEDIR="/vmfs/volumes/{{ vm_datastore }}"
### End Variable Declaration


mkdir $STORAGEDIR/$NVMDIR
exec 6>&1 # Sets up write to file
exec 1>$STORAGEDIR/$NVMDIR/$NVM.vmx # Open file
# write the WMX configuration
echo config.version = '"'8'"'
echo virtualHW.version = '"'13'"' # I keep this to 12 in order to assure compatibility between esxi version
echo numvcpus = '"'$VMVCPU'"'
echo memsize = '"'$VMMEMSIZE'"'
echo floppy0.present = '"'FALSE'"'
echo displayName = '"'$NVM'"' # name of virtual machine
echo guestOS = '"'$NVMOS'"'
echo
echo powerType.powerOff = '"'default'"'
echo powerType.reset = '"'default'"'
echo powerType.suspend = '"'soft'"'
echo
echo ide0:0.present = '"'TRUE'"'
echo ide0:0.deviceType = '"'cdrom-raw'"'
echo ide:0.startConnected = '"'false'"'
echo floppy0.startConnected = '"'FALSE'"'
echo
echo ethernet0.virtualDev = '"'e1000e'"'
echo ethernet0.networkName = '"'$VMNET'"'
echo ethernet0.addressType = '"'generated'"'
echo ethernet0.uptCompatibility = '"'TRUE'"'
echo ethernet0.present = '"'TRUE'"'
echo
echo scsi0.present = '"'true'"'
echo scsi0.sharedBus = '"'none'"'
echo scsi0.virtualDev = '"'pvscsi'"'
echo scsi0:0.present = '"'true'"' # Virtual Disk Settings
echo scsi0:0.fileName = '"'$NVM.vmdk'"'
echo scsi0:0.deviceType = '"'scsi-hardDisk'"'
echo
echo sched.cpu.min = '"'0'"'
echo sched.cpu.shares = '"'normal'"'
echo sched.mem.min = '"'0'"'
echo sched.mem.minSize = '"'0'"'
echo sched.mem.shares = '"'normal'"'
echo nvram = '"'$NVM.nvram'"'
echo pciBridge0.present = '"'TRUE'"'
echo svga.present = '"'TRUE'"'
echo pciBridge4.present = '"'TRUE'"'
echo pciBridge4.virtualDev = '"'pcieRootPort'"'
echo pciBridge4.functions = '"'8'"'
echo pciBridge5.present = '"'TRUE'"'
echo pciBridge5.virtualDev = '"'pcieRootPort'"'
echo pciBridge5.functions = '"'8'"'
echo pciBridge6.present = '"'TRUE'"'
echo pciBridge6.virtualDev = '"'pcieRootPort'"'
echo pciBridge6.functions = '"'8'"'
echo pciBridge7.present = '"'TRUE'"'
echo pciBridge7.virtualDev = '"'pcieRootPort'"'
echo pciBridge7.functions = '"'8'"'
echo numa.autosize.cookie = '"'10001'"'
echo numa.autosize.vcpu.maxPerVirtualNode = '"'4'"'

# close file
exec 1>&-
# make stdout a copy of FD 6 (reset stdout), and close FD6
exec 1>&6
exec 6>&-
# Change permissions on the file so it can be executed by anyone
chmod 755 $STORAGEDIR/$NVMDIR/$NVM.vmx
#Creates Virtual disk
cd $STORAGEDIR/$NVMDIR #change to the VM dir
vmkfstools -c $NVMSIZE $NVM.vmdk
#register VM
vim-cmd solo/registervm $STORAGEDIR/$NVMDIR/$NVM.vmx
