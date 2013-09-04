#!/bin/sh

echo "current state"
ls -alh /home/ubuntu
ls -alh /home/ubuntu/bin


if [ ! -d $GOROOT ]; then 
  echo "INSTALLING LATEST GO RELEASE"
  cd $HOME
  mkdir -p $GOROOT
  hg clone -u release https://code.google.com/p/go 
  cd $GOROOT/src
  ./all.bash
  export PATH="$GOROOT/bin:$PATH"
fi

if [ ! -d $HOME/boto ]; then
  echo "INSTALLING BOTO"
  pip install -t $HOME/boto boto
fi


if [ ! -d $GOPATH ]; then
  echo "setting up gopath"

  #set up GOPATH
  mkdir -p $GOPATH/src
  mkdir -p $GOPATH/pkg
  mkdir -p $GOPATH/bin
fi


