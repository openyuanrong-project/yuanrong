# Minimal makefile for Sphinx documentation
#

# You can set these variables from the command line, and also
# from the environment for the first two.
# should use  -W --keep-going -n
SPHINXOPTS    ?=
SPHINXBUILD   ?= sphinx-build
SOURCEDIR     = .
RUNTIMEDIR    = ../
BUILDDIR      = _build
GIT_COMMIT_ID := $(shell git rev-parse HEAD)

# Put it first so that "make" without argument is like "make help".
help:
	@YR_DOC_GIT_COMMIT_ID=$(GIT_COMMIT_ID) $(SPHINXBUILD) -M help "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)

.PHONY: help Makefile clean

# Catch-all target: route all unknown targets to Sphinx using the new
# "make mode" option.  $(O) is meant as a shortcut for $(SPHINXOPTS).
%: Makefile
	@YR_DOC_GIT_COMMIT_ID=$(GIT_COMMIT_ID) $(SPHINXBUILD) -M $@ "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)

html: apidoc
	@YR_DOC_GIT_COMMIT_ID=$(GIT_COMMIT_ID) $(SPHINXBUILD) -M html "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)

livehtml:
	sphinx-autobuild "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) --host "0.0.0.0" --port 9999 $(O)

apidoc:
	cd "$(RUNTIMEDIR)" && doxygen "Doxyfile"

clean:
	rm -rf ${SOURCEDIR}/docs/multi_language_function_programming_interface/distributed_programming_interface/api/Python/generated
	rm -rf ${SOURCEDIR}/docs/_build/
