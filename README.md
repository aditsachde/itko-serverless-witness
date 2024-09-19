# Serverless Witness

A couple witnesses can be found at the following domains:

```
https://europe-north1.gcp.witness.itko.dev
https://us-central1.gcp.witness.itko.dev
https://us-east1.gcp.witness.itko.dev
https://us-west1.gcp.witness.itko.dev
```

```
europe-north1.gcp.witness.itko.dev+aea5e162+AcheK8sShpWLJwXaV3y+yRvA9VSKSjFw2+I/2wNaV6qO
us-central1.gcp.witness.itko.dev+a6044e5a+AemmFe58r+TtVpwsPj9nJ3FIhoPBjG+/dbHrUN0Bi1JQ
us-east1.gcp.witness.itko.dev+5b4c83cb+AbwpB77LBXaJ1ht21Bh18OHf1nUJ8XM2H6x67Fe56gq7
us-west1.gcp.witness.itko.dev+5fe7b537+AaaMycrWp2dQgBdye1B40/yU5TpKLgyRTGP5YiFl+jRK
```

## Deploy to GCP

### TF Part 1

- Create a new GCP project and fill the ID in terraform.tfvars
- Select the regions to deploy to and fill into terraform.tfvars
- Fill the current git sha short hash (git rev-parse --short HEAD) in terraform.tfvars
- Specify a verified base domain (such as gcp.witness.itko.dev) in terraform.tfvars

Finally, run `terraform apply`. This will create some resources, but is expected to fail.

### Non TF stuff

Deploy a database somewhere. I used cockroachdb serverless. Make sure to have the connection string handy.

Create the table in the database `go run ./cmd/init-db -db-url 'REPLACE'`

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

The TF run will have spit out a bunch of urls, one for each region the service was deployed to. CNAME these to ghs.googlehosted.com in your DNS provider.
