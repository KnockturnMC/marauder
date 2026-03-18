plugins {
	java
	`maven-publish`
	`java-library`
	id("com.google.protobuf") version "0.9.6"
}

java.toolchain.languageVersion = JavaLanguageVersion.of(21)
tasks.compileJava.configure { options.release = 21 }

repositories {
	mavenCentral()
}

dependencies {
	api("com.google.protobuf:protobuf-java:4.34.0")
}

publishing {
	repositories {
		maven("https://repo.knockturnmc.com/content/repositories/knockturn-public-snapshot/") {
			name = "knockturnPublic"
			credentials(PasswordCredentials::class)
		}
	}

	publications.create<MavenPublication>("maven") {
		artifactId = project.name.lowercase()
		from(components["java"])
	}
}
