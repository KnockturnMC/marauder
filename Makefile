# A collection of all phony goals that are not cacheable.
.PHONY: format lint unittest functiontest generateCode

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
