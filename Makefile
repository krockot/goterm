include $(GOROOT)/src/Make.inc

.PHONY: all term install examples clean

all: install examples

pam:
	gomake -C term

install: pam
	gomake -C term install

examples:
	gomake -C examples

clean:
	gomake -C term clean
	gomake -C examples clean


