# ruller-dsl-feature-flag
A feature flag engine that can be used to enable, change or rollout features of a system dynamically based on system or user attributes

A system can check for enabled features by performing a REST call to (ex.: /menu) having a JSON body with some input attributes (user info, environment info etc). Some conditions will be evaluated and a JSON containing all enabled features will be returned. Then the client system can use this information to decide on what to enable/disable/configure from its internals.

This DSL tool will get a JSON written with some feature tree rules and generate a Go code that can be run as a REST service. We use [Ruller](http://github.com/flaviostutz/ruller) framework on our code generation and it will be responsible for the runtime execution of those rules.

While developing, enter '/sample' dir and perform ```docker-compose build``` so that you can run your code against sample rules json files and check for results.

If you want to create your own feature flag service, fork https://github.com/flaviostutz/ruller-sample-feature-flag and create your own rules.

## Usage

* Create a project structure just like https://github.com/flaviostutz/ruller-sample-feature-flag

* Create infra.json

```json
{
    "_config": {
        "flatten": true,
    },
    "_items": [{
            "provider": "aws",
            "_condition": "randomPerc(10, input:customerid)"
        },
        {
            "provider": "azure",
            "_condition": "randomPercRange(10, 50, input:customerid)"
        },
        {
            "provider": "vpsdime"
        }
    ]
}
```

* Create a docker-compose.yml

```yml
version: '3.5'
services:
  sample:
    build: .
    environment:
      - LOG_LEVEL=info
    ports:
      - 3001:3000
```

* Run server

```sh
docker-compose up -d --build
```

* Execute some queries to determine which infra structure to use

```sh
curl -X POST \
  http://localhost:3000/rules/infra \
  -H 'Content-Type: application/json' \
  -H 'X-Forwarded-For: 177.79.35.49' \
  -H 'cache-control: no-cache' \
  -d '{}
```

## Runtime parameters

* The same as http://github.com/flaviostutz/ruller. Check documentation for details

## Feature selection language

* The language is a JSON file with some semantics regarding to feature attributes and condition attributes organized in a tree, so that attributes and conditions from a parent are inherited by its children

* Features are identified by an id and may have some custom attributes bound to it. All features whose "condition" attribute evaluates to true will be returned as the result of the REST call

* All special functions that this DSL supports can be seen on "_condition" fields on the following sample files
  * [menu.json](https://github.com/flaviostutz/ruller-sample-feature-flag/blob/master/rules/menu.json)
  * [domains.json](https://github.com/flaviostutz/ruller-sample-feature-flag/blob/master/rules/domains.json)
  * [screens.json](https://github.com/flaviostutz/ruller-sample-feature-flag/blob/master/rules/screens.json)

