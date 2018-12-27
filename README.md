# ruller-dsl-feature-flag-tree
A feature flag engine that can be used to enable, change or rollout features of a system dynamically based on system or user attributes

A system can check for enabled features by performing a REST call to (ex.: /menu) having a JSON body with some input attributes (user info, environment info etc). Some conditions will be evaluated and a JSON containing all enabled features will be returned. Then the client system can use this information to decide on what to enable/disable/configure from its internals.

This DSL tool will get a JSON written with some feature tree rules and generate a Go code that can be run as a REST service. We use [Ruller](http://github.com/flaviostutz/ruller) framework on our code generation and it will be responsible for the runtime execution of those rules.

While developing, enter '/sample' dir and perform ```docker-compose build``` so that you can run your code against sample rules json files and check for results.

## Feature selection language

* The language is a JSON file with some semantics regarding to feature attributes and condition attributes organized in a tree, so that attributes and conditions from a parent are inherited by its children

* Features are identified by an id and may have some custom attributes bound to it. All features whose "condition" attribute evaluates to true will be returned as the result of the REST call

* Example of a dynamic menu:
```
{
    "label": "not specified",
    "_config": {
        "seed ": 123,
        "default_condition": true
    },
    "_items": [{
            "label": "menu1",
            "_items": [{
                    "label": "menu1 .1",
                    "_condition": "until('2018-12-31 23:32:21')"
                },
                {
                    "label": "menu1.2",
                    "_condition ": "input: idade > 30 and random(30, input: customerid)",
                    "_items": [{
                            "label": "menu1 .2 .1 "
                        },
                        {
                            "label": "menu1 .2 .2",
                            "_condition": "random(20)"
                        }
                    ]
                }
            ]
        },
        {
            "label": "menu2",
            "_condition": "input:_city=='Brasília'",
            "_items": [{
                    "label": "menu2.1",
                    "_condition": "from('2018-11-31 23:32:21') or input:state~='DF|RJ'"
                },
                {
                    "label": "menu2.2"
                }
            ]
        }
    ]
}
```

* Example of infrastructure selection by domain name from client:
```
{
    "_config": {
        "seed": 123,
        "default_condition": true
    },
    "_items": [{
            "provider": "aws",
            "_condition": "random(10, input:customerid)",
            "_items": [{
                    "region": "sa-east-1",
                    "login": "login.sa-east-1.test.com",
                    "bootcamp": "bootcamp.sa-east-1.test.com",
                    "_condition": "input:_country=='Brazil'"
                },
                {
                    "region": "us-west-1",
                    "login": "login.us-west-1.test.com",
                    "events": "events.us-west-1.test.com",
                    "bootcamp": "bootcamp.us-west-1.test.com"
                }
            ]
        },
        {
            "provider": "azure",
            "_condition": "random(10, 50, input:customerid)",
            "_items": [{
                    "region": "azure-brazil",
                    "login": "login.azure-brazil.test.com",
                    "_condition": "input:_country=='Brazil'"
                },
                {
                    "region": "azure-ny",
                    "login": "login.azure-ny.test.com",
                    "events": "events.azure-ny.test.com",
                    "bootcamp": "bootcamp.azure-ny.test.com"
                }
            ]
        },
        {
            "provider": "vpsdime",
            "_items": [{
                "region": "vpsdime",
                "login": "login.vpsdime.test.com",
                "events": "events.azure-ny.test.com",
                "bootcamp": "bootcamp.azure-ny.test.com"
            }]
        }
    ]
}
```
