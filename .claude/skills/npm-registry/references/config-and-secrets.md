# npm registry — config & secrets at run time

## The runtime is the consumer's machine

A registry target has no host to inject configuration. Configuration
reaches the CLI **when it runs, on the user's machine**, from its flags,
the environment it was started in (`process.env`) and its per-repo config
file. That is the product's recorded choice
(`config: flags-and-repo-config-file`). How those sources are parsed,
ranked and validated belongs to the language and framework components, not
to this target.
The `environment.md` catalog for `cli` lists the names it **reads**, never
values. The package itself contains none.

## Nothing is baked in at build time

A build that inlines `process.env.X` into the bundle ships the build
machine's value to every consumer, unchanged, forever. Every version is
immutable. So the build reads no environment that changes behaviour. Any
variable the CLI honours is read when it starts, from the consumer's
environment, as `process.env` gives it.

This includes anything that acts like configuration: an endpoint, a
default path, a log level, a feature switch.

## Do not auto-load the user's `.env`

Node can load a `.env` file into `process.env`. By default
`process.loadEnvFile()` and `--env-file` read `./.env` relative to the
current directory. For a CLI run inside someone else's repository, that
file is **their project's** environment, often full of their secrets, and
the tool has no business reading it. Do not call `loadEnvFile()` in the
shipped entry point, and do not put `--env-file` in the shebang. If
development needs a `.env`, load it in the dev task, not in published code.
(Principle: least privilege. Loading that file gives the tool secrets its
task never needed.)

## Secrets: the CLI holds none of its own

This product has no accounts and calls no service of its own
(`auth: none`, `integrations: []`), so there is no secret the CLI needs.
Keep it that way inside the package. A credential compiled into a published
tarball is published to everyone, and an immutable registry cannot recall
it. The only remedy is rotation.

If a future feature needs a user's credential, it reaches the CLI at run
time from the user's side, never from the package.
