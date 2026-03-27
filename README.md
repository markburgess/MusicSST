
# Scan music module


This shows how to merge knowledge stuff with data stuff


## Install NFS stuff

First install all NFS, RPC stuff, open firewall ports

 sudo systemctl enable rpcbind
 sudo systemctl start rpcbind

Mount the music:

sudo mount -t nfs 192.168.0.251:/Recordings /mnt

Then run http_server -resources /mnt

