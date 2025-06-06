package pkg

// ManifestFileName defines the name of the manifest file inside the tarball of an artefact.
const ManifestFileName = "manifest.json"

// TLSCertificateFileName represents the certificate file name.
const TLSCertificateFileName = "tls.crt"

// TLSKeyFileName represents the key file name.
const TLSKeyFileName = "tls.key"

// TLSPoolDir defines the dir in which all pool certs are found.
const TLSPoolDir = "pool"

// The FileParentDirectoryInArtefact holds the prefix under which all files are stored in the artefact tarball.
const FileParentDirectoryInArtefact = "files/"

// The MarauderEnvironmentBranchOverride constant defines the environment name that may be defined when calling the
// marauder client to overwrite the git branch detection.
const MarauderEnvironmentBranchOverride = "MARAUDER_BRANCH_OVERRIDE"

// The MarauderEnvironmentBuildSpecificVersionOverride constant defines the environment name that may be defined when calling the
// marauder client to overwrite the build-specific version.
const MarauderEnvironmentBuildSpecificVersionOverride = "MARAUDER_BUILD_SPECIFIC_VERSION_OVERRIDE"
