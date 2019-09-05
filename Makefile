COVERDIR=$(CURDIR)/.cover
COVERAGEFILE=$(COVERDIR)/cover.out
COVERAGEREPORT=$(COVERDIR)/report.html

GINKGO=go run github.com/onsi/ginkgo/ginkgo

test:
	@$(GINKGO) --failFast ./...

test-watch:
	@$(GINKGO) watch -cover -r ./...

coverage-ci:
	@mkdir -p $(COVERDIR)
	@$(GINKGO) -r -covermode=atomic --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --race --compilers=2 ./
	@echo "mode: count" > "${COVERAGEFILE}"
	@find . -type f -name '*.coverprofile' -exec cat {} \; -exec rm -f {} \; | grep -h -v "^mode:" >> ${COVERAGEFILE}

coverage: coverage-ci
	@sed -i -e "s|_$(PROJECT_ROOT)/|./|g" "${COVERAGEFILE}"
	@cp "${COVERAGEFILE}" coverage.txt

coverage-html:
	@go tool cover -html="${COVERAGEFILE}" -o $(COVERAGEREPORT)
	@xdg-open $(COVERAGEREPORT) 2> /dev/null > /dev/null

vet:
	@go vet ./...

fmt:
	@go fmt ./...

.PHONY: test test-watch coverage coverage-ci coverage-html vet fmt
