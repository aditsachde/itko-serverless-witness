# Serverless Witness

## Deploy to GCP

### TF Part 1

Create a new GCP project and fill the ID in terraform.tfvars
Select the regions to deploy to and fill into terraform.tfvars
Fill the current git sha short hash (git rev-parse --short HEAD) in terraform.tfvars

Finally, run `terraform apply`. This will create some resources, but is expected to fail.

### Non TF stuff

Deploy a database somewhere. I used cockroachdb serverless. Make sure to have the connection string handy.

Generate some credentials
`go run ./cmd/generate-config -db-url 'REPLACE' > .env`

Go to the following URL on your project:
https://console.cloud.google.com/security/secret-manager/secret/configuration/versions

Click new version and paste the value from CONFIG in the .env file generated.

The rest of the file contains the public keys corresponding to the private keys in config. Delete the CONFIG line but keep the rest of the file.

Build and push your docker image

```
export PROJECT_ID=<replace>
docker build --platform linux/amd64 -t "us-central1-docker.pkg.dev/$PROJECT_ID/witness/witness:$(git rev-parse --short HEAD)" .
docker push us-central1-docker.pkg.dev/$PROJECT_ID/witness/witness:$(git rev-parse --short HEAD)
```

### TF Part 2

Run `terraform apply` again.

The TF run will have spit out a bunch of urls, one for each region the service was deployed to. CNAMEs can be configured in your own DNS provider.
