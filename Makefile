# A collection of all phony goals that are not cacheable.
.PHONY: format lint unittest functiontest regenMocks

format:
	@scripts/format.sh
lint:
	@scripts/lint.sh
unittest:
	@scripts/unittest.sh
functiontest:
	@scripts/functiontest.sh
regenMocks:
	@scripts/regenerateMocks.sh
