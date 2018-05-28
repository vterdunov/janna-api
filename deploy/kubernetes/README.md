# Deploy Janna API to Kubernetes using Helm

## Deploy
Change `values.yaml`
`helm install --name=janna-api -f values.yaml ./janna-api/`

## Upgrade
`helm upgrade janna-api ./janna-api/`

## Delete
`helm delete janna-api`

## Debug Dry-run install
`helm install --dry-run --debug ./janna-api/`
