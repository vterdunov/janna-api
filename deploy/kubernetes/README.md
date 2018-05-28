# Deploy Janna API to Kubernetes using Helm

## Deploy
Override default `values.yaml` by coping it to `your-values.yaml`
`helm install --name=janna-api --namespace=janna-api -f your-values.yaml ./janna-api/`

## Upgrade
`helm upgrade janna-api ./janna-api/`

## Delete
`helm delete janna-api`

## Debug Dry-run install
`helm install --dry-run --debug ./janna-api/`
