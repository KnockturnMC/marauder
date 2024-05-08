# marauder

The marauder project aims to fully cover knockturn's server instances from start, stopping and upgrading instances.
This also includes deployment of plugins to the servers.
In this way, marauder is a form of operator.

The project is split into three main parts, as defined below.

## Controller

The marauder controller is the source of truth for marauder.
It holds onto the current state of the network and its deployments as well as the target state of the network.
The controller manages and stores [artefacts](#marauder-artefact) as well as the configuration of each server.
A server configuration defines the used docker image, allocated cpus and memory as well as joined docker networks or host-exposed ports.

The controller is also aware of each [operator](#operator) to actually execute requests on physical machines.
As such, the controller can be understood as the control plane of the network.

## Operator

The marauder operator is responsible for applying the state defined by the controller.
The operator hence is responsible for applying
new [artefact](#marauder-artefact) versions defined by the controller if asked, as well as managing the containers in which the actual servers run.
The operators general loop starts with a restart of a running server. 
1. The operator shuts down and deletes the docker container of the server, leaving the server data intact as it is bind-mounted in.
2. The operator queries the [controller](#controller) for updates its needs to install. The controller knows both the ***is*** and ***target*** state
   of the server, so the controller can supply the exact difference to the operator.
3. The operator requests the flattened files of the current artifacts that need updating from the controller to delete the exact files provided by
   the currently installed artifact.
4. The operator downloads the new artifacts and installs them into the server.
5. The operator notifies the controller about the update, through which the controller can update its ***is*** state.
6. The operator starts the server again.

## Client

The marauder client is a simple cli tool that is capable of building [marauder artefacts](#marauder-artefact) and uploading them
to the [controller](#controller).
For usage instructions, its `--help` page can be queried.

# Marauder artefact

A marauder artifact defines a specific version deployment of a plugin.
It is a .tar.gz archive holding onto a manifest file in json format and the files that are part of the deployment.
An example of this layout for the "MyPlugin" plugin might look like this:

```
myplugin-7.0.8-artefact.tar.gz
├── files
│   └── plugins
│       └── MyPlugin.jar
└── manifest.json
```

where the `manifest.json` file might look like this

```json
{
	"identifier": "myplugin",
	"version": "7.0.8",
	"files": [
		{
			"target": "plugins/",
			"ciSourceGlob": "downloads/MyPlugin.jar",
			"restrictions": {
				"exact": 1
			},
			"matchedFiles": {
				"files/plugins/WorldGuard.jar": "26c51844c3d9ad678b9935d8f84fca018b0620b8fb4ca494d1d8dbf031d02c8d"
			}
		}
	],
	"buildInformation": {
		"repository": "git@git.github.com:KnockturnMC/MyPlugin.git",
		"branch": "master",
		"commitUser": "Bjarne Koll",
		"commitEmail": "lynxplay101@gmail.com",
		"commitHash": "df1f6176e8ac2e88e710a860e47a88edff3b7810",
		"commitMessage": "Release 7.0.8y",
		"timestamp": "2024-05-08T02:21:56.672296787+02:00",
		"buildSpecificVersion": "gdf1f617"
	},
	"deploymentTargets": {
		"integration": [
			"integration-server"
		],
		"production": [
			"skyblock",
			"hub",
			"builder",
			"events"
		]
	}
}
```
