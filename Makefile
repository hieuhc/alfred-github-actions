
TARGET_DIR := target
WORKFLOW_FILE := $(TARGET_DIR)/alfred-gha.alfredworkflow


target:
	mkdir target

clean:
	@[ -d $(TARGET_DIR) ] && rm -r $(TARGET_DIR) || true

workflow: build clean target
	zip $(WORKFLOW_FILE) \
	info.plist \
	icon.png \
	icons/* \
	bin/entry \
	bin/fetch_repos \
	bin/fetch_workflows \
	bin/fetch_runs \
	bin/watch_run \

build-entry:
	go build -ldflags='-s -w' -trimpath -o bin/entry gha/entry/entry.go

build-repos:
	go build -ldflags='-s -w' -trimpath -o bin/fetch_repos gha/fetch_repo/fetch_repos.go

build-workflows:
	go build -ldflags='-s -w' -trimpath -o bin/fetch_workflows gha/fetch_workflows/fetch_workflows.go

build-runs:
	go build -ldflags='-s -w' -trimpath -o bin/fetch_runs gha/fetch_runs/fetch_runs.go

build-watch-run:
	go build -ldflags='-s -w' -trimpath -o bin/watch_run gha/watch_run/watch_run.go

build: build-entry build-repos build-workflows build-runs build-watch-run
