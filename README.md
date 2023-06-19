# marauder

The marauder project aims to fully cover knockturn's server instances from start, stopping and upgrading instances.
This also includes deployment of plugins to the servers.
In this way, marauder is a form of operator.

The project is split into three main component, as defined below.

## Controller

The marauder controller is the source of truth for marauder. It holds onto the current state of the network and its deployments as well as the
target state of the network.
The controller manages and stores [artefacts](#marauder-artefact) as well as the configuration of each server.

## Operator

The marauder operator is responsible for applying the state defined by the controller. The operator hence is responsible for applying
new [artefact](#marauder-artefact) version defined by the controller if asked, as well as managing the containers in which the actual servers run.
The operators general loop starts with a restart of a running server. 
1. The operator shuts down and deletes the docker container of the server, leaving the server data intact as it is bind-mounted in.
2. The operator queries the [controller](#controller) for updates its needs to install. The controller knows both the ***is*** and ***target*** state
   of the server, so the controller can supply the exact difference to the operator.
3. The operator requests the flattened files of the current artefacts that need updating from the controller to delete the exact files provided by
   the currently installed artefact.
4. The operator downloads the new artefacts and installs them into the server.
5. The operator notifies the controller about the update, through which the controller can update its ***is*** state.
6. The operator starts the server again.

## Builder

The marauder builder is a simple cli tool that is capable of building [marauder artefacts](#marauder-artefact) and uploading them
to the [controller](#controller).

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
