# A collection of all phony goals that are not cacheable.
.PHONY: format lint unittest regenMocks

format:
	@scripts/format.sh
lint:
	@scripts/lint.sh
unittest:
	@scripts/unittest.sh
regenMocks:
	@scripts/regenerateMocks.sh
