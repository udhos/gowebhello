#!/bin/bash
# CodeDeploy:
yum -y update
yum install -y ruby
cd /home/ec2-user
curl -O https://aws-codedeploy-us-east-1.s3.amazonaws.com/latest/install
chmod +x ./install
./install auto
# GoLang:
cd /usr/local
rm go
wget https://storage.googleapis.com/golang/go1.8rc3.linux-amd64.tar.gz
rm -rf go1.8rc3.linux-amd64
tar xf go1.8rc3.linux-amd64.tar.gz
mv go go1.8rc3.linux-amd64
ln -s go1.8rc3.linux-amd64 go
