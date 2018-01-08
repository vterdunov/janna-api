# janna-api
Learning Go with implements https://github.com/vterdunov/janna with Go and [rbvmomi](https://github.com/vmware/rbvmomi)

## Dependencies
- Install Golang
- Install `go dep`  
```
go get -u github.com/golang/dep/cmd/dep
```  
- Install dependencies
```
make dep
```

## Start
- Configure
```bash
export VMWARE_URI=username:password@vsphere.address.com
```
- Compile and Run
```
make run
```
