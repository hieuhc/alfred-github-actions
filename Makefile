
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
	bin/main \

build: 
	go build -ldflags='-s -w' -trimpath -o bin/main ./gha
