# GKE Policy Automation development

## Requirements

* [Go](https://go.dev/doc/install) 1.22 or newer (to build the application)
* [GNU Make](https://www.gnu.org/software/make) (to build and test easier)
* [Open Policy Agent](https://www.openpolicyagent.org/docs/latest/#1-download-opa) (to test REGO policies)

## Building the application

*Note:* This project uses [Go Modules](https://blog.golang.org/using-go-modules)
making it safe to work with it outside of your existing [GOPATH](http://golang.org/doc/code.html#GOPATH).
The instructions that follow assume a directory in your home directory outside of
the standard GOPATH (i.e `$HOME/development/`).

1. Clone GKE Policy Automation repository

   ```sh
   mkdir -p $HOME/development; cd $HOME/development 
   git clone https://github.com/google/gke-policy-automation.git
   ```

2. Enter the application directory and compile it

   ```sh
   cd gke-policy-automation
   make build
   ```

## Testing the application

* To run unit tests, use make `test` target

  ```sh
  make test
  ```

* To check code and report suspicious constructs, use make `vet` target

  ```sh
  make vet
  ```

* To check code formatting, use make `fmtcheck` target

  ```sh
  make fmtcheck
  ```

## Testing the REGO rules

The application repository comes with a set of [recommended REGO rules](./gke-policies-v2/) that cover
GKE cluster best practices. Rego rules can be tested with [OPA Policy Testing framework](https://www.openpolicyagent.org/docs/latest/policy-testing/).

*NOTE*: `-v` flag sets verbose reporting mode.

```sh
opa test <POLICY_DIR> -v
```

To test set of project policies:

```sh
opa test gke-policies-v2 -v
```

## Developing REGO rules

Please check [GKE Policy authoring guide](./gke-policies-v2/README.md) for guides on authoring REGO rules
for GKE Policy Automation.
