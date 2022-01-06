docs_out_dir := ./.public

docker_opts ?= --rm --tty --user "$$(id -u)"

antora_build_version ?= 2.3.3
antora_cmd ?= $(DOCKER_CMD) run $(docker_opts) --volume "$${PWD}":/antora docker.io/vshn/antora:$(antora_build_version)
antora_opts ?= --cache-dir=.cache/antora

.PHONY: docs
docs: docs-html ## All-in-one docs build

.PHONY: docs-html
docs-html: $(docs_out_dir)/index.html ## Generate HTML version of documentation with Antora, output at ./.public
	@touch $(docs_out_dir)/.nojekyll

$(docs_out_dir)/index.html:
	$(antora_cmd) $(antora_opts) docs/antora-playbook.yml

.PHONY: docs-publish
docs-publish: node_modules docs-html ## Publishes the Antora documentation on Github Pages
	npm run deploy

node_modules:
	npm install
