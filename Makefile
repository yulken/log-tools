BIN_DIR := ./bin
INSTALL_DIR := $(HOME)/go/bin

# Descobre automaticamente todas as pastas que contêm um arquivo .go com o mesmo nome
TOOLS := $(notdir $(patsubst %/,%,$(dir $(wildcard */*.go))))

.PHONY: all clean $(TOOLS)

all:
	@echo "Ferramentas disponíveis: $(TOOLS)"
	@echo "Uso: make <nome-da-ferramenta>"

$(TOOLS):
	@if [ -d "$@" ]; then \
		echo "==> Building Go: $@" ; \
		mkdir -p $(BIN_DIR) ; \
		go build -o $(BIN_DIR)/$@ ./$@/main.go ; \
		mkdir -p $(INSTALL_DIR) ; \
		cp $(BIN_DIR)/$@ $(INSTALL_DIR)/ ; \
		echo "==> Available in $(INSTALL_DIR)/$@" ; \
	fi

clean:
	rm -rf $(BIN_DIR)