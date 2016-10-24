.PHONY: clean install

BOTTICELLI = botticelli-linux-amd64
DEPLOY_HOST = # To be set from the command line

$(BOTTICELLI): main.go
	GOARCH=amd64 GOOS=linux go build -v -o $(BOTTICELLI)

clean:
	rm -rf -- $(BOTTICELLI)

deploy: $(BOTTICELLI) botticelli.service deploy.sh rc.local
	if test -z "$(DEPLOY_HOST)"; then                                      \
	  echo "usage: make deploy DEPLOY_HOST=1.2.3.4" 1>&2;                  \
	  exit 1;                                                              \
	fi
	scp $(BOTTICELLI) botticelli.service deploy.sh rc.local $(DEPLOY_HOST):
