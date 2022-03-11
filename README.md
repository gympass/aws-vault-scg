# AWS Vault - SSO Config Generator

[![Downloads](https://img.shields.io/github/downloads/Gympass/aws-vault-scg/total.svg)](https://github.com/Gympass/aws-vault-scg/releases)
![Continuous Integration](https://github.com/Gympass/aws-vault-scg/actions/workflows/release.yaml/badge.svg)

Automatically generate your profiles overwriting the file `~/.aws/config` or print on stdout based on your AWS SSO accounts and roles to use with [99designs/aws-vault](https://github.com/99designs/aws-vault)

## **Prerequisites**

- [AWS CLI](https://aws.amazon.com/cli/)
- [AWS-Vault](https://github.com/99designs/aws-vault)

## **Installing**

### **Using [Homebrew](https://brew.sh/)** _(Linux and MacOS)_

```shell
brew tap Gympass/homebrew-tools
brew install aws-vault-scg
```

### **Using go install**

```shell
go install github.com/Gympass/aws-vault-scg
```

### **Or by downloading the** _[latest release](https://github.com/Gympass/aws-vault-scg/releases/latest)_

## Usage

To generate the config file run the command:

```shell
aws-vault-scg generate -s <your-aws-sso-start-url>
```
