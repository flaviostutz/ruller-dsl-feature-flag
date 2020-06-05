# ruller-dsl-feature-flag
A feature flag engine that can be used to enable, change or rollout features of a system dynamically based on system or user attributes

A system can check for enabled features by performing a REST call to (ex.: /menu) having a JSON body with some input attributes (user info, environment info etc). Some conditions will be evaluated and a JSON containing all enabled features will be returned. Then the client system can use this information to decide on what to enable/disable/configure from its internals.

This DSL tool will get a JSON written with some feature tree rules and generate a Go code that can be run as a REST service. We use [Ruller](http://github.com/flaviostutz/ruller) framework on our code generation and it will be responsible for the runtime execution of those rules.

While developing, enter '/sample' dir and perform ```docker-compose build``` so that you can run your code against sample rules json files and check for results.

If you want to create your own feature flag service, fork https://github.com/flaviostutz/ruller-sample-feature-flag and create your own rules.

## Usage

```
cd sample-json
docker-compose up -d --build
curl -X POST \
  http://localhost:3000/rules/menu \
  -H 'Content-Type: application/json' \
  -H 'X-Forwarded-For: 177.79.35.49' \
  -H 'cache-control: no-cache' \
  -d '{
	"age": 44,
	"customerid": "22111",
	"state": "DF",
	"app_version": "2.2"
}
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

