# ruller-dsl-feature-flag
A feature flag engine that can be used to enable, change or rollout features of a system dynamically based on system or user attributes

A system can check for enabled features by performing a REST call to (ex.: /menu) having a JSON body with some input attributes (user info, environment info etc). Some conditions will be evaluated and a JSON containing all enabled features will be returned. Then the client system can use this information to decide on what to enable/disable/configure from its internals.

This was crafted to have complexity of O(1) so that it can support large scale deployments with minimum cost.

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
  -d '{"customerid": "2118"}
```

* In this case, customer "2118" will always use the "aws" infrastructure

```sh
{"_condition_debug":"randomPerc(10, input:customerid)","_rule":"2","provider":"aws"}
```

* Change customerid and see other infra structures being selected

```sh
curl -X POST \
  http://localhost:3000/rules/infra \
  -H 'Content-Type: application/json' \
  -H 'X-Forwarded-For: 177.79.35.49' \
  -H 'cache-control: no-cache' \
  -d '{"customerid": "6843"}
```

```sh
{"_condition_debug":"randomPercRange(10, 50, input:customerid)","_rule":"3","provider":"azure"}
```

## Runtime parameters

* The same as http://github.com/flaviostutz/ruller with the addition of `--templdir`:

1. `--source`: a comma-separated list of Glob patterns or `.json` files where `ruller-dsl-feature-flag` will read the rules from. Attention: Glob patterns must be enclosed by double quotes on the command line;
2. `--target`: the path of the file where the Golang ruller engine will be generated;
3. `--templdir`: the path where the template files can be found;
4. `--log-level`: the level of logging `ruller-dsl-feature-flag` will output to STDOUT;
5. `--condition-debug`: if present, the resulting Golang ruller engine will output debug info when processing feature flags;

## Feature selection language

* The language is a JSON file with some semantics regarding to feature attributes and condition attributes organized in a tree, so that attributes and conditions from a parent are inherited by its children

* Features are identified by an id and may have some custom attributes bound to it. All features whose "condition" attribute evaluates to true will be returned as the result of the REST call

* All special functions that this DSL supports can be seen on "_condition" fields on the following sample files
  * [menu.json](https://github.com/flaviostutz/ruller-sample-feature-flag/blob/master/rules/menu.json)
  * [domains.json](https://github.com/flaviostutz/ruller-sample-feature-flag/blob/master/rules/domains.json)
  * [screens.json](https://github.com/flaviostutz/ruller-sample-feature-flag/blob/master/rules/screens.json)

* You can also control some rule computation behaviors with the help of `_config` key on your `.json` file. E.g.
  ```json
  {
    "_config": {
      "seed": 123,
      "default_condition": true,
      "lazy_evaluation": true,
      "flatten": true
    },
    "_groups": [...],
    "_items": [...]
  }
  ```
  * `seed`: controls the pseudo-random generator seed in use. Everytime you change it, new random numbers will be assigned from you input;
  * `default_condition`: tells which value should be returned when a rule have no condition assigned;
  * `lazy_evaluation`: if set to `true`, for any missing input the resulting ruller return error on runtime and does not evaluate the rules, even those that have nothing to do with the missing input. The default value of this param is `false`, meaning the resulting ruller will evaluate all rules even though an input is missing;
  * `flatten`: the output map will get flattened.

## Development tips

* Always explicitly define the `target` and `templdir` runtime parameters on development time. E.g.:

```sh
go build
./ruller-dsl-feature-flag --source "example-rules/*.json" --target ./rules.go --templdir ./templates
```

* After closing a version
  * Tag the repository code
  * Check the autobuild on Dockerhub to see if all went well and that the new tag was created
  * Update and test the example project https://github.com/flaviostutz/ruller-sample-feature-flag


