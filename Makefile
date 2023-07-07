BACKUP ?= undefined
DEPLOYMENT ?= default
HELM ?= helm
PLAYER ?= undefined
REPOSITORY ?= aufish/discordant
REVISION ?= undefined

SET_REPOSITORY ?=
ifneq ($(REPOSITORY), undefined)
SET_REPOSITORY = --set repository=$(REPOSITORY)
endif

SET_REVISION ?=
ifneq ($(REVISION), undefined)
SET_REVISION = --set revision=$(REVISION)
endif

docker:
	REPOSITORY=$(REPOSITORY) REVISION=$(REVISION) docker compose build

publish:
	REPOSITORY=$(REPOSITORY) REVISION=$(REVISION) docker compose push

install:
	$(HELM) install $(DEPLOYMENT) -f "`pwd`/_ops/helm/discordant/$(DEPLOYMENT).yaml" _ops/helm/discordant $(SET_REPOSITORY) $(SET_REVISION)

upgrade:
	$(HELM) upgrade $(DEPLOYMENT) -f "`pwd`/_ops/helm/discordant/$(DEPLOYMENT).yaml" _ops/helm/discordant $(SET_REPOSITORY) $(SET_REVISION)

template:
	$(HELM) template $(DEPLOYMENT) -f "`pwd`/_ops/helm/discordant/$(DEPLOYMENT).yaml" _ops/helm/discordant --debug $(SET_REPOSITORY) $(SET_REVISION)

uninstall:
	$(HELM) uninstall $(DEPLOYMENT)
