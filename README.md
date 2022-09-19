# Terraform Provider for IndyKite

- Website: [terraform.io](https://terraform.io)
- Tutorials: [learn.hashicorp.com](https://learn.hashicorp.com/terraform?track=getting-started#getting-started)
- Forum: [discuss.hashicorp.com](https://discuss.hashicorp.com/c/terraform-providers/tf-indykite/)
- Chat: [gitter](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing List: [Google Groups](http://groups.google.com/group/terraform-tool)

The Terraform IndyKite provider is a plugin for Terraform that allows for the full
lifecycle management of IndyKite resources.
This provider is maintained internally by the IndyKite Provider team.

Please note: We take Terraform's security and our users' trust very seriously.
If you believe you have found a security issue in the IndyKite Terraform Provider,
please responsibly disclose by contacting us at security@indykite.com.

## Quick Starts

- [Using the provider](https://www.terraform.io/docs/providers/indykite/index.html)
- [Provider development](docs/DEVELOPMENT.md)

## Documentation

Full, comprehensive documentation is available on the Terraform website:

[Documentation](https://terraform.io/docs/providers/indykite/index.html)

## Roadmap

Our roadmap for expanding support in Terraform for IndyKite resources can be found
in our [Roadmap](ROADMAP.md) which is published quarterly.

## Frequently Asked Questions

Responses to our most frequently asked questions can be found in our
[FAQ](docs/FAQ.md )

## Contributing

The Terraform IndyKite Provider is the work of many contributors.
We appreciate your help!

To contribute, please read the contribution guidelines:
[Contributing to Terraform - IndyKite Provider](docs/CONTRIBUTING.md)

## Debug Provider

[Install delve](https://github.com/go-delve/delve/blob/master/Documentation/installation/README.md)

On Go version 1.16 or later, this command will also work:

```shell
 $ go install github.com/go-delve/delve/cmd/dlv@latest
```

### Starting A Provider In Debug Mode

Run your debugger, and pass it the provider binary as the command to run, specifying whatever flags,
environment variables, or other input is necessary to start your provider in debug mode:

Set the environment variable and point to the service credential configuration file.

```shell
export INDYKITE_APPLICATION_CREDENTIALS_FILE=/{FIXME}/service_credential.json
```

Start the provider in remote debug mode:

```shell
 $ dlv --listen=:40000 --headless --api-version=2 exec ./terraform-provider-indykite -- --debug
```

Connect your debugger (whether it's your IDE or the debugger client) to the debugger server.
Have it continue execution (it pauses the process by default) and it will print output like the following to stdout:

Provider started, to attach Terraform set the TF_REATTACH_PROVIDERS env var:

```shell
  TF_REATTACH_PROVIDERS='{"terraform.indykite.com/indykite/indykite":{"Protocol":"grpc" ....}}'
```

#### Running Terraform With A Provider In Debug Mode

Copy the line starting with `TF_REATTACH_PROVIDERS` from your provider's output.
Either export it, or prefix every Terraform command with it:

```shell
 $ export TF_REATTACH_PROVIDERS='{"terraform.indykite.com/indykite/indykite":{"Protocol":"grpc" ....}}'
 $ terraform apply
```

Run Terraform as usual. Any breakpoints you have set will halt execution and show you the current variable values.
