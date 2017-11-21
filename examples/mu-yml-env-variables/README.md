# Examples
These examples are not intended to be run directly.  Rather, they serve as a reference that can be consulted when creating your own `mu.yml` files.

For detailed steps to create your own project, check out the [quickstart](https://github.com/stelligent/mu/wiki/Quickstart#steps).

NOTE: This feature describes *LOCAL* environment variable substitution within `mu.yml`,
which is a completely different feature than described in `./examples/service-env-vars`.  

In this example's mu.yml file, the special syntax `${env:XXXXXX}` is shown.
Lines containing this pattern are replaced with the environment variables
of the user's local environment at the time of parsing.

The syntax `${env:XXXX}` was chosen to distinguish the intention
of local environment substitution from other uses of `${XXXX}`
in CloudFormation, that lack the `env:` prefix.

Although this feature is originally envisioned as being used for
yaml values (after the `:` character), there may also be an
occasional use for yaml keys.  As implemented, the environment
variable substitution occurs on a line-by-line basis during
the parsing of the mu.yml.  

NOTE: This effectively makes the environment an input to `mu`
itself, so if you use this feature extensively, you probably
want to consider a way of controlling this environment via
the creation of a setenv.sh file.  Otherwise you may have
difficulties in replicating results from one user or system
to another.
