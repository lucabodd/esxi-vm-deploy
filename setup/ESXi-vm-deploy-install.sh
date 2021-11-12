!#/bin/bash
# check if ansible is installed
if [ $(dpkg-query -W -f='${Status}' ansible 2>/dev/null | grep -c "ok installed") -eq 0 ];
then
  apt install -y ansible;
fi
wget https://dl.google.com/go/go1.14.4.linux-amd64.tar.gz && tar -xvf go1.14.4.linux-amd64.tar.gz -C /usr/local
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
echo 'export GOROOT=/usr/local/go' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$GOPATH/bin:$GOROOT/bin:$PATH' >> ~/.bashrc
echo "checking go installed version"
go version
echo "Environment settings"
go env
echo "Installing scanner ..."
go get github.com/lucabodd/esxi-vm-deploy
go install github.com/lucabodd/esxi-vm-deploy
echo "Done!"
echo "please 'source ~/.bashrc' "
