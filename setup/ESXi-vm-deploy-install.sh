sudo apt-get install ansible
wget https://dl.google.com/go/go1.14.4.linux-amd64.tar.gz
tar -xvf go1.14.4.linux-amd64.tar.gz
sudo mv go /usr/local
# remove go folder in case /usr/local/go is not empty or if move is not allowed.
rm -rf go
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
go get github.com/lucabodd/ESXi-vm-deploy
go install github.com/lucabodd/ESXi-vm-deploy
echo "Done!"
echo "please 'source ~/.bashrc' "
