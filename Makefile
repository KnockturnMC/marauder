# A collection of all phony goals that are not cacheable.
.PHONY: all format lint unittest functiontest generateCode buildDocker buildDockerAndPush

all: format lint unittest functiontest generateCode

format:
	@scripts/format.sh
lint:
	@scripts/lint.sh
unittest:
	@scripts/unittest.sh
functiontest:
	@scripts/functiontest.sh
generateCode:
	@scripts/generateCode.sh
buildDocker:
	@scripts/buildDocker.sh
buildDockerAndPush:
	@scripts/buildDocker.sh --push
