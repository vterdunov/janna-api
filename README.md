# janna-api
Learning Go with implements https://github.com/vterdunov/janna with Go and [govmomi](https://github.com/vmware/govmomi)

## Dependencies
- Golang 

## Start
- Configure

Janna accept environment variables as its config.  
See available environment variables examples in [.env.example](https://github.com/vterdunov/janna-api/blob/master/.env.example). E.g.:
```bash
export VMWARE_URL=username:password@vsphere.address.com
export VMWARE_INSECURE=1
export VMWARE_DC=DC1
```
- Compile and Run
```
make start
```

## API
See [swagger file](https://github.com/vterdunov/janna-api/blob/master/api/swagger.yaml)
