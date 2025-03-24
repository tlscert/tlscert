# tlscert.dev

This monorepo holds the client and server implementation for the tlscert.dev service.

## Development

We make use of [mise](https://mise.jdx.dev/) to manage developer tooling. After installing mise, it should automatically make the tooling available when inside the repository folder.

Run `mise local:up` to start a local kubernetes development cluster with the tlscert server running. After that, you can run the client with `mise client`.
