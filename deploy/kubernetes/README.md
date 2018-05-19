## Debug Dry-run install
helm install --dry-run --debug ./janna-api/

## Deploy
helm install --name=janna-api -f values.yaml ./janna-api/
