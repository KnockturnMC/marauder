# A collection of all phony goals that are not cacheable.
.PHONY: format lint

format:
	@scripts/format.sh
lint:
	@scripts/lint.sh
