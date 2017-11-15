# Examples
These examples are not intended to be run directly.  Rather, they serve as a reference that can be consulted when creating your own `mu.yml` files.

For detailed steps to create your own project, check out the [quickstart](https://github.com/stelligent/mu/wiki/Quickstart#steps).

Scheduled Tasks Notes:
  * Due to the way ECS containerOverrides work, your Dockerfile must
    contain a CMD line, and not an ENTRYPOINT line.
  * As of November 15, 2017, only ECS deployment is supported.
    This means you must have 'provider: ecs' in your services: section.
  * The commands must be provided as a JSON array. See the example mu.yml file in this directory.

