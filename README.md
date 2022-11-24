# glubng
BNG Control plane for FDio VPP and Kea DHCP written in Go

## Forward API Unix Socket
In order to develop this control plane sometimes is useful to forward VPP Unix socket from vpp device to a development machine. We can use SSH forwarding capabilities:

```
ssh root@<vpp-management-ip> -L<local-sock>:/run/vpp/vpp.sock
```

In this projecte we also have another socket for Kea triggering. We also should forward this socket:
```
ssh root@<vpp-management-ip> -R<local-sock>:/hook.sock
```