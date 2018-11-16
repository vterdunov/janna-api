# janna-api
Janna provides a REST API interface for some VMware/vSphere functions. Like deploy Virtual Machnine from OVA, manage snapshots.

# Quick start
- Choose a [docker tag](https://hub.docker.com/r/vterdunov/janna-api/tags/)
- Pull image `docker pull vterdunov/janna-api:<tag>`
- Pass desired environment variables using `--env` or `--env-file` directives. And run it:  
```
docker run -d --name=janna-api --env-file=envfile vterdunov/janna-api:<tag>
```

## Configuration
Janna accept environment variables as its config.  
See available environment variables examples in [.env.example](https://github.com/vterdunov/janna-api/blob/master/config/.env.example). E.g.:
```bash
export VMWARE_URL=username:password@vsphere.address.com
export VMWARE_INSECURE=1
export VMWARE_DATACENTER=DC1
export VMWARE_FOLDER=DevVMs
```

## Development
- Install Go v1.11+ environment.
- Copy `cp .env.example .env` and change env file.

- Compile and Run
```
make start
```
or using Docker and Docker Compose
```
make dc
```

Run `make help` to additional useful commands.

## API docs
See [OpenAPI file](https://github.com/vterdunov/janna-api/blob/master/api/openapi.yaml)
