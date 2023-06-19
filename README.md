# marauder

The marauder project aims to fully cover knockturn's server instances from start, stopping and upgrading instances.
This also includes deployment of plugins to the servers.
In this way, marauder is a form of operator.

The project is split into three main component, as defined below.

## Controller

The marauder controller is the source of truth for marauder. It holds onto the current state of the network and its deployments as well as the
target state of the network.

## Operator

The marauder operator is responsible for applying the state defined by the controller. The operator hence is responsible for applying new deployments
defined by the controller if asked to as well as managing the containers in which the actual servers run.

## Builder

The marauder builder is a simple cli tool that is capable of building [marauder artefacts]()

# Marauder artefact

A marauder artefact defines a specific version deployment of a plugin. It is a .tar.gz archive holding onto a manifest file in json format and
the files that are part of the deployment. An example of this layout for, e.g., the spellcore plugin would be:

```
spellcore-1.14.0+e0eef91.tar.gz
├── files
│   └── plugins
│       ├── KnockturnCore
│       │   └── modules
│       │       └── spellcore-api-1.14.0+e0eef914.jar
│       └── spellcore-plugin-1.14.0+e0eef914.jar
└── manifest.json

```

where the `manifest.json` file might look like this
```json
{
  "identifier": "spellcore",
  "version": "1.14.0+e0eef91",
    "files": [
        {
            "target": "plugins/spellcore-plugin-{{.Version}}.jar",
            "ciSourceGlob": "./spell-plugin/build/libs/spell-plugin-*-final.jar"
        },
        {
            "target": "plugins/KnockturnCore/modules/spellcore-api-{{.Version}}.jar",
            "ciSourceGlob": "./spell-api/build/libs/spell-api-*-final.jar "
        }
    ]
}
```
