#!/bin/bash
set -e

curl -LO https://feeldayone-public.s3.ap-northeast-2.amazonaws.com/release/latest/kubenx.tar.gz
gzip -d kubenx.tar.gz
tar -xzvf kubenx.tar
mv kubenx /usr/local/bin
kubenx version

rm -rf kubenx.tar*

cat > $HOME/.kubenx/config<<EOF
{
  "session_name": "Role name you want assume from",
  "assume": {
    "dev" : "arn:aws:iam::22222:role/role-name",
    "stg" : "arn:aws:iam::33333:role/role-name",
    "prod" : "arn:aws:iam::11111:role/role-name",
    "security" : "arn:aws:iam::44444:role/role-name"
  },
  "eks-assume-mapping": {
    "eks-prod-apnortheast2": "prod",
    "eks-dev-apnortheast2": "dev",
    "eks-stg-apnortheast2": "stg",
    "eks-security-apnortheast2": "prod",
  }
}
EOF